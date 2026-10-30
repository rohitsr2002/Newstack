package models

import "time"

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Todo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Priority    Priority  `json:"priority"`
	DueDate     string    `json:"due_date,omitempty"` // YYYY-MM-DD
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TodoStats struct {
	Total       int `json:"total"`
	Completed   int `json:"completed"`
	Active      int `json:"active"`
	Overdue     int `json:"overdue"`
	HighCount   int `json:"high_priority"`
	MediumCount int `json:"medium_priority"`
	LowCount    int `json:"low_priority"`
}

type CreateTodoRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Priority    Priority `json:"priority"`
	DueDate     string   `json:"due_date"`
}

type UpdateTodoRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Priority    Priority `json:"priority"`
	DueDate     string   `json:"due_date"`
	Completed   *bool    `json:"completed,omitempty"`
}
