package repository

import (
	"context"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TestRepository struct{ pool *pgxpool.Pool }

func NewTestRepository(pool *pgxpool.Pool) *TestRepository { return &TestRepository{pool: pool} }

func (r *TestRepository) Create(ctx context.Context, title, description string) (domain.Test, error) {
	var t domain.Test
	err := r.pool.QueryRow(ctx, `INSERT INTO tests (title, description) VALUES ($1, $2) RETURNING id, title, description`, title, description).Scan(&t.ID, &t.Title, &t.Description)
	return t, err
}

func (r *TestRepository) CreateForUser(ctx context.Context, userID int64, title, description string) (domain.Test, error) {
	var t domain.Test
	err := r.pool.QueryRow(ctx, `INSERT INTO tests (title, description, owner_user_id) VALUES ($1, $2, $3) RETURNING id, title, description`, title, description, userID).Scan(&t.ID, &t.Title, &t.Description)
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

func (r *TestRepository) List(ctx context.Context) ([]domain.TestSummary, error) {
	rows, err := r.pool.Query(ctx, `SELECT t.id, t.title, t.description, COUNT(q.id)::int FROM tests t LEFT JOIN test_questions q ON q.test_id=t.id GROUP BY t.id ORDER BY t.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.TestSummary, 0)
	for rows.Next() {
		var item domain.TestSummary
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.QuestionCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *TestRepository) ListForUser(ctx context.Context, userID int64) ([]domain.TestSummary, error) {
	rows, err := r.pool.Query(ctx, `SELECT t.id, t.title, t.description, COUNT(q.id)::int FROM tests t LEFT JOIN test_questions q ON q.test_id=t.id WHERE t.owner_user_id=$1 GROUP BY t.id ORDER BY t.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.TestSummary, 0)
	for rows.Next() {
		var item domain.TestSummary
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.QuestionCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *TestRepository) Get(ctx context.Context, id int64) (domain.TestDetails, error) {
	var details domain.TestDetails
	err := r.pool.QueryRow(ctx, `SELECT id, title, description FROM tests WHERE id=$1`, id).Scan(&details.ID, &details.Title, &details.Description)
	if err != nil {
		return details, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id, test_id, text, options, answer FROM test_questions WHERE test_id=$1 ORDER BY id`, id)
	if err != nil {
		return details, err
	}
	defer rows.Close()
	details.Questions = make([]domain.QuestionResult, 0)
	for rows.Next() {
		var item domain.QuestionResult
		if err := rows.Scan(&item.ID, &item.TestID, &item.Text, &item.Options, &item.Answer); err != nil {
			return details, err
		}
		details.Questions = append(details.Questions, item)
	}
	return details, rows.Err()
}

func (r *TestRepository) GetForUser(ctx context.Context, id, userID int64) (domain.TestDetails, error) {
	var details domain.TestDetails
	err := r.pool.QueryRow(ctx, `SELECT id, title, description FROM tests WHERE id=$1 AND owner_user_id=$2`, id, userID).Scan(&details.ID, &details.Title, &details.Description)
	if err != nil {
		return details, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id, test_id, text, options, answer FROM test_questions WHERE test_id=$1 ORDER BY id`, id)
	if err != nil {
		return details, err
	}
	defer rows.Close()
	details.Questions = make([]domain.QuestionResult, 0)
	for rows.Next() {
		var item domain.QuestionResult
		if err := rows.Scan(&item.ID, &item.TestID, &item.Text, &item.Options, &item.Answer); err != nil {
			return details, err
		}
		details.Questions = append(details.Questions, item)
	}
	return details, rows.Err()
}

func (r *TestRepository) AddQuestionForUser(ctx context.Context, testID, userID int64, q domain.Question) (domain.QuestionResult, error) {
	var result domain.QuestionResult
	err := r.pool.QueryRow(ctx, `INSERT INTO test_questions (test_id, text, options, answer) SELECT id, $3, $4, $5 FROM tests WHERE id=$1 AND owner_user_id=$2 RETURNING id, test_id, text, options, answer`, testID, userID, q.Text, nullableJSON(q.Options), nullableJSON(q.Answer)).Scan(&result.ID, &result.TestID, &result.Text, &result.Options, &result.Answer)
	return result, err
}

func (r *TestRepository) AddQuestionsForUser(ctx context.Context, testID, userID int64, questions []domain.Question) ([]domain.QuestionResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tests WHERE id=$1 AND owner_user_id=$2)`, testID, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, pgx.ErrNoRows
	}
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
