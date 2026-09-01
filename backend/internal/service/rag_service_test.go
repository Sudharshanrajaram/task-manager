package service_test

import (
	"context"
	"testing"

	"github.com/taskflow/backend/internal/db"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
	"github.com/taskflow/backend/pkg/groq"
)

type mockGroqClient struct{}

func (m *mockGroqClient) CreateEmbedding(ctx context.Context, input string) ([]float32, error) {
	return groq.GenerateDeterministicEmbedding(input, groq.EmbeddingDimension), nil
}

func (m *mockGroqClient) CreateChatCompletion(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return `{"subtasks": ["Step 1: Design architecture", "Step 2: Implement logic", "Step 3: Write tests", "Step 4: Deploy"]}`, nil
}

func setupRAGTestEnvironment(t *testing.T) (service.RAGService, service.ProjectService, repository.EmbeddingRepository) {
	testDB, err := db.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to init test DB: %v", err)
	}

	projectRepo := repository.NewProjectRepository(testDB)
	embeddingRepo := repository.NewEmbeddingRepository(testDB)
	groqClient := &mockGroqClient{}

	projSvc := service.NewProjectService(projectRepo)
	ragSvc := service.NewRAGService(groqClient, embeddingRepo)

	return ragSvc, projSvc, embeddingRepo
}

func TestRAG_SuggestAndFeedbackLoop(t *testing.T) {
	ragSvc, projSvc, _ := setupRAGTestEnvironment(t)
	ctx := context.Background()

	proj, err := projSvc.CreateProject("RAG Test Proj", "RAG", "#6366F1")
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// 1. First suggestion on empty vector store (Cold Start)
	res1, err := ragSvc.SuggestSubtasks(ctx, &proj.ID, "Implement Google OAuth Login", 4)
	if err != nil {
		t.Fatalf("SuggestSubtasks failed: %v", err)
	}
	if len(res1.SuggestedSubtasks) == 0 {
		t.Errorf("Expected non-empty suggested subtasks")
	}
	if len(res1.GroundedContext) != 0 {
		t.Errorf("Expected 0 grounded context on cold start, got %d", len(res1.GroundedContext))
	}

	// 2. User edits and accepts subtasks (Simulating User Review & Ingestion)
	acceptedSubtasks := []string{
		"Create OAuth credentials in Google Cloud Console",
		"Implement redirect URL and state parameter validation",
		"Exchange authorization code for access and ID tokens",
		"Issue TaskFlow session JWT token",
	}
	err = ragSvc.SaveAcceptedSubtasks(ctx, &proj.ID, "Implement Google OAuth Login", acceptedSubtasks)
	if err != nil {
		t.Fatalf("SaveAcceptedSubtasks failed: %v", err)
	}

	// 3. Next suggestion on a related task (Warm Start / Grounded Retrieval)
	res2, err := ragSvc.SuggestSubtasks(ctx, &proj.ID, "Implement GitHub OAuth Login", 4)
	if err != nil {
		t.Fatalf("SuggestSubtasks (warm) failed: %v", err)
	}

	if len(res2.SuggestedSubtasks) == 0 {
		t.Errorf("Expected suggested subtasks on warm query")
	}

	// Verify that the Google OAuth task was retrieved as Grounded Context!
	if len(res2.GroundedContext) == 0 {
		t.Fatalf("Expected at least 1 grounded historical task in context, got 0")
	}

	grounded := res2.GroundedContext[0]
	if grounded.SourceTitle != "Implement Google OAuth Login" {
		t.Errorf("Expected grounded source title 'Implement Google OAuth Login', got '%s'", grounded.SourceTitle)
	}
	if len(grounded.FinalSubtasks) != 4 {
		t.Errorf("Expected 4 historical subtasks in grounded context, got %d", len(grounded.FinalSubtasks))
	}
	if grounded.SimilarityScore <= 0 {
		t.Errorf("Expected positive similarity score, got %f", grounded.SimilarityScore)
	}
}
