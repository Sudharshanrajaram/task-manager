package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/pkg/groq"
)

type GroundedExample struct {
	SourceTitle     string   `json:"source_title"`
	FinalSubtasks   []string `json:"final_subtasks"`
	SimilarityScore float32  `json:"similarity_score"`
}

type SubtaskSuggestionResult struct {
	SuggestedSubtasks []string          `json:"suggested_subtasks"`
	GroundedContext   []GroundedExample `json:"grounded_context"`
	Title             string            `json:"title"`
	Count             int               `json:"count"`
}

type RAGService interface {
	SuggestSubtasks(ctx context.Context, projectID *uuid.UUID, title string, count int) (*SubtaskSuggestionResult, error)
	SaveAcceptedSubtasks(ctx context.Context, projectID *uuid.UUID, title string, subtasks []string) error
}

type ragService struct {
	groqClient    groq.Client
	embeddingRepo repository.EmbeddingRepository
}

func NewRAGService(groqClient groq.Client, embeddingRepo repository.EmbeddingRepository) RAGService {
	return &ragService{
		groqClient:    groqClient,
		embeddingRepo: embeddingRepo,
	}
}

func (s *ragService) SuggestSubtasks(ctx context.Context, projectID *uuid.UUID, title string, count int) (*SubtaskSuggestionResult, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrTaskTitleRequired
	}

	if count <= 0 {
		count = 4
	} else if count > 12 {
		count = 12
	}

	// 1. Generate embedding for query title
	queryVector, err := s.groqClient.CreateEmbedding(ctx, title)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for title: %w", err)
	}

	// 2. Retrieve top-K similar past tasks (RAG retrieval)
	similar, err := s.embeddingRepo.FindSimilar(projectID, queryVector, 3)
	if err != nil {
		slog.Warn("Failed to retrieve similar task embeddings", slog.String("error", err.Error()))
	}

	groundedExamples := make([]GroundedExample, 0, len(similar))
	var contextBuilder strings.Builder

	if len(similar) > 0 {
		contextBuilder.WriteString("Here is historical context of past task breakdowns from this engineer:\n")
		for idx, sim := range similar {
			// Only include if positive similarity
			if sim.Score > 0.1 && len(sim.Embedding.FinalSubtasks) > 0 {
				groundedExamples = append(groundedExamples, GroundedExample{
					SourceTitle:     sim.Embedding.SourceTitle,
					FinalSubtasks:   sim.Embedding.FinalSubtasks,
					SimilarityScore: sim.Score,
				})

				contextBuilder.WriteString(fmt.Sprintf(
					"Example %d: Title: \"%s\"\nAccepted subtasks:\n",
					idx+1, sim.Embedding.SourceTitle,
				))
				for _, sub := range sim.Embedding.FinalSubtasks {
					contextBuilder.WriteString(fmt.Sprintf("  - %s\n", sub))
				}
				contextBuilder.WriteString("\n")
			}
		}
	}

	// 3. Construct LLM prompt
	systemPrompt := `You are an expert software engineer and technical task planner.
Your job is to break down a technical task or ticket title into concise, concrete, actionable engineering subtasks.
Keep each subtask title clear, imperative, and focused (e.g. "Implement OAuth callback handler", "Add database index for user_id").
Respond ONLY in valid JSON matching this schema:
{
  "subtasks": [
    "Subtask 1 description",
    "Subtask 2 description"
  ]
}`

	var userPrompt strings.Builder
	if contextBuilder.Len() > 0 {
		userPrompt.WriteString(contextBuilder.String())
	}
	userPrompt.WriteString(fmt.Sprintf(
		"Now, please break down this NEW task into exactly %d actionable subtasks:\nTask Title: \"%s\"",
		count, title,
	))

	// 4. Call Groq Chat Completion
	respJSON, err := s.groqClient.CreateChatCompletion(ctx, systemPrompt, userPrompt.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate subtasks: %w", err)
	}

	// 5. Parse JSON output
	type subtasksResponse struct {
		Subtasks []string `json:"subtasks"`
	}
	var parsed subtasksResponse

	// Clean any markdown code fences if LLM wrapped it in ```json ... ```
	cleanJSON := strings.TrimSpace(respJSON)
	if strings.HasPrefix(cleanJSON, "```json") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
		cleanJSON = strings.TrimSuffix(cleanJSON, "```")
	} else if strings.HasPrefix(cleanJSON, "```") {
		cleanJSON = strings.TrimPrefix(cleanJSON, "```")
		cleanJSON = strings.TrimSuffix(cleanJSON, "```")
	}
	cleanJSON = strings.TrimSpace(cleanJSON)

	if err := json.Unmarshal([]byte(cleanJSON), &parsed); err != nil {
		slog.Warn("Failed to unmarshal strict LLM JSON, parsing lines fallback",
			slog.String("raw", respJSON),
			slog.String("error", err.Error()),
		)
		// Fallback: extract non-empty lines
		lines := strings.Split(cleanJSON, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(strings.TrimLeft(l, "-*0123456789. \""))
			l = strings.TrimRight(l, "\",")
			if l != "" && !strings.HasPrefix(l, "{") && !strings.HasPrefix(l, "}") && !strings.Contains(l, "subtasks") {
				parsed.Subtasks = append(parsed.Subtasks, l)
			}
		}
	}

	// Clamp to requested count
	if len(parsed.Subtasks) > count {
		parsed.Subtasks = parsed.Subtasks[:count]
	}

	return &SubtaskSuggestionResult{
		SuggestedSubtasks: parsed.Subtasks,
		GroundedContext:   groundedExamples,
		Title:             title,
		Count:             len(parsed.Subtasks),
	}, nil
}

func (s *ragService) SaveAcceptedSubtasks(ctx context.Context, projectID *uuid.UUID, title string, subtasks []string) error {
	title = strings.TrimSpace(title)
	if title == "" || len(subtasks) == 0 {
		return nil
	}

	// Clean subtasks
	cleanedSubs := make([]string, 0, len(subtasks))
	for _, sub := range subtasks {
		sub = strings.TrimSpace(sub)
		if sub != "" {
			cleanedSubs = append(cleanedSubs, sub)
		}
	}
	if len(cleanedSubs) == 0 {
		return nil
	}

	// Generate embedding for this title
	vector, err := s.groqClient.CreateEmbedding(ctx, title)
	if err != nil {
		slog.Warn("Failed to generate embedding for accepted task", slog.String("error", err.Error()))
		return err
	}

	entry := &model.TaskTitleEmbedding{
		ProjectID:     projectID,
		SourceTitle:   title,
		FinalSubtasks: cleanedSubs,
	}

	if err := s.embeddingRepo.SaveEmbedding(entry, vector); err != nil {
		slog.Error("Failed to save task title embedding to vector store", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Persisted task embedding to RAG feedback store",
		slog.String("title", title),
		slog.Int("subtasks_count", len(cleanedSubs)),
	)

	return nil
}
