# 🗂️ Task Management System

> A multi-workspace collaboration API with **hierarchical Role-Based Access Control (RBAC)**, a structured task lifecycle, and secure JWT-based authentication — built with Go, Gin, and PostgreSQL.

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Framework](https://img.shields.io/badge/Framework-Gin-blue)](https://gin-gonic.com/)
[![Database](https://img.shields.io/badge/Database-PostgreSQL-336791?logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![API Version](https://img.shields.io/badge/API-v1-orange)](http://localhost:8080/api/v1)

---

## 📚 Table of Contents

- [Overview](#-overview)
- [Architecture](#-architecture)
- [Tech Stack](#-tech-stack)
- [Project Structure](#-project-structure)
- [Getting Started](#-getting-started)
  - [Prerequisites](#prerequisites)
  - [Environment Variables](#environment-variables)
  - [Running the Server](#running-the-server)
- [Authentication & Security](#-authentication--security)
- [RBAC Model](#-rbac-model)
- [Task Lifecycle](#-task-lifecycle)
- [API Reference](#-api-reference)
  - [Auth](#auth)
  - [Workspaces](#workspaces)
  - [Tasks](#tasks)
  - [User Profile & Invitations](#user-profile--invitations)
- [Request & Response Format](#-request--response-format)
- [cURL Examples](#-curl-examples)
- [Database Schema](#-database-schema)
- [Development](#-development)
- [Deployment](#-deployment)
- [Contributing](#-contributing)

---

## 🌟 Overview

The **Task Management System** is a RESTful API built around the concept of **workspaces** — isolated collaboration environments where users are assigned specific roles and manage tasks through a governed lifecycle.

**Core design principles (from PRD v1.0):**

| Principle | How it is implemented |
|-----------|----------------------|
| **Multi-tenancy** | All data is scoped to workspaces; users can belong to multiple workspaces with different roles in each |
| **RBAC as backbone** | Every meaningful action is gated by a role. Three tiers: `super_admin` → `admin` → `member` |
| **Governed task lifecycle** | Tasks follow a strict state machine: `todo → in_progress → submitted → completed`, with role-gated transitions |
| **Secure sessions** | JWT access tokens + rotating refresh tokens, both issued as `HttpOnly` cookies |
| **Predictable API** | All routes under `/api/v1`, uniform JSON envelope `{ success, message, data }`, standard HTTP codes |

> **Out of scope for v1:** WebSockets, real-time updates, push notifications.

---

## 🏗️ Architecture

```
+------------------------------------------------------------------+
|                         HTTP Client                              |
+------------------------------+-----------------------------------+
                               | HTTP/HTTPS
+------------------------------v-----------------------------------+
|                    Gin Router  (/api/v1)                        |
|                                                                  |
|  +---------------+  +------------------+  +------------------+  |
|  |  Auth Routes  |  | Workspace Routes |  | Task/User Routes |  |
|  +-------+-------+  +--------+---------+  +--------+---------+  |
|          |                   |                      |            |
|          +-------------------+----------------------+            |
|                              |                                   |
|                    +---------v---------+                         |
|                    |  Auth Middleware  |  JWT + session check    |
|                    +---------+---------+                         |
|                              |                                   |
|                    +---------v---------+                         |
|                    |   Rate Limiter    |  token bucket per IP    |
|                    +---------+---------+                         |
|                              |                                   |
|                    +---------v---------+                         |
|                    |     Handlers      |  parse + validate       |
|                    +---------+---------+                         |
|                              |                                   |
|                    +---------v---------+                         |
|                    |     Services      |  business logic + RBAC  |
|                    +---------+---------+                         |
|                              |                                   |
|                    +---------v---------+                         |
|                    |   Repositories    |  SQL via sqlx           |
|                    +---------+---------+                         |
+------------------------------+-----------------------------------+
                               |
               +---------------v---------------+
               |          PostgreSQL            |
               |  users, workspaces,            |
               |  workspace_members,            |
               |  workspace_invitations,        |
               |  tasks, sessions               |
               +--------------------------------+
```

**Request flow:** Client → Router → Auth Middleware → Rate Limiter → Handler → Service (RBAC) → Repository → PostgreSQL

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26 |
| HTTP Framework | Gin v1.12 |
| Database | PostgreSQL (lib/pq + sqlx) |
| Authentication | golang-jwt/jwt v5 |
| Password Hashing | bcrypt (golang.org/x/crypto) |
| Token Encryption | AES-GCM (crypto/aes) |
| Validation | go-playground/validator v10 |
| Config | godotenv |
| UUIDs | google/uuid |

---

## 📁 Project Structure

```
TASK-SYSTEM/
├── cmd/
│   └── api/                   # Application entry point (main.go)
├── internal/
│   ├── config/                # Environment config loader
│   ├── database/              # DB connection + auto-migration
│   ├── handlers/              # HTTP handlers (request/response)
│   │   ├── auth_handler.go
│   │   ├── workspace.handler.go
│   │   ├── task.handler.go
│   │   └── user.handler.go
│   ├── middleware/            # Auth, rate limiting, token extraction
│   │   ├── auth_middleware.go
│   │   ├── rate_limit.go
│   │   └── token_extractor.go
│   ├── models/                # Structs / DTOs for DB and API
│   ├── repository/            # All SQL queries (data access layer)
│   ├── routes/                # Route registration per domain
│   │   ├── auth.routes.go
│   │   ├── workspace.routes.go
│   │   ├── task.routes.go
│   │   └── user.routes.go
│   ├── service/               # Business logic + RBAC enforcement
│   └── utils/                 # JWT helpers, encryption, shared utilities
├── migrations/
│   └── schema.sql             # Auto-applied on startup
├── .env.example               # Environment variable template
├── .air.toml                  # Hot-reload config (Air)
├── go.mod
└── go.sum
```

---

## 🚀 Getting Started

### Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.22+ | Core runtime |
| PostgreSQL | 14+ | Primary database |
| Air (optional) | latest | Hot-reload for development |

### Environment Variables

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | PostgreSQL hostname | `localhost` | Yes |
| `DB_PORT` | PostgreSQL port | `5432` | Yes |
| `DB_USER` | Database username | `postgres` | Yes |
| `DB_PASSWORD` | Database password | — | Yes |
| `DB_NAME` | Database name | `task_system` | Yes |
| `DB_SSLMODE` | SSL mode (`disable`/`require`) | `disable` | Yes |
| `JWT_SECRET` | Secret key for signing JWTs | — | Yes |
| `ENCRYPTION_KEY` | 32-byte hex key for AES-GCM token encryption | — | Yes |

> **ENCRYPTION_KEY** must be exactly **32 bytes** (256-bit). Generate one with:
> ```bash
> openssl rand -hex 32
> ```

### Running the Server

**Standard (Go run):**
```bash
go mod tidy
go run ./cmd/api
```

**With hot-reload (Air):**
```bash
air
```

**Pre-built binary:**
```bash
go build -o task-api ./cmd/api
./task-api
```

The server starts on **`http://localhost:8080`**.

> The schema in `migrations/schema.sql` is **automatically applied** on every startup — no manual migration step needed.

---

## 🔐 Authentication & Security

The system uses a **dual-token session model**:

| Token | Storage | Lifetime | Purpose |
|-------|---------|----------|---------|
| **Access Token** (JWT) | `HttpOnly` cookie | Short-lived | Authenticates each request |
| **Refresh Token** | `HttpOnly` cookie + hashed & AES-GCM encrypted in DB | Long-lived | Issues a new access token |

**Security properties:**

- **Passwords** — bcrypt hashed; never stored in plaintext
- **Refresh tokens** — stored as `hash(AES-GCM(token))` — compromise of the DB does not expose raw tokens
- **Token rotation** — every `/refresh` call issues a new access + refresh token and invalidates the old one
- **Cookies** — `HttpOnly`, `SameSite=Strict`; set `Secure=true` behind HTTPS in production
- **Rate limiting** — token-bucket middleware applied globally to prevent abuse

**Auth flow:**
```
POST /register  =>  user created
POST /signin    =>  access_token cookie + refresh_token cookie set
POST /refresh   =>  old refresh token invalidated; new pair issued
POST /logout    =>  session deleted from DB; cookies cleared
```

---

## 🔒 RBAC Model

Every action is gated by a **workspace-scoped role**. A user can have different roles in different workspaces.

### Role Hierarchy

```
super_admin  >  admin  >  member
```

| Role | Assigned When | Key Powers |
|------|--------------|-----------|
| `super_admin` | Workspace creator; or via ownership transfer | Full control — only role that can delete the workspace or transfer ownership |
| `admin` | Invited with `admin` role, or promoted by `super_admin` | Member management, task approval, workspace editing |
| `member` | Invited with `member` role (default) | Create tasks, submit assigned tasks |

### Invariants (enforced by the system)

- A workspace **always has at least one `super_admin`** — the last one cannot leave or be removed
- `super_admin` cannot be removed; ownership must be **transferred** first
- `admin` cannot remove other `admin`s or `super_admin`s
- Role promotion/demotion is exclusively a `super_admin` action

### Full RBAC Matrix

| Capability | super_admin | admin | member |
|-----------|:-----------:|:-----:|:------:|
| Create workspace | Yes | Yes | Yes |
| Edit workspace details | Yes | Yes | No |
| Delete workspace | Yes | No | No |
| Transfer ownership | Yes | No | No |
| Leave workspace | No [1] | Yes | Yes |
| Invite member | Yes | Yes | No |
| Remove member | Yes | Yes [2] | No |
| Promote / Demote role | Yes | No | No |
| List members | Yes | Yes | Yes |
| Create task | Yes | Yes | Yes |
| View task | Yes | Yes | Yes |
| Edit any task | Yes | Yes | No |
| Edit own / assigned task | Yes | Yes | Yes |
| Delete task | Yes | Yes | No |
| Assign task | Yes | Yes | No |
| Submit task | Yes | Yes | Yes [3] |
| Approve / Complete task | Yes | Yes | No |
| Reopen task | Yes | Yes | No |

> [1] `super_admin` must transfer ownership before leaving
> [2] `admin` cannot remove other `admin`s or the `super_admin`
> [3] Only the assignee can submit their own task

---

## 🔄 Task Lifecycle

```
[Create]
    |
    v
 +------+    assign (admin+)    +-------------+
 | todo | --------------------> | in_progress |
 +------+                       +------+------+
                                       |
                                       | submit (assignee)
                                       v
                                 +-----------+
                                 | submitted |
                                 +-----+-----+
                                       |
                    +------------------+------------------+
                    |                                     |
         complete (admin+)                     reopen (admin+)
                    |                                     |
                    v                                     v
             +-----------+                      +-------------+
             | completed |                      | in_progress |
             +-----------+                      +-------------+
```

| Transition | Action Endpoint | Who |
|-----------|----------------|-----|
| `todo => in_progress` | `POST /task/:id/assign` | `admin` / `super_admin` |
| `in_progress => submitted` | `POST /task/:id/submit` | Assignee only |
| `submitted => completed` | `POST /task/:id/complete` | `admin` / `super_admin` |
| `submitted/completed => in_progress` | `POST /task/:id/reopen` | `admin` / `super_admin` |

---

## 📖 API Reference

**Base URL:** `http://localhost:8080/api/v1`

**Response envelope:** All endpoints return:
```json
{
  "success": true,
  "message": "Operation completed",
  "data": { }
}
```

---

### Auth

| Method | Endpoint | Auth Required | Body Fields | Description |
|--------|----------|:-------------:|-------------|-------------|
| `POST` | `/auth/register` | No | `name`, `email`, `password` | Register a new user |
| `POST` | `/auth/signin` | No | `email`, `password` | Sign in; sets `access_token` + `refresh_token` cookies |
| `POST` | `/auth/refresh` | refresh cookie | — | Rotate tokens (old pair invalidated) |
| `POST` | `/auth/logout` | access cookie | — | Invalidate session and clear cookies |

---

### Workspaces

All workspace routes require a valid `access_token` cookie.

| Method | Endpoint | Min Role | Body / Params | Description |
|--------|----------|----------|---------------|-------------|
| `POST` | `/workspace/create` | Any auth | `name`, `description?` | Create workspace; creator becomes `super_admin` |
| `GET` | `/workspace` | Member+ | — | List all workspaces user belongs to (with role) |
| `GET` | `/workspace/:id` | Member+ | `:id` workspace UUID | Get workspace details |
| `PUT` | `/workspace/:id` | Admin+ | `name?`, `description?` | Update workspace name / description |
| `DELETE` | `/workspace/:id` | Super Admin | — | Delete workspace (cascades to all data) |
| `POST` | `/workspace/invite` | Admin+ | `workspace_id`, `email`, `role` | Invite user by email (`member` or `admin`) |
| `POST` | `/workspace/join` | Invitee | `workspace_id` | Accept a pending invitation |
| `GET` | `/workspace/:id/members` | Member+ | — | List all members and their roles |
| `PUT` | `/workspace/:id/members/:user_id` | Super Admin | `role` | Change a member's role |
| `DELETE` | `/workspace/:id/members/:user_id` | Admin+ [2] | — | Remove a member from the workspace |
| `POST` | `/workspace/:id/leave` | Member / Admin | — | Leave the workspace |
| `POST` | `/workspace/:id/transfer` | Super Admin | `user_id` | Transfer `super_admin` role to another member |

---

### Tasks

All task routes require a valid `access_token` cookie.

| Method | Endpoint | Min Role | Body / Params | Description |
|--------|----------|----------|---------------|-------------|
| `POST` | `/task/create` | Member+ | `workspace_id`, `title`, `description?` | Create a task (state: `todo`) |
| `GET` | `/task/:id` | Member+ | — | Get task details |
| `GET` | `/task/workspace/:workspace_id` | Member+ | query: `status?` | List tasks in workspace (filterable by status) |
| `PUT` | `/task/:id` | Creator / Assignee / Admin+ | `title?`, `description?` | Edit task |
| `DELETE` | `/task/:id` | Creator / Admin+ | — | Delete task |
| `POST` | `/task/:id/assign` | Admin+ | `user_id` | Assign / reassign task to a workspace member |
| `POST` | `/task/:id/submit` | Assignee | — | Mark task as `submitted` |

---

### User Profile & Invitations

All user routes require a valid `access_token` cookie.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/user/profile` | Get the authenticated user's profile |
| `PUT` | `/user/profile` | Update name, email, or password |
| `DELETE` | `/user/profile` | Delete account (must transfer owned workspaces first) |
| `GET` | `/user/invitations` | List pending workspace invitations |

---

## 🧪 cURL Examples

Save cookies between requests using `-c` (write) and `-b` (read) flags.

### 1. Register
```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Alice\",\"email\":\"alice@example.com\",\"password\":\"Secret123!\"}"
```

### 2. Sign In (saves cookies)
```bash
curl -s -c cookies.txt -X POST http://localhost:8080/api/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"alice@example.com\",\"password\":\"Secret123!\"}"
```

### 3. Create Workspace
```bash
curl -s -b cookies.txt -X POST http://localhost:8080/api/v1/workspace/create \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Acme HQ\",\"description\":\"Main workspace\"}"
```

### 4. Invite a Member
```bash
curl -s -b cookies.txt -X POST http://localhost:8080/api/v1/workspace/invite \
  -H "Content-Type: application/json" \
  -d "{\"workspace_id\":\"<WORKSPACE_UUID>\",\"email\":\"bob@example.com\",\"role\":\"member\"}"
```

### 5. Create a Task
```bash
curl -s -b cookies.txt -X POST http://localhost:8080/api/v1/task/create \
  -H "Content-Type: application/json" \
  -d "{\"workspace_id\":\"<WORKSPACE_UUID>\",\"title\":\"Design landing page\",\"description\":\"Figma mockup first\"}"
```

### 6. Assign Task (as admin)
```bash
curl -s -b cookies.txt -X POST http://localhost:8080/api/v1/task/<TASK_UUID>/assign \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"<USER_UUID>\"}"
```

### 7. Submit Task (as assignee)
```bash
curl -s -b cookies2.txt -X POST http://localhost:8080/api/v1/task/<TASK_UUID>/submit
```

### 8. Refresh Tokens
```bash
curl -s -c cookies.txt -b cookies.txt -X POST http://localhost:8080/api/v1/auth/refresh
```

### 9. Logout
```bash
curl -s -b cookies.txt -X POST http://localhost:8080/api/v1/auth/logout
```

---

## 🗄️ Database Schema

The schema is auto-applied from `migrations/schema.sql` on startup.

```
users
  id (UUID PK) | name | email (UNIQUE) | password (bcrypt) | created_at | updated_at

sessions
  id (UUID PK) | user_id (FK->users, CASCADE) | token (UNIQUE) | expires_at | created_at

workspaces
  id (UUID PK) | name | description | user_id (FK->users, creator) | created_at | updated_at

workspace_members
  id (UUID PK) | workspace_id (FK->workspaces) | user_id (FK->users) | role | created_at
  UNIQUE(workspace_id, user_id)

workspace_invitations
  id (UUID PK) | workspace_id (FK->workspaces) | email | invited_by (FK->users)
  role | status (pending/accepted) | created_at | updated_at
  UNIQUE(workspace_id, email)

tasks
  id (UUID PK) | workspace_id (FK->workspaces) | title | description
  assigned_to (FK->users, nullable SET NULL) | created_by (FK->users)
  status (todo/in_progress/submitted/completed) | created_at | updated_at
```

**Indexes:**
- `idx_workspace_members_user` on `workspace_members(user_id)`
- `idx_tasks_workspace` on `tasks(workspace_id)`
- `idx_tasks_assigned_to` on `tasks(assigned_to)`

All foreign keys use `ON DELETE CASCADE` (or `SET NULL` for nullable references), ensuring referential integrity without orphaned records.

---

## 🔧 Development

### Hot Reload
```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Start with hot-reload
air
```

Air is configured via `.air.toml` in the project root.

### Run Tests
```bash
go test ./...
```

### Code Formatting
```bash
gofmt -w .
```

### Linting
```bash
go vet ./...
```

---

## 📦 Deployment

### Build Binary
```bash
go build -o task-api ./cmd/api
```

### Docker Example

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o task-api ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/task-api .
COPY migrations/ migrations/
EXPOSE 8080
CMD ["./task-api"]
```

### Production Checklist

- [ ] Set all environment variables (especially `JWT_SECRET` and `ENCRYPTION_KEY`)
- [ ] Use `DB_SSLMODE=require` for the database connection
- [ ] Place behind a reverse proxy (NGINX / Caddy) for TLS termination
- [ ] Ensure cookies are served over HTTPS so the `Secure` flag takes effect
- [ ] Use a process manager (systemd, Docker, PM2) for automatic restarts
- [ ] Set up PostgreSQL connection pooling (PgBouncer) for high traffic
- [ ] Monitor with structured logging or a tool like Prometheus + Grafana

---

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork** the repository and create a feature branch: `git checkout -b feat/your-feature`
2. **Write clean Go** — follow standard Go conventions and run `gofmt` before committing
3. **Add tests** for new business logic in the `service/` layer
4. **Commit messages** — use conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`
5. **Open a Pull Request** with a clear description of what changed and why

### Reporting Issues

Open a GitHub Issue with:
- Steps to reproduce
- Expected vs actual behaviour
- Go version and OS

---

Built with Go · Gin · PostgreSQL
