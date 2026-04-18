package service

import (
	"context"

	"github.com/mmryalloc/todo-app/internal/entity"
)

type CreateTaskInput struct {
	Title       string
	Description string
}

type UpdateTaskInput struct {
	Title       *string
	Description *string
	Completed   *bool
}

type TaskRepository interface {
	Create(ctx context.Context, t *entity.Task) error
	List(ctx context.Context, userID int64, limit, offset int) ([]entity.Task, int, error)
	GetByID(ctx context.Context, userID, id int64) (entity.Task, error)
	Update(ctx context.Context, t *entity.Task) error
	Delete(ctx context.Context, userID, id int64) error
}

type taskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *taskService {
	return &taskService{
		repo: repo,
	}
}

func (s *taskService) CreateTask(ctx context.Context, userID int64, t CreateTaskInput) (entity.Task, error) {
	task := entity.Task{
		UserID:      userID,
		Title:       t.Title,
		Description: t.Description,
	}
	if err := s.repo.Create(ctx, &task); err != nil {
		return entity.Task{}, err
	}
	return task, nil
}

func (s *taskService) ListTasks(ctx context.Context, userID int64, page, limit int) ([]entity.Task, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	return s.repo.List(ctx, userID, limit, offset)
}

func (s *taskService) GetTask(ctx context.Context, userID, id int64) (entity.Task, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *taskService) UpdateTask(ctx context.Context, userID, id int64, in UpdateTaskInput) (entity.Task, error) {
	task, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return entity.Task{}, err
	}

	if in.Title != nil {
		task.Title = *in.Title
	}
	if in.Description != nil {
		task.Description = *in.Description
	}
	if in.Completed != nil {
		task.Completed = *in.Completed
	}

	if err := s.repo.Update(ctx, &task); err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (s *taskService) DeleteTask(ctx context.Context, userID, id int64) error {
	return s.repo.Delete(ctx, userID, id)
}
