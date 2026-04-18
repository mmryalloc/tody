package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mmryalloc/todo-app/internal/entity"
)

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *taskRepository {
	return &taskRepository{
		db: db,
	}
}

func (r *taskRepository) Create(ctx context.Context, t *entity.Task) error {
	query := `
		INSERT INTO tasks (user_id, title, description, completed)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		t.UserID,
		t.Title,
		t.Description,
		t.Completed,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository task create: %w", err)
	}

	return nil
}

func (r *taskRepository) List(ctx context.Context, userID int64, limit, offset int) ([]entity.Task, int, error) {
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at,
		       COUNT(*) OVER () AS total
		FROM tasks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository task list: %w", err)
	}
	defer rows.Close()

	var (
		tasks = []entity.Task{}
		total int
	)
	for rows.Next() {
		var t entity.Task
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Title, &t.Description, &t.Completed,
			&t.CreatedAt, &t.UpdatedAt, &total,
		); err != nil {
			return nil, 0, fmt.Errorf("repository task list scan: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository task list iteration: %w", err)
	}

	if len(tasks) == 0 {
		if err := r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM tasks WHERE user_id = $1`, userID,
		).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("repository task list count: %w", err)
		}
	}

	return tasks, total, nil
}

func (r *taskRepository) GetByID(ctx context.Context, userID, id int64) (entity.Task, error) {
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM tasks
		WHERE id = $1 AND user_id = $2
	`
	var t entity.Task
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Task{}, entity.ErrTaskNotFound
		}
		return entity.Task{}, fmt.Errorf("repository task get: %w", err)
	}

	return t, nil
}

func (r *taskRepository) Update(ctx context.Context, t *entity.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, completed = $3, updated_at = NOW()
		WHERE id = $4 AND user_id = $5
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		t.Title,
		t.Description,
		t.Completed,
		t.ID,
		t.UserID,
	).Scan(&t.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrTaskNotFound
		}
		return fmt.Errorf("repository task update: %w", err)
	}

	return nil
}

func (r *taskRepository) Delete(ctx context.Context, userID, id int64) error {
	query := `DELETE FROM tasks WHERE id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repository task delete: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository task delete rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return entity.ErrTaskNotFound
	}

	return nil
}
