package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"todo-app/internal/models"
)

type JSONStore struct {
	mu       sync.RWMutex
	filePath string
	todos    map[string]*models.Todo
}

func NewJSONStore(filePath string) (*JSONStore, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for todos: %w", err)
	}

	store := &JSONStore{
		filePath: filePath,
		todos:    make(map[string]*models.Todo),
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		store.seedInitialData()
		if err := store.saveToFileLocked(); err != nil {
			return nil, fmt.Errorf("failed to save seed data: %w", err)
		}
	} else {
		if err := store.loadFromFile(); err != nil {
			return nil, fmt.Errorf("failed to load todos: %w", err)
		}
	}

	return store, nil
}

func generateID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *JSONStore) seedInitialData() {
	now := time.Now()
	today := now.Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	nextWeek := now.AddDate(0, 0, 7).Format("2006-01-02")

	seeds := []*models.Todo{
		{
			ID:          "todo-001",
			Title:       "Design modern dark-mode CRUD interface",
			Description: "Implement glassmorphism, responsive cards, and accessible form controls for seamless task management.",
			Category:    "Design",
			Priority:    models.PriorityHigh,
			DueDate:     today,
			Completed:   true,
			CreatedAt:   now.Add(-4 * time.Hour),
			UpdatedAt:   now.Add(-1 * time.Hour),
		},
		{
			ID:          "todo-002",
			Title:       "Implement Go REST API endpoints",
			Description: "Set up net/http routing, JSON marshaling, in-memory sync storage, and persistent JSON disk backup.",
			Category:    "Backend",
			Priority:    models.PriorityHigh,
			DueDate:     today,
			Completed:   true,
			CreatedAt:   now.Add(-3 * time.Hour),
			UpdatedAt:   now.Add(-30 * time.Minute),
		},
		{
			ID:          "todo-003",
			Title:       "Add filtering, searching, and sorting",
			Description: "Provide real-time keyword search, priority chips, category filters, and overdue indicators.",
			Category:    "Frontend",
			Priority:    models.PriorityMedium,
			DueDate:     tomorrow,
			Completed:   false,
			CreatedAt:   now.Add(-2 * time.Hour),
			UpdatedAt:   now.Add(-2 * time.Hour),
		},
		{
			ID:          "todo-004",
			Title:       "Prepare weekly sprint report",
			Description: "Summarize delivered milestones, architecture improvements, and performance metrics for the team.",
			Category:    "Work",
			Priority:    models.PriorityMedium,
			DueDate:     nextWeek,
			Completed:   false,
			CreatedAt:   now.Add(-1 * time.Hour),
			UpdatedAt:   now.Add(-1 * time.Hour),
		},
		{
			ID:          "todo-005",
			Title:       "Evening 30-minute cardio workout",
			Description: "Stay active, stretch, and track heart rate targets.",
			Category:    "Health",
			Priority:    models.PriorityLow,
			DueDate:     today,
			Completed:   false,
			CreatedAt:   now.Add(-45 * time.Minute),
			UpdatedAt:   now.Add(-45 * time.Minute),
		},
	}

	for _, t := range seeds {
		s.todos[t.ID] = t
	}
}

func (s *JSONStore) loadFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var list []*models.Todo
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	s.todos = make(map[string]*models.Todo)
	for _, t := range list {
		s.todos[t.ID] = t
	}
	return nil
}

func (s *JSONStore) saveToFileLocked() error {
	list := make([]*models.Todo, 0, len(s.todos))
	for _, t := range s.todos {
		list = append(list, t)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *JSONStore) GetAll(search, status, priority, category, sortBy string) []*models.Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	search = strings.ToLower(strings.TrimSpace(search))
	status = strings.ToLower(strings.TrimSpace(status))
	priority = strings.ToLower(strings.TrimSpace(priority))
	category = strings.ToLower(strings.TrimSpace(category))
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))

	result := make([]*models.Todo, 0, len(s.todos))

	for _, t := range s.todos {
		// Filter by status
		if status == "active" && t.Completed {
			continue
		}
		if status == "completed" && !t.Completed {
			continue
		}

		// Filter by priority
		if priority != "" && priority != "all" && string(t.Priority) != priority {
			continue
		}

		// Filter by category
		if category != "" && category != "all" && strings.ToLower(t.Category) != category {
			continue
		}

		// Search term match
		if search != "" {
			titleMatch := strings.Contains(strings.ToLower(t.Title), search)
			descMatch := strings.Contains(strings.ToLower(t.Description), search)
			catMatch := strings.Contains(strings.ToLower(t.Category), search)
			if !titleMatch && !descMatch && !catMatch {
				continue
			}
		}

		// Copy pointer to return list
		clone := *t
		result = append(result, &clone)
	}

	// Sorting
	sort.Slice(result, func(i, j int) bool {
		switch sortBy {
		case "created_asc":
			return result[i].CreatedAt.Before(result[j].CreatedAt)
		case "due_date":
			if result[i].DueDate == "" && result[j].DueDate != "" {
				return false
			}
			if result[i].DueDate != "" && result[j].DueDate == "" {
				return true
			}
			return result[i].DueDate < result[j].DueDate
		case "priority":
			pScore := func(p models.Priority) int {
				switch p {
				case models.PriorityHigh:
					return 3
				case models.PriorityMedium:
					return 2
				case models.PriorityLow:
					return 1
				default:
					return 0
				}
			}
			return pScore(result[i].Priority) > pScore(result[j].Priority)
		case "title":
			return strings.ToLower(result[i].Title) < strings.ToLower(result[j].Title)
		case "created_desc":
			fallthrough
		default:
			return result[i].CreatedAt.After(result[j].CreatedAt)
		}
	})

	return result
}

