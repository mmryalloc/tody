package handler

import "net/http"

type Router struct {
	task *TaskHandler
}

func NewRouter(taskHandler *TaskHandler) *Router {
	return &Router{
		task: taskHandler,
	}
}

func (r *Router) Setup() http.Handler {
	mux := http.NewServeMux()

	const v1 = "/api/v1"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	mux.HandleFunc("POST "+v1+"/tasks", r.task.CreateTask)
	mux.HandleFunc("GET "+v1+"/tasks", r.task.ListTasks)
	mux.HandleFunc("GET "+v1+"/tasks/{id}", r.task.GetTask)
	mux.HandleFunc("PUT "+v1+"/tasks/{id}", r.task.UpdateTask)
	mux.HandleFunc("DELETE "+v1+"/tasks/{id}", r.task.DeleteTask)

	return mux
}
