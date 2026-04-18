package repository

import (
	"context"
	"database/sql"
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
		INSERT INTO tasks (title, description, completed)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		t.Title,
		t.Description,
		t.Completed,
	).Scan(&t.ID)

	if err != nil {
		return fmt.Errorf("repository task create: %w", err)
	}

	return nil
}

func (r *taskRepository) List(ctx context.Context, limit, offset int) ([]entity.Task, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("repository task list count: %w", err)
	}

	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM tasks
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository task list: %w", err)
	}
	defer rows.Close()

	tasks := []entity.Task{}
	for rows.Next() {
		var t entity.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("repository task list scan: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository task list iteration: %w", err)
	}

	return tasks, total, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id int64) (entity.Task, error) {
	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	var t entity.Task
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.Task{}, fmt.Errorf("repository task get: not found: %w", err)
		}
		return entity.Task{}, fmt.Errorf("repository task get: %w", err)
	}

	return t, nil
}

func (r *taskRepository) Update(ctx context.Context, t *entity.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, completed = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		t.Title,
		t.Description,
		t.Completed,
		t.ID,
	).Scan(&t.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("repository task update: not found: %w", err)
		}
		return fmt.Errorf("repository task update: %w", err)
	}

	return nil
}

func (r *taskRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository task delete: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository task delete rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("repository task delete: not found: %w", sql.ErrNoRows)
	}

	return nil
}
