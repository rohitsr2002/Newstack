package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"todo-app/internal/handlers"
	"todo-app/internal/storage"
)

// loadEnv loads key-value pairs from a .env file into environment variables
func loadEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path == "/" || len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			log.Printf("[%s] %s - %v", r.Method, r.URL.Path, time.Since(start))
		}
	})
}

func main() {
	loadEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	var store storage.Store
	var dbType string

	if databaseURL != "" {
		log.Println("🔌 Connecting to PostgreSQL database...")
		pgStore, err := storage.NewPostgresStore(databaseURL)
		if err != nil {
			log.Printf("⚠️ Unable to connect to PostgreSQL: %v", err)
			log.Println("🔄 Falling back to local JSON file storage (data/todos.json)...")
			jsonStore, err := storage.NewJSONStore("data/todos.json")
			if err != nil {
				log.Fatalf("Fatal: unable to initialize JSON fallback storage: %v", err)
			}
			store = jsonStore
			dbType = "JSON File (Fallback)"
		} else {
			store = pgStore
			dbType = "PostgreSQL"
			log.Println("✅ Connected to PostgreSQL successfully! Tables and migrations verified.")
		}
	} else {
		log.Println("ℹ️ No DATABASE_URL specified. Initializing local JSON storage...")
		jsonStore, err := storage.NewJSONStore("data/todos.json")
		if err != nil {
			log.Fatalf("Fatal: unable to initialize JSON storage: %v", err)
		}
		store = jsonStore
		dbType = "JSON File"
	}

	todoHandler := handlers.NewTodoHandler(store)

	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("GET /api/todos", todoHandler.GetTodos)
	mux.HandleFunc("POST /api/todos", todoHandler.CreateTodo)
	mux.HandleFunc("GET /api/todos/{id}", todoHandler.GetTodoByID)
	mux.HandleFunc("PUT /api/todos/{id}", todoHandler.UpdateTodo)
	mux.HandleFunc("PATCH /api/todos/{id}/toggle", todoHandler.ToggleTodo)
	mux.HandleFunc("DELETE /api/todos/{id}", todoHandler.DeleteTodo)
	mux.HandleFunc("POST /api/todos/clear-completed", todoHandler.ClearCompleted)
	mux.HandleFunc("GET /api/stats", todoHandler.GetStats)

	// Static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("==================================================")
		fmt.Printf("🚀 Todo App server running at http://localhost:%s\n", port)
		fmt.Printf("🗄️ Database: %s\n", dbType)
		fmt.Println("⚡ Press Ctrl+C to stop the server")
		fmt.Println("==================================================")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")

	if closer, ok := store.(io.Closer); ok {
		_ = closer.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
