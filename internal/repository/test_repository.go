package repository

import (
	"context"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TestRepository struct{ pool *pgxpool.Pool }

func NewTestRepository(pool *pgxpool.Pool) *TestRepository { return &TestRepository{pool: pool} }

func (r *TestRepository) Create(ctx context.Context, title, description string) (domain.Test, error) {
	var t domain.Test
	err := r.pool.QueryRow(ctx, `INSERT INTO tests (title, description) VALUES ($1, $2) RETURNING id, title, description`, title, description).Scan(&t.ID, &t.Title, &t.Description)
	return t, err
}

func (r *TestRepository) AddQuestion(ctx context.Context, testID int64, q domain.Question) (domain.QuestionResult, error) {
	var result domain.QuestionResult
	err := r.pool.QueryRow(ctx, `INSERT INTO test_questions (test_id, text, options, answer) VALUES ($1, $2, $3, $4) RETURNING id, test_id, text, options, answer`, testID, q.Text, nullableJSON(q.Options), nullableJSON(q.Answer)).Scan(&result.ID, &result.TestID, &result.Text, &result.Options, &result.Answer)
	return result, err
}

func nullableJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func (r *TestRepository) AddQuestions(ctx context.Context, testID int64, questions []domain.Question) ([]domain.QuestionResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	results := make([]domain.QuestionResult, 0, len(questions))
	for _, q := range questions {
		var result domain.QuestionResult
		err := tx.QueryRow(ctx, `INSERT INTO test_questions (test_id, text, options, answer) VALUES ($1, $2, $3, $4) RETURNING id, test_id, text, options, answer`, testID, q.Text, nullableJSON(q.Options), nullableJSON(q.Answer)).Scan(&result.ID, &result.TestID, &result.Text, &result.Options, &result.Answer)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return results, nil
}
func (r *TestRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tests WHERE id=$1)`, id).Scan(&exists)
	return exists, err
}
