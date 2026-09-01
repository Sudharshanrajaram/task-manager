package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

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
	SummarizeTask(ctx context.Context, task *model.Task, subtasks []model.Subtask) (summary string, fromCache bool, err error)
	GenerateStandup(ctx context.Context, tasks []model.Task, totalSeconds int64) (string, error)
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

func (s *ragService) SummarizeTask(ctx context.Context, task *model.Task, subtasks []model.Subtask) (string, bool, error) {
	// 1. Calculate input content hash
	hasher := sha256.New()
	hasher.Write([]byte(task.Title))
	hasher.Write([]byte("\n" + task.Description))
	for _, sub := range subtasks {
		hasher.Write([]byte("\n- " + sub.Title))
	}
	currentHash := hex.EncodeToString(hasher.Sum(nil))

	// 2. Check if cache is valid
	if task.AISummary != nil && task.AISummarySourceHash != nil && *task.AISummarySourceHash == currentHash {
		return *task.AISummary, true, nil
	}

	// 3. Prompt LLM for concise 1-2 sentence plain-English summary
	systemPrompt := "You are an expert software engineering assistant. In 1 to 2 concise, clear sentences (maximum 50 words), summarize what this engineering ticket accomplishes and its core scope. Return plain text only with no markdown formatting."

	var userPromptBuilder strings.Builder
	userPromptBuilder.WriteString(fmt.Sprintf("Title: %s\n", task.Title))
	if task.Description != "" {
		userPromptBuilder.WriteString(fmt.Sprintf("Description: %s\n", task.Description))
	}
	if len(subtasks) > 0 {
		userPromptBuilder.WriteString("Subtasks:\n")
		for _, sub := range subtasks {
			userPromptBuilder.WriteString(fmt.Sprintf("- %s\n", sub.Title))
		}
	}

	summary, err := s.groqClient.CreateChatCompletion(ctx, systemPrompt, userPromptBuilder.String())
	if err != nil {
		return "", false, err
	}

	summary = strings.TrimSpace(summary)
	now := time.Now().UTC()
	task.AISummary = &summary
	task.AISummarySourceHash = &currentHash
	task.AISummaryGeneratedAt = &now

	return summary, false, nil
}

func (s *ragService) GenerateStandup(ctx context.Context, tasks []model.Task, totalSeconds int64) (string, error) {
	systemPrompt := `You are an engineering standup assistant. Generate a clean, professional Daily Standup report formatted in Markdown based on the engineer's recent task updates and time logged.

Use this format:
### 🚀 Completed & Delivered
(bullet list of completed tasks with ticket keys)

### 🔨 In Progress & Next Focus
(bullet list of tasks in progress or in review)

### ⚠️ Blockers & Risks
(bullet list of any blocked tasks with reasons, or "No active blockers")`

	var promptBuilder strings.Builder
	h := totalSeconds / 3600
	m := (totalSeconds % 3600) / 60
	promptBuilder.WriteString(fmt.Sprintf("Total tracked time today: %dh %dm\n\nTickets:\n", h, m))

	for _, t := range tasks {
		status := string(t.Status)
		blockedStr := ""
		if t.IsBlocked {
			blockedStr = " [BLOCKED"
			if t.BlockedReason != nil {
				blockedStr += ": " + *t.BlockedReason
			}
			blockedStr += "]"
		}
		promptBuilder.WriteString(fmt.Sprintf("- [%s] %s (Status: %s)%s\n", t.TicketKey, t.Title, status, blockedStr))
	}

	report, err := s.groqClient.CreateChatCompletion(ctx, systemPrompt, promptBuilder.String())
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(report), nil
}
