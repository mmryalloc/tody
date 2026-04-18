package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mmryalloc/todo-app/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTaskRepository struct {
	CreateFunc  func(ctx context.Context, t *entity.Task) error
	ListFunc    func(ctx context.Context, limit, offset int) ([]entity.Task, int, error)
	GetByIDFunc func(ctx context.Context, id int64) (entity.Task, error)
	UpdateFunc  func(ctx context.Context, t *entity.Task) error
	DeleteFunc  func(ctx context.Context, id int64) error
}

func (m *mockTaskRepository) Create(ctx context.Context, t *entity.Task) error {
	return m.CreateFunc(ctx, t)
}

func (m *mockTaskRepository) List(ctx context.Context, limit, offset int) ([]entity.Task, int, error) {
	return m.ListFunc(ctx, limit, offset)
}

func (m *mockTaskRepository) GetByID(ctx context.Context, id int64) (entity.Task, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *mockTaskRepository) Update(ctx context.Context, t *entity.Task) error {
	return m.UpdateFunc(ctx, t)
}

func (m *mockTaskRepository) Delete(ctx context.Context, id int64) error {
	return m.DeleteFunc(ctx, id)
}

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateTaskInput
		mock    func(t *testing.T) TaskRepository
		want    entity.Task
		wantErr bool
	}{
		{
			name: "success",
			input: CreateTaskInput{
				Title:       "Test Title",
				Description: "Test Description",
			},
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					CreateFunc: func(ctx context.Context, task *entity.Task) error {
						task.ID = 1
						return nil
					},
				}
			},
			want: entity.Task{
				ID:          1,
				Title:       "Test Title",
				Description: "Test Description",
			},
			wantErr: false,
		},
		{
			name: "error",
			input: CreateTaskInput{
				Title:       "Test Title",
				Description: "Test Description",
			},
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					CreateFunc: func(ctx context.Context, task *entity.Task) error {
						return errors.New("db error")
					},
				}
			},
			want:    entity.Task{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTaskService(tt.mock(t))
			got, err := s.CreateTask(context.Background(), tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListTasks(t *testing.T) {
	mockTasks := []entity.Task{
		{ID: 1, Title: "Task 1", Description: "Desc 1"},
		{ID: 2, Title: "Task 2", Description: "Desc 2"},
	}

	tests := []struct {
		name      string
		page      int
		limit     int
		mock      func(t *testing.T) TaskRepository
		wantTasks []entity.Task
		wantTotal int
		wantErr   bool
	}{
		{
			name:  "success",
			page:  1,
			limit: 10,
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					ListFunc: func(ctx context.Context, limit, offset int) ([]entity.Task, int, error) {
						assert.Equal(t, 10, limit)
						assert.Equal(t, 0, offset)
						return mockTasks, 2, nil
					},
				}
			},
			wantTasks: mockTasks,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:  "invalid pagination params",
			page:  0,
			limit: 101,
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					ListFunc: func(ctx context.Context, limit, offset int) ([]entity.Task, int, error) {
						assert.Equal(t, 10, limit)
						assert.Equal(t, 0, offset)
						return mockTasks, 2, nil
					},
				}
			},
			wantTasks: mockTasks,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:  "error",
			page:  1,
			limit: 10,
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					ListFunc: func(ctx context.Context, limit, offset int) ([]entity.Task, int, error) {
						return nil, 0, errors.New("db error")
					},
				}
			},
			wantTasks: nil,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTaskService(tt.mock(t))
			gotTasks, gotTotal, err := s.ListTasks(context.Background(), tt.page, tt.limit)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTasks, gotTasks)
			assert.Equal(t, tt.wantTotal, gotTotal)
		})
	}
}

func TestGetTask(t *testing.T) {
	mockTask := entity.Task{ID: 1, Title: "Task 1", Description: "Desc 1"}

	tests := []struct {
		name    string
		id      int64
		mock    func(t *testing.T) TaskRepository
		want    entity.Task
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					GetByIDFunc: func(ctx context.Context, id int64) (entity.Task, error) {
						return mockTask, nil
					},
				}
			},
			want:    mockTask,
			wantErr: false,
		},
		{
			name: "error not found",
			id:   1,
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					GetByIDFunc: func(ctx context.Context, id int64) (entity.Task, error) {
						return entity.Task{}, errors.New("not found")
					},
				}
			},
			want:    entity.Task{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTaskService(tt.mock(t))
			got, err := s.GetTask(context.Background(), tt.id)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateTask(t *testing.T) {
	mockTask := entity.Task{ID: 1, Title: "Task 1", Description: "Desc 1", Completed: false}

	newTitle := "New Title"
	newDescription := "New Desc"
	newCompleted := true

	tests := []struct {
		name    string
		id      int64
		input   UpdateTaskInput
		mock    func(t *testing.T) TaskRepository
		want    entity.Task
		wantErr bool
	}{
		{
			name: "success update all fields",
			id:   1,
			input: UpdateTaskInput{
				Title:       &newTitle,
				Description: &newDescription,
				Completed:   &newCompleted,
			},
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					GetByIDFunc: func(ctx context.Context, id int64) (entity.Task, error) {
						return mockTask, nil
					},
					UpdateFunc: func(ctx context.Context, task *entity.Task) error {
						return nil
					},
				}
			},
			want: entity.Task{
				ID:          1,
				Title:       newTitle,
				Description: newDescription,
				Completed:   newCompleted,
			},
			wantErr: false,
		},
		{
			name: "success partial update",
			id:   1,
			input: UpdateTaskInput{
				Completed: &newCompleted,
			},
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					GetByIDFunc: func(ctx context.Context, id int64) (entity.Task, error) {
						return mockTask, nil
					},
					UpdateFunc: func(ctx context.Context, task *entity.Task) error {
						return nil
					},
				}
			},
			want: entity.Task{
				ID:          1,
				Title:       mockTask.Title,
				Description: mockTask.Description,
				Completed:   newCompleted,
			},
			wantErr: false,
		},
		{
			name: "error get by id",
			id:   1,
			input: UpdateTaskInput{
				Title: &newTitle,
			},
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					GetByIDFunc: func(ctx context.Context, id int64) (entity.Task, error) {
						return entity.Task{}, errors.New("not found")
					},
				}
			},
			want:    entity.Task{},
			wantErr: true,
		},
		{
			name: "error update",
			id:   1,
			input: UpdateTaskInput{
				Title: &newTitle,
			},
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					GetByIDFunc: func(ctx context.Context, id int64) (entity.Task, error) {
						return mockTask, nil
					},
					UpdateFunc: func(ctx context.Context, task *entity.Task) error {
						return errors.New("update error")
					},
				}
			},
			want:    entity.Task{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTaskService(tt.mock(t))
			got, err := s.UpdateTask(context.Background(), tt.id, tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		mock    func(t *testing.T) TaskRepository
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					DeleteFunc: func(ctx context.Context, id int64) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "error",
			id:   1,
			mock: func(t *testing.T) TaskRepository {
				return &mockTaskRepository{
					DeleteFunc: func(ctx context.Context, id int64) error {
						return errors.New("delete error")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTaskService(tt.mock(t))
			err := s.DeleteTask(context.Background(), tt.id)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
