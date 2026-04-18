package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mmryalloc/todo-app/internal/entity"
	"github.com/mmryalloc/todo-app/internal/service"
)

type createTaskRequest struct {
	Title       string `json:"title" validate:"required,notblank,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type updateTaskRequest struct {
	Title       *string `json:"title" validate:"omitempty,notblank,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Completed   *bool   `json:"completed" validate:"omitempty"`
}

type taskResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type TaskService interface {
	CreateTask(ctx context.Context, t service.CreateTaskInput) (entity.Task, error)
	ListTasks(ctx context.Context, page, limit int) ([]entity.Task, int, error)
	GetTask(ctx context.Context, id int64) (entity.Task, error)
	UpdateTask(ctx context.Context, id int64, in service.UpdateTaskInput) (entity.Task, error)
	DeleteTask(ctx context.Context, id int64) error
}

type TaskHandler struct {
	svc TaskService
}

func NewTaskHandler(svc TaskService) *TaskHandler {
	return &TaskHandler{
		svc: svc,
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if !bind(w, r, &req) {
		return
	}

	t, err := h.svc.CreateTask(r.Context(), service.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		slog.Error("handler task create", "error", err)
		internalError(w, "failed to create task")
		return
	}

	created(w, taskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
	})
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10
	if p, err := strconv.Atoi(pageStr); err == nil {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil {
		limit = l
	}

	tasks, total, err := h.svc.ListTasks(r.Context(), page, limit)
	if err != nil {
		slog.Error("handler list tasks", "error", err)
		internalError(w, "failed to list tasks")
		return
	}

	res := make([]taskResponse, len(tasks))
	for i, t := range tasks {
		res[i] = taskResponse{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Completed:   t.Completed,
		}
	}

	okPaginated(w, res, page, limit, total)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(w, errorCodeBadRequest, "invalid task id", nil)
		return
	}

	t, err := h.svc.GetTask(r.Context(), id)
	if err != nil {
		slog.Error("handler get task", "error", err)
		notFound(w, "task not found")
		return
	}

	ok(w, taskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
	})
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(w, errorCodeBadRequest, "invalid task id", nil)
		return
	}

	var req updateTaskRequest
	if !bind(w, r, &req) {
		return
	}

	t, err := h.svc.UpdateTask(r.Context(), id, service.UpdateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Completed:   req.Completed,
	})
	if err != nil {
		slog.Error("handler update task", "error", err)
		notFound(w, "task not found")
		return
	}

	ok(w, taskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
	})
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequest(w, errorCodeBadRequest, "invalid task id", nil)
		return
	}

	if err := h.svc.DeleteTask(r.Context(), id); err != nil {
		slog.Error("handler delete task", "error", err)
		notFound(w, "task not found")
		return
	}

	var empty struct{}
	ok(w, empty)
}
