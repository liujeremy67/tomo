Tomo MVP Backend

Minimal Go (Golang) backend for the **Tomo Pomodoro app** MVP.  
Built using the **Gin** framework and **Postgres** (via GORM or native SQL).  
Simplicity, local development, and realistic foundations before cloud scaling.

---

## Starting Folder Structure
backend/
├── main.go
├── go.mod
│
├── config/
│   └── db.go                  # Database connection setup (Postgres via database/sql)
│
├── models/
│   ├── user.go                # User struct definition
│   └── session.go             # Session struct definition
│
├── controllers/
│   └── sessionController.go   # Controller for handling session endpoints (CRUD, start/end)
│
├── routes/
│   └── routes.go              # Gin route definitions
│
├── sql/
│   └── schema.sql             # Raw SQL schema (User + Session tables)

## Folder Purpose

| Folder / File | Description |
|----------------|-------------|
| **main.go** | Application entry point — starts the Gin server, loads routes, and connects to the database. |
| **go.mod** | Go module definition (dependency tracking). Created via `go mod init`. |
| **config/** | Holds configuration logic such as database setup, environment variables, and constants. |
| **config/db.go** | Handles connecting to Postgres (via GORM or raw SQL). Keeps connection logic separate from app logic. |
| **routes/** | Defines all API endpoints and groups them (e.g., `/sessions`, `/users`). Connects URLs to controllers. |
| **routes/routes.go** | Sets up route mappings like `GET /sessions` → `sessionController.GetSessions`. |
| **controllers/** | Contains the logic for handling incoming requests and responses. Business logic lives here. |
| **controllers/sessionController.go** | Handles session-related operations (start, end, get history). |
| **models/** | Contains Go struct definitions mapping to DB tables (User, Session). |