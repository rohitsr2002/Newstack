package storage

import "todo-app/internal/models"

// Store defines the interface for Todo persistence (PostgreSQL or JSON file store).
type Store interface {
	GetAll(search, status, priority, category, sortBy string) []*models.Todo
	GetByID(id string) (*models.Todo, bool)
	Create(req models.CreateTodoRequest) (*models.Todo, error)
	Update(id string, req models.UpdateTodoRequest) (*models.Todo, error)
	ToggleComplete(id string) (*models.Todo, error)
	Delete(id string) error
	ClearCompleted() (int, error)
	GetStats() models.TodoStats
}
