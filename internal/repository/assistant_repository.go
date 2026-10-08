package repository

import (
	"context"
	"errors"

	"github.com/Kudzeri/job-prep-backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAssistantRequestFinal = errors.New("assistant request is already finalized")

type AssistantRepository struct{ pool *pgxpool.Pool }

func NewAssistantRepository(pool *pgxpool.Pool) *AssistantRepository {
	return &AssistantRepository{pool: pool}
}

func (r *AssistantRepository) Create(ctx context.Context, id string, userID int64) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO assistant_requests (id,user_id) VALUES ($1,$2)`, id, userID)
	return err
}

func (r *AssistantRepository) Get(ctx context.Context, id string, userID int64) (domain.AssistantRequest, error) {
	var item domain.AssistantRequest
	err := r.pool.QueryRow(ctx, `SELECT id::text,status,response_message,test_id,error_message,created_at,updated_at FROM assistant_requests WHERE id=$1 AND user_id=$2`, id, userID).Scan(&item.ID, &item.Status, &item.ResponseMessage, &item.TestID, &item.ErrorMessage, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *AssistantRepository) Fail(ctx context.Context, id, message string) error {
	result, err := r.pool.Exec(ctx, `UPDATE assistant_requests SET status='failed',error_message=$2,updated_at=NOW() WHERE id=$1 AND status='pending'`, id, message)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AssistantRepository) CompleteMessage(ctx context.Context, id, message string) error {
	result, err := r.pool.Exec(ctx, `UPDATE assistant_requests SET status='completed',response_message=$2,updated_at=NOW() WHERE id=$1 AND status='pending'`, id, message)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AssistantRepository) CompleteTest(ctx context.Context, id string, test domain.Test, questions []domain.Question) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var userID int64
	var status string
	err = tx.QueryRow(ctx, `SELECT user_id,status FROM assistant_requests WHERE id=$1 FOR UPDATE`, id).Scan(&userID, &status)
	if err != nil {
		return 0, err
	}
	if status != "pending" {
		return 0, ErrAssistantRequestFinal
	}
	var testID int64
	err = tx.QueryRow(ctx, `INSERT INTO tests (title,description,owner_user_id) VALUES ($1,$2,$3) RETURNING id`, test.Title, test.Description, userID).Scan(&testID)
	if err != nil {
		return 0, err
	}
	for _, q := range questions {
		if _, err = tx.Exec(ctx, `INSERT INTO test_questions (test_id,text,options,answer) VALUES ($1,$2,$3,$4)`, testID, q.Text, nullableJSON(q.Options), nullableJSON(q.Answer)); err != nil {
			return 0, err
		}
	}
	_, err = tx.Exec(ctx, `UPDATE assistant_requests SET status='completed',test_id=$2,response_message='',updated_at=NOW() WHERE id=$1`, id, testID)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return testID, nil
}

func IsAssistantRequestMissing(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
