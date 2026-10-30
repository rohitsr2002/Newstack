# ⚡ Zenith TaskFlow - Go CRUD Todo Application

A full-stack, responsive, and modern Todo CRUD application built with **Go** for the backend REST API, **PostgreSQL** (with automatic fallback to JSON storage) for persistence, and **Vanilla HTML5 / CSS3 / JavaScript** for the UI.

![Go Version](https://img.shields.io/badge/Go-1.27-00ADD8?style=flat&logo=go)
![Database](https://img.shields.io/badge/Database-PostgreSQL_|_JSON-336791?logo=postgresql)
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
- **Database Persistence**:
  - **PostgreSQL**: Native SQL connection pool (`database/sql` + `github.com/lib/pq`).
  - **Automatic Migrations**: Automatically creates the `todos` table and indexes upon server startup.
  - **Zero-Downtime Fallback**: If PostgreSQL is not configured or offline, it gracefully falls back to local JSON storage (`data/todos.json`).
- **Modern UI & Aesthetics**:
  - Curated glassmorphism palette with ambient glow accents.
  - Light & Dark mode switch (persisted in `localStorage`).
  - Native `<dialog>` modal with backdrop blur and accessible keyboard controls.
  - Keyboard shortcuts (`N` for new task, `/` for search, `Esc` to close).
  - Toast notification system for instantaneous user feedback.
  - Live Database indicator badge in the footer (`🐘 PostgreSQL Active`).

---

## 🗄️ Database Setup & Configuration

Configure your PostgreSQL connection in [.env](file:///.env):

```env
DATABASE_URL="postgresql://rohitrohra@localhost:5432/todos?sslmode=disable"
PORT=8080
```

### PostgreSQL Table Schema (Auto-Created)
```sql
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
```

---

## 🚀 Quick Start

### 1. Run the application

In the project root directory:

```bash
# Restart the server to apply the PostgreSQL connection
go run .
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
| `GET` | `/api/stats` | Retrieve metrics summary (total, active, completed, overdue, priority breakdown, database type) |

---

## 📂 Project Structure

```
Newstack/
├── .env                        # PostgreSQL connection string & server port
├── main.go                     # Server initialization, .env loader, DB connector & graceful shutdown
├── go.mod                      # Go module declaration with github.com/lib/pq
├── internal/
│   ├── models/
│   │   └── todo.go             # Todo domain structs & request types
│   ├── storage/
│   │   ├── interface.go        # Common Store interface
│   │   ├── postgres.go         # PostgreSQL database implementation (auto-migrated)
│   │   └── storage.go          # JSON fallback store
│   └── handlers/
│       └── handlers.go         # REST API HTTP handlers
├── static/
│   ├── index.html              # Modern semantic HTML5 structure & native dialogs
│   ├── css/
│   │   └── style.css           # Glassmorphism design system & micro-animations
│   └── js/
│       └── app.js              # Client state, AJAX fetch calls, keyboard shortcuts & toasts
└── data/
    └── todos.json              # Fallback local storage
```
