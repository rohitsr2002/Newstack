# ⚡ Zenith TaskFlow - Go CRUD Todo Application

A full-stack, responsive, and modern Todo CRUD application built with **Go** for the backend REST API and **Vanilla HTML5 / CSS3 / JavaScript** for the UI.

![Go Version](https://img.shields.io/badge/Go-1.27-00ADD8?style=flat&logo=go)
![Architecture](https://img.shields.io/badge/Architecture-REST_API_+_SPA-6366f1)
![UI](https://img.shields.io/badge/Design-Glassmorphism_&_Dark_Mode-a855f7)

---

## ✨ Features

- **Full CRUD Capabilities**:
  - **Create**: 1-click Quick Add bar or detailed modal with Priority, Category, Due Date, and Description.
  - **Read**: Dynamic responsive cards with real-time statistics (Total, In Progress, Completed, Completion Rate %, Overdue alerts).
  - **Update**: Checkbox completion toggle with visual feedback, or full edit modal to adjust any field.
  - **Delete**: Task deletion with confirmation dialog and "Clear Completed" batch cleanup.
- **Search & Filters**:
  - Real-time debounced keyword search (`/` shortcut).
  - Filter tabs: All, Active, Completed.
  - Granular filters by Priority (High, Medium, Low) and Category (Work, Backend, Frontend, Design, Health, Study, General).
  - Dynamic Sorting: Newest First, Oldest First, Due Date, Priority Level, Alphabetical.
- **Modern UI & Aesthetics**:
  - Curated glassmorphism palette with ambient glow accents.
  - Light & Dark mode switch (persisted in `localStorage`).
  - Native `<dialog>` modal with backdrop blur and accessible keyboard controls.
  - Keyboard shortcuts (`N` for new task, `/` for search, `Esc` to close).
  - Toast notification system for instantaneous user feedback.
- **Robust Go Backend**:
  - Built with Go's standard library (`net/http` enhanced routing).
  - Thread-safe storage with read/write mutexes (`sync.RWMutex`).
  - Automatic JSON persistence to `data/todos.json` with pre-seeded sample tasks on first launch.
  - Graceful server shutdown on `SIGINT` / `SIGTERM`.

---

## 🚀 Quick Start

### 1. Run the application

In the project root directory:

```bash
go run .
```

Or build an executable binary:

```bash
go build -o todo-app .
./todo-app
```

### 2. Access the UI

Open your browser and navigate to:
```
http://localhost:8080
```

---

## 📡 REST API Reference

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/todos` | List tasks (supports `?q=`, `?status=`, `?priority=`, `?category=`, `?sort=`) |
| `POST` | `/api/todos` | Create a new task |
| `GET` | `/api/todos/{id}` | Get task details by ID |
| `PUT` | `/api/todos/{id}` | Update task title, description, category, priority, due date, status |
| `PATCH` | `/api/todos/{id}/toggle` | Quick toggle completion status |
| `DELETE` | `/api/todos/{id}` | Delete task by ID |
| `POST` | `/api/todos/clear-completed` | Batch delete all completed tasks |
| `GET` | `/api/stats` | Retrieve metrics summary (total, active, completed, overdue, priority breakdown) |

---

## 📂 Project Structure

```
Newstack/
├── main.go                     # Server initialization, routing & graceful shutdown
├── go.mod                      # Go module declaration
├── internal/
│   ├── models/
│   │   └── todo.go             # Todo domain structs & request types
│   ├── storage/
│   │   └── storage.go          # Thread-safe in-memory + JSON file persistence store
│   └── handlers/
│       └── handlers.go         # REST API HTTP handlers
├── static/
│   ├── index.html              # Modern semantic HTML5 structure & native dialogs
│   ├── css/
│   │   └── style.css           # Glassmorphism design system & micro-animations
│   └── js/
│       └── app.js              # Client state, AJAX fetch calls, keyboard shortcuts & toasts
└── data/
    └── todos.json              # Auto-created persistent storage file
```
