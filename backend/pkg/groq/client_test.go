package groq_test

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/taskflow/backend/pkg/groq"
)

func TestCosineSimilarity(t *testing.T) {
	// 1. Identical vectors -> Cosine = 1.0
	vecA := []float32{1.0, 2.0, 3.0}
	vecB := []float32{1.0, 2.0, 3.0}
	sim := groq.CosineSimilarity(vecA, vecB)
	if math.Abs(float64(sim-1.0)) > 0.0001 {
		t.Errorf("Expected similarity 1.0 for identical vectors, got %f", sim)
	}

	// 2. Orthogonal vectors -> Cosine = 0.0
	vecC := []float32{1.0, 0.0, 0.0}
	vecD := []float32{0.0, 1.0, 0.0}
	simOrth := groq.CosineSimilarity(vecC, vecD)
	if math.Abs(float64(simOrth)) > 0.0001 {
		t.Errorf("Expected similarity 0.0 for orthogonal vectors, got %f", simOrth)
	}

	// 3. Empty vectors -> 0.0
	if groq.CosineSimilarity([]float32{}, []float32{}) != 0 {
		t.Errorf("Expected 0 for empty vectors")
	}
}

func TestGenerateDeterministicEmbedding(t *testing.T) {
	// Same text should produce identical vectors
	vec1 := groq.GenerateDeterministicEmbedding("Configure PostgreSQL Database Connection", groq.EmbeddingDimension)
	vec2 := groq.GenerateDeterministicEmbedding("Configure PostgreSQL Database Connection", groq.EmbeddingDimension)

	if len(vec1) != groq.EmbeddingDimension {
		t.Errorf("Expected vector dimension %d, got %d", groq.EmbeddingDimension, len(vec1))
	}

	simSame := groq.CosineSimilarity(vec1, vec2)
	if math.Abs(float64(simSame-1.0)) > 0.0001 {
		t.Errorf("Expected identical text to have similarity 1.0, got %f", simSame)
	}

	// Related text should have higher similarity than unrelated text
	vecRelated := groq.GenerateDeterministicEmbedding("Setup Postgres DB Migrations", groq.EmbeddingDimension)
	vecUnrelated := groq.GenerateDeterministicEmbedding("Fix CSS Navbar Padding on Mobile Safari", groq.EmbeddingDimension)

	simRelated := groq.CosineSimilarity(vec1, vecRelated)
	simUnrelated := groq.CosineSimilarity(vec1, vecUnrelated)

	if simRelated <= simUnrelated {
		t.Errorf("Expected related text similarity (%f) > unrelated text similarity (%f)", simRelated, simUnrelated)
	}
}

func TestLoudFailureWithoutAPIKey(t *testing.T) {
	client := groq.NewClient("", "", "")
	ctx := context.Background()

	// Embedding must fail loudly
	_, err := client.CreateEmbedding(ctx, "Test Task Title")
	if err == nil {
		t.Fatal("Expected CreateEmbedding to fail loudly when GROQ_API_KEY is not set")
	}
	if !strings.Contains(err.Error(), "GROQ_API_KEY is not configured") {
		t.Errorf("Expected loud GROQ_API_KEY error message, got: %v", err)
	}

	// Chat completion must fail loudly
	_, err = client.CreateChatCompletion(ctx, "system prompt", "user prompt")
	if err == nil {
		t.Fatal("Expected CreateChatCompletion to fail loudly when GROQ_API_KEY is not set")
	}
	if !strings.Contains(err.Error(), "GROQ_API_KEY is not configured") {
		t.Errorf("Expected loud GROQ_API_KEY error message, got: %v", err)
	}
}
