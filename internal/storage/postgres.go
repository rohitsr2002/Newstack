package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"todo-app/internal/models"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database connection string: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	store := &PostgresStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run database migration: %w", err)
	}

	return store, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id VARCHAR(64) PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		category VARCHAR(100) NOT NULL DEFAULT 'General',
		priority VARCHAR(20) NOT NULL DEFAULT 'medium',
		due_date VARCHAR(20) NOT NULL DEFAULT '',
		completed BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`
	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	// Seed if empty
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&count)
	if err == nil && count == 0 {
		s.seedInitialData()
	}

	return nil
}

func (s *PostgresStore) seedInitialData() {
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
			Title:       "Connect Go backend to PostgreSQL",
			Description: "Configure database/sql connection pool, auto-migrations, and parameterized SQL queries.",
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
		_, _ = s.db.Exec(`
			INSERT INTO todos (id, title, description, category, priority, due_date, completed, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING
		`, t.ID, t.Title, t.Description, t.Category, t.Priority, t.DueDate, t.Completed, t.CreatedAt, t.UpdatedAt)
	}
}

func (s *PostgresStore) GetAll(search, status, priority, category, sortBy string) []*models.Todo {
	search = strings.ToLower(strings.TrimSpace(search))
	status = strings.ToLower(strings.TrimSpace(status))
	priority = strings.ToLower(strings.TrimSpace(priority))
	category = strings.ToLower(strings.TrimSpace(category))
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))

	query := "SELECT id, title, description, category, priority, due_date, completed, created_at, updated_at FROM todos WHERE 1=1"
	var args []any
	idx := 1

	if status == "active" {
		query += " AND completed = false"
	} else if status == "completed" {
		query += " AND completed = true"
	}

	if priority != "" && priority != "all" {
		query += fmt.Sprintf(" AND LOWER(priority) = $%d", idx)
		args = append(args, priority)
		idx++
	}

	if category != "" && category != "all" {
		query += fmt.Sprintf(" AND LOWER(category) = $%d", idx)
		args = append(args, category)
		idx++
	}

	if search != "" {
		pattern := "%" + search + "%"
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d OR category ILIKE $%d)", idx, idx, idx)
		args = append(args, pattern)
		idx++
	}

	switch sortBy {
	case "created_asc":
		query += " ORDER BY created_at ASC"
	case "due_date":
		query += " ORDER BY CASE WHEN due_date = '' THEN 1 ELSE 0 END, due_date ASC, created_at DESC"
	case "priority":
		query += " ORDER BY CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END ASC, created_at DESC"
	case "title":
		query += " ORDER BY LOWER(title) ASC"
	case "created_desc":
		fallthrough
	default:
		query += " ORDER BY created_at DESC"
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("PostgreSQL query error: %v", err)
		return []*models.Todo{}
	}
	defer rows.Close()

	result := make([]*models.Todo, 0)
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.Category,
			&t.Priority,
			&t.DueDate,
			&t.Completed,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			log.Printf("PostgreSQL scan error: %v", err)
			continue
		}
		result = append(result, &t)
	}

	return result
}

func (s *PostgresStore) GetByID(id string) (*models.Todo, bool) {
	query := "SELECT id, title, description, category, priority, due_date, completed, created_at, updated_at FROM todos WHERE id = $1"
	var t models.Todo
	err := s.db.QueryRow(query, id).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Category,
		&t.Priority,
		&t.DueDate,
		&t.Completed,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		log.Printf("PostgreSQL GetByID error: %v", err)
		return nil, false
	}
	return &t, true
}

func (s *PostgresStore) Create(req models.CreateTodoRequest) (*models.Todo, error) {
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

	query := `
		INSERT INTO todos (id, title, description, category, priority, due_date, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.db.Exec(query, todo.ID, todo.Title, todo.Description, todo.Category, todo.Priority, todo.DueDate, todo.Completed, todo.CreatedAt, todo.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert todo: %w", err)
	}

	return todo, nil
}

func (s *PostgresStore) Update(id string, req models.UpdateTodoRequest) (*models.Todo, error) {
	current, found := s.GetByID(id)
	if !found {
		return nil, errors.New("todo not found")
	}

	trimmedTitle := strings.TrimSpace(req.Title)
	if trimmedTitle != "" {
		current.Title = trimmedTitle
	}

	current.Description = strings.TrimSpace(req.Description)

	if req.Category != "" {
		current.Category = strings.TrimSpace(req.Category)
	}

	if req.Priority == models.PriorityHigh || req.Priority == models.PriorityMedium || req.Priority == models.PriorityLow {
		current.Priority = req.Priority
	}

	current.DueDate = strings.TrimSpace(req.DueDate)

	if req.Completed != nil {
		current.Completed = *req.Completed
	}

	current.UpdatedAt = time.Now()

	query := `
		UPDATE todos
		SET title = $1, description = $2, category = $3, priority = $4, due_date = $5, completed = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := s.db.Exec(query, current.Title, current.Description, current.Category, current.Priority, current.DueDate, current.Completed, current.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return current, nil
}

func (s *PostgresStore) ToggleComplete(id string) (*models.Todo, error) {
	query := `
		UPDATE todos
		SET completed = NOT completed, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, description, category, priority, due_date, completed, created_at, updated_at
	`
	var t models.Todo
	err := s.db.QueryRow(query, id).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Category,
		&t.Priority,
		&t.DueDate,
		&t.Completed,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("todo not found")
		}
		return nil, fmt.Errorf("failed to toggle completion: %w", err)
	}

	return &t, nil
}

func (s *PostgresStore) Delete(id string) error {
	res, err := s.db.Exec("DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("todo not found")
	}

	return nil
}

func (s *PostgresStore) ClearCompleted() (int, error) {
	res, err := s.db.Exec("DELETE FROM todos WHERE completed = true")
	if err != nil {
		return 0, fmt.Errorf("failed to clear completed todos: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rows), nil
}

func (s *PostgresStore) GetStats() models.TodoStats {
	todayStr := time.Now().Format("2006-01-02")
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE completed = true) as completed,
			COUNT(*) FILTER (WHERE completed = false) as active,
			COUNT(*) FILTER (WHERE completed = false AND due_date != '' AND due_date < $1) as overdue,
			COUNT(*) FILTER (WHERE priority = 'high') as high_count,
			COUNT(*) FILTER (WHERE priority = 'medium') as medium_count,
			COUNT(*) FILTER (WHERE priority = 'low') as low_count
		FROM todos;
	`
	var stats models.TodoStats
	err := s.db.QueryRow(query, todayStr).Scan(
		&stats.Total,
		&stats.Completed,
		&stats.Active,
		&stats.Overdue,
		&stats.HighCount,
		&stats.MediumCount,
		&stats.LowCount,
	)
	if err != nil {
		log.Printf("PostgreSQL GetStats error: %v", err)
	}

	stats.DatabaseType = "PostgreSQL"
	return stats
}