func (s *JSONStore) GetByID(id string) (*models.Todo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, exists := s.todos[id]
	if !exists {
		return nil, false
	}
	clone := *t
	return &clone, true
}

func (s *JSONStore) Create(req models.CreateTodoRequest) (*models.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trimmedTitle := strings.TrimSpace(req.Title)
	if trimmedTitle == "" {
		return nil, errors.New("title cannot be empty")
	}

	priority := req.Priority
	if priority != models.PriorityHigh && priority != models.PriorityMedium && priority != models.PriorityLow {
		priority = models.PriorityMedium
	}

	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = "General"
	}

	now := time.Now()
	todo := &models.Todo{
		ID:          "todo-" + generateID(),
		Title:       trimmedTitle,
		Description: strings.TrimSpace(req.Description),
		Category:    category,
		Priority:    priority,
		DueDate:     strings.TrimSpace(req.DueDate),
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.todos[todo.ID] = todo
	if err := s.saveToFileLocked(); err != nil {
		return nil, fmt.Errorf("failed to persist todo: %w", err)
	}

	clone := *todo
	return &clone, nil
}

func (s *JSONStore) Update(id string, req models.UpdateTodoRequest) (*models.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, errors.New("todo not found")
	}

	trimmedTitle := strings.TrimSpace(req.Title)
	if trimmedTitle != "" {
		todo.Title = trimmedTitle
	}

	todo.Description = strings.TrimSpace(req.Description)

	if req.Category != "" {
		todo.Category = strings.TrimSpace(req.Category)
	}

	if req.Priority == models.PriorityHigh || req.Priority == models.PriorityMedium || req.Priority == models.PriorityLow {
		todo.Priority = req.Priority
	}

	todo.DueDate = strings.TrimSpace(req.DueDate)

	if req.Completed != nil {
		todo.Completed = *req.Completed
	}

	todo.UpdatedAt = time.Now()

	if err := s.saveToFileLocked(); err != nil {
		return nil, fmt.Errorf("failed to persist update: %w", err)
	}

	clone := *todo
	return &clone, nil
}

func (s *JSONStore) ToggleComplete(id string) (*models.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, exists := s.todos[id]
	if !exists {
		return nil, errors.New("todo not found")
	}

	todo.Completed = !todo.Completed
	todo.UpdatedAt = time.Now()

	if err := s.saveToFileLocked(); err != nil {
		return nil, fmt.Errorf("failed to persist toggle: %w", err)
	}

	clone := *todo
	return &clone, nil
}

func (s *JSONStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.todos[id]; !exists {
		return errors.New("todo not found")
	}

	delete(s.todos, id)
	return s.saveToFileLocked()
}

func (s *JSONStore) ClearCompleted() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, t := range s.todos {
		if t.Completed {
			delete(s.todos, id)
			count++
		}
	}

	if count > 0 {
		if err := s.saveToFileLocked(); err != nil {
			return 0, fmt.Errorf("failed to persist deletions: %w", err)
		}
	}

	return count, nil
}

func (s *JSONStore) GetStats() models.TodoStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var stats models.TodoStats
	stats.Total = len(s.todos)
	todayStr := time.Now().Format("2006-01-02")

	for _, t := range s.todos {
		if t.Completed {
			stats.Completed++
		} else {
			stats.Active++
			if t.DueDate != "" && t.DueDate < todayStr {
				stats.Overdue++
			}
		}

		switch t.Priority {
		case models.PriorityHigh:
			stats.HighCount++
		case models.PriorityMedium:
			stats.MediumCount++
		case models.PriorityLow:
			stats.LowCount++
		}
	}

	stats.DatabaseType = "Local JSON File"
	return stats
}
