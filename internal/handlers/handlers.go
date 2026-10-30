package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"todo-app/internal/models"
	"todo-app/internal/storage"
)

type TodoHandler struct {
	store *storage.Store
}

func NewTodoHandler(store *storage.Store) *TodoHandler {
	return &TodoHandler{store: store}
}

type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ApiResponse{
		Success: false,
		Message: message,
	})
}

// GET /api/todos
func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")
	category := r.URL.Query().Get("category")
	sort := r.URL.Query().Get("sort")

	todos := h.store.GetAll(q, status, priority, category, sort)
	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Data:    todos,
	})
}

// GET /api/todos/{id}
func (h *TodoHandler) GetTodoByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	todo, found := h.store.GetByID(id)
	if !found {
		writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Data:    todo,
	})
}

// POST /api/todos
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Malformed JSON request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "Task title is required")
		return
	}

	todo, err := h.store.Create(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "Task created successfully",
		Data:    todo,
	})
}

// PUT /api/todos/{id}
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req models.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Malformed JSON request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "Task title cannot be empty")
		return
	}

	todo, err := h.store.Update(id, req)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Task updated successfully",
		Data:    todo,
	})
}

// PATCH /api/todos/{id}/toggle
func (h *TodoHandler) ToggleTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	todo, err := h.store.ToggleComplete(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Task completion status updated",
		Data:    todo,
	})
}

// DELETE /api/todos/{id}
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	if err := h.store.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Task deleted successfully",
	})
}

// POST /api/todos/clear-completed
func (h *TodoHandler) ClearCompleted(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.ClearCompleted()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Completed tasks cleared",
		Data: map[string]int{
			"cleared_count": count,
		},
	})
}

// GET /api/stats
func (h *TodoHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.store.GetStats()
	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Data:    stats,
	})
}
