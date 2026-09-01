package repository

import (
	"encoding/json"
	"sort"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/pkg/groq"
	"gorm.io/gorm"
)

type ScoredEmbedding struct {
	Embedding model.TaskTitleEmbedding
	Score     float32
}

type EmbeddingRepository interface {
	SaveEmbedding(entry *model.TaskTitleEmbedding, vector []float32) error
	FindSimilar(projectID *uuid.UUID, queryVector []float32, topK int) ([]ScoredEmbedding, error)
}

type embeddingRepository struct {
	db *gorm.DB
}

func NewEmbeddingRepository(db *gorm.DB) EmbeddingRepository {
	return &embeddingRepository{db: db}
}

func (r *embeddingRepository) SaveEmbedding(entry *model.TaskTitleEmbedding, vector []float32) error {
	vecJSON, err := json.Marshal(vector)
	if err != nil {
		return err
	}
	entry.EmbeddingJSON = string(vecJSON)

	return r.db.Create(entry).Error
}

func (r *embeddingRepository) FindSimilar(projectID *uuid.UUID, queryVector []float32, topK int) ([]ScoredEmbedding, error) {
	var candidates []model.TaskTitleEmbedding

	// Fetch embeddings
	q := r.db.Model(&model.TaskTitleEmbedding{})
	if projectID != nil {
		q = q.Where("project_id = ? OR project_id IS NULL", *projectID)
	}

	err := q.Order("created_at desc").Limit(100).Find(&candidates).Error
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return []ScoredEmbedding{}, nil
	}

	scored := make([]ScoredEmbedding, 0, len(candidates))
	for _, c := range candidates {
		if c.EmbeddingJSON == "" {
			continue
		}

		var candidateVec []float32
		if err := json.Unmarshal([]byte(c.EmbeddingJSON), &candidateVec); err != nil {
			continue
		}

		sim := groq.CosineSimilarity(queryVector, candidateVec)
		scored = append(scored, ScoredEmbedding{
			Embedding: c,
			Score:     sim,
		})
	}

	// Sort descending by cosine similarity score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > topK {
		scored = scored[:topK]
	}

	return scored, nil
}
