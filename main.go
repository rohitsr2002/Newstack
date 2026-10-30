package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-app/internal/handlers"
	"todo-app/internal/storage"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// Only log API and page requests, keep static asset noise low
		if r.URL.Path == "/" || len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			log.Printf("[%s] %s - %v", r.Method, r.URL.Path, time.Since(start))
		}
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	storagePath := "data/todos.json"
	store, err := storage.NewStore(storagePath)
	if err != nil {
		log.Fatalf("Fatal: unable to initialize storage: %v", err)
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
		fmt.Printf("🚀 Todo App server is running at http://localhost:%s\n", port)
		fmt.Println("📁 Local storage: " + storagePath)
		fmt.Println("⚡ Press Ctrl+C to stop the server")
		fmt.Println("==================================================")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
