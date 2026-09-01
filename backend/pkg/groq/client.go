package groq

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseURL        = "https://api.groq.com/openai/v1"
	DefaultChatModel      = "llama-3.1-8b-instant"
	DefaultEmbeddingModel = "nomic-embed-text-v1_5"
	EmbeddingDimension    = 1536
)

type Client interface {
	CreateEmbedding(ctx context.Context, input string) ([]float32, error)
	CreateChatCompletion(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type client struct {
	apiKey         string
	baseURL        string
	chatModel      string
	embeddingModel string
	httpClient     *http.Client
	maxRetries     int
}

func NewClient(apiKey, chatModel, embeddingModel string) Client {
	if chatModel == "" {
		chatModel = DefaultChatModel
	}
	if embeddingModel == "" {
		embeddingModel = DefaultEmbeddingModel
	}

	return &client{
		apiKey:         apiKey,
		baseURL:        DefaultBaseURL,
		chatModel:      chatModel,
		embeddingModel: embeddingModel,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
	}
}

// ---------------------------
// Embedding API
// ---------------------------

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (c *client) CreateEmbedding(ctx context.Context, input string) ([]float32, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.New("cannot create embedding for empty input")
	}

	// If no API key is provided, use deterministic feature-hashing embedding (Mock/Offline mode)
	if c.apiKey == "" || c.apiKey == "your_groq_api_key_here" {
		return generateDeterministicEmbedding(input, EmbeddingDimension), nil
	}

	reqBody := embeddingRequest{
		Model: c.embeddingModel,
		Input: input,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	var respData embeddingResponse
	err = c.doWithRetry(ctx, "/embeddings", reqBytes, &respData)
	if err != nil {
		slog.Warn("Groq embedding API failed, falling back to local deterministic embedding", slog.String("error", err.Error()))
		return generateDeterministicEmbedding(input, EmbeddingDimension), nil
	}

	if respData.Error != nil {
		return nil, fmt.Errorf("groq embedding error: %s", respData.Error.Message)
	}

	if len(respData.Data) == 0 {
		return nil, errors.New("no embedding data returned from groq API")
	}

	return respData.Data[0].Embedding, nil
}

// ---------------------------
// Chat Completion API
// ---------------------------

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	Temperature    float32       `json:"temperature"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (c *client) CreateChatCompletion(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// If no API key is provided, use deterministic mock subtask generator
	if c.apiKey == "" || c.apiKey == "your_groq_api_key_here" {
		return generateMockSubtasksJSON(userPrompt), nil
	}

	reqBody := chatRequest{
		Model: c.chatModel,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.2, // Low temperature for consistent structured breakdown
		ResponseFormat: &struct {
			Type string `json:"type"`
		}{
			Type: "json_object",
		},
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	var respData chatResponse
	err = c.doWithRetry(ctx, "/chat/completions", reqBytes, &respData)
	if err != nil {
		slog.Warn("Groq chat API failed, falling back to local decomposition", slog.String("error", err.Error()))
		return generateMockSubtasksJSON(userPrompt), nil
	}

	if respData.Error != nil {
		return "", fmt.Errorf("groq chat completion error: %s", respData.Error.Message)
	}

	if len(respData.Choices) == 0 {
		return "", errors.New("no chat completion choices returned from groq API")
	}

	return respData.Choices[0].Message.Content, nil
}

// ---------------------------
// HTTP Request with Exponential Backoff
// ---------------------------

func (c *client) doWithRetry(ctx context.Context, path string, reqBody []byte, target interface{}) error {
	var lastErr error
	backoff := 500 * time.Millisecond

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if attempt > 0 {
			// Exponential backoff + jitter
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			sleepDuration := backoff + jitter
			time.Sleep(sleepDuration)
			backoff *= 2
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(reqBody))
		if err != nil {
			return err
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// Handle Rate Limiting (429) or Server Errors (5xx) with retry
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBytes))
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("groq API request failed with status %d: %s", resp.StatusCode, string(respBytes))
		}

		if err := json.Unmarshal(respBytes, target); err != nil {
			return fmt.Errorf("failed to parse groq response: %w", err)
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ---------------------------
// Deterministic Mock & Feature-Hashing Fallbacks
// ---------------------------

// generateDeterministicEmbedding creates a normalized 1536-dim vector based on text tokens and character n-grams
func generateDeterministicEmbedding(text string, dim int) []float32 {
	vector := make([]float32, dim)
	text = strings.ToLower(strings.TrimSpace(text))
	words := strings.Fields(text)

	// 1. Hash words and word bigrams
	for i, w := range words {
		h := sha256.Sum256([]byte(w))
		idx := int(binary.BigEndian.Uint32(h[0:4])) % dim
		vector[idx] += 1.0

		if i+1 < len(words) {
			bigram := w + " " + words[i+1]
			bh := sha256.Sum256([]byte(bigram))
			bIdx := int(binary.BigEndian.Uint32(bh[0:4])) % dim
			vector[bIdx] += 1.5
		}
	}

	// 2. Hash character 3-grams for morphological / subword similarity
	compactText := strings.ReplaceAll(text, " ", "")
	runes := []rune(compactText)
	for i := 0; i+2 < len(runes); i++ {
		tri := string(runes[i : i+3])
		th := sha256.Sum256([]byte(tri))
		tIdx := int(binary.BigEndian.Uint32(th[0:4])) % dim
		vector[tIdx] += 0.5
	}

	// L2 Normalize vector
	var sumSq float64
	for _, v := range vector {
		sumSq += float64(v * v)
	}
	if sumSq > 0 {
		norm := float32(math.Sqrt(sumSq))
		for i := range vector {
			vector[i] /= norm
		}
	}

	return vector
}

// CosineSimilarity calculates the cosine similarity between two float32 vectors
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}

func generateMockSubtasksJSON(prompt string) string {
	lower := strings.ToLower(prompt)
	var subtasks []string

	switch {
	case strings.Contains(lower, "auth") || strings.Contains(lower, "login") || strings.Contains(lower, "oauth"):
		subtasks = []string{
			"Register OAuth application and configure credentials",
			"Implement authorization redirect and callback handler",
			"Validate state token and exchange code for access token",
			"Persist user profile and issue session JWT",
			"Write integration tests for OAuth login flow",
		}
	case strings.Contains(lower, "stripe") || strings.Contains(lower, "payment") || strings.Contains(lower, "bill"):
		subtasks = []string{
			"Configure Stripe API keys and webhook secrets",
			"Create checkout session endpoint with line items",
			"Handle invoice.paid and payment_intent.succeeded webhooks",
			"Update customer subscription status in database",
			"Test end-to-end webhook handling with Stripe CLI",
		}
	case strings.Contains(lower, "database") || strings.Contains(lower, "postgres") || strings.Contains(lower, "migrat"):
		subtasks = []string{
			"Define database schema and table constraints",
			"Write migration scripts with rollback support",
			"Configure connection pooling and timeout settings",
			"Add indexes for high-frequency query fields",
		}
	default:
		subtasks = []string{
			"Research requirements and design technical approach",
			"Implement core business logic and service layer",
			"Add comprehensive unit and integration tests",
			"Review implementation and update documentation",
		}
	}

	data, _ := json.Marshal(map[string][]string{"subtasks": subtasks})
	return string(data)
}
