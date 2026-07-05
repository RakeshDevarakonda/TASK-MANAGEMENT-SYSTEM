# Task Management System

## Overview
A **robust** Go‑based task management API built with **Gin**, **SQLX**, and **PostgreSQL**. It supports:
- User authentication (register, signin, refresh, logout)
- Workspaces creation, invitation, membership management, and ownership transfer
- Full CRUD for tasks with assignment and submission flows
- Rate limiting middleware and secure cookie handling

## Quick Start
1. **Clone & configure**
   ```bash
   git clone <repo-url>
   cd TASK-SYSTEM
   cp .env.example .env   # edit DB credentials & secrets
   ```
2. **Run the server**
   ```powershell
   $env:JWT_SECRET='testsecret'; $env:ENCRYPTION_KEY='0123456789abcdef0123456789abcdef';
   $env:DB_HOST='localhost'; $env:DB_PORT='5432'; $env:DB_USER='postgres';
   $env:DB_PASS=''; $env:DB_NAME='tasksystems'; $env:DB_SSLMODE='disable';
   go run ./cmd/api
   ```
3. The API listens on `:8080`. Use the documented routes below or the generated `curl` examples in the repo.

## API Routes (v1)
| Group | Method | Path | Description |
|-------|--------|------|-------------|
| **Auth** | `POST` | `/api/v1/auth/register` | Register a new user |
|  | `POST` | `/api/v1/auth/signin` | Sign‑in and receive JWT cookies |
|  | `POST` | `/api/v1/auth/refresh` | Refresh access token |
|  | `POST` | `/api/v1/auth/logout` | Invalidate session |
| **Workspace** | `POST` | `/api/v1/workspace/create` | Create a workspace |
|  | `POST` | `/api/v1/workspace/:id/invite` | Invite a member |
|  | `POST` | `/api/v1/workspace/:id/join` | Join an invitation |
|  | `GET` | `/api/v1/workspace/:id` | Get workspace details |
|  | `GET` | `/api/v1/workspace` | List user's workspaces |
|  | `PUT` | `/api/v1/workspace/:id` | Update workspace |
|  | `DELETE` | `/api/v1/workspace/:id` | Delete workspace |
|  | `GET` | `/api/v1/workspace/:id/members` | List members |
|  | `PUT` | `/api/v1/workspace/:id/members/:user_id` | Update member role |
|  | `DELETE` | `/api/v1/workspace/:id/members/:user_id` | Kick member |
|  | `POST` | `/api/v1/workspace/:id/leave` | Leave workspace |
|  | `POST` | `/api/v1/workspace/:id/transfer` | Transfer ownership |
| **Task** | `POST` | `/api/v1/task/create` | Create a task |
|  | `GET` | `/api/v1/task/:id` | Get task |
|  | `GET` | `/api/v1/task/workspace/:workspace_id` | List tasks in workspace |
|  | `PUT` | `/api/v1/task/:id` | Update task |
|  | `DELETE` | `/api/v1/task/:id` | Delete task |
|  | `POST` | `/api/v1/task/:id/assign` | Assign task |
|  | `POST` | `/api/v1/task/:id/submit` | Submit task |
| **User** | `GET` | `/api/v1/user/profile` | Get own profile |
|  | `PUT` | `/api/v1/user/profile` | Update profile |
|  | `DELETE` | `/api/v1/user/profile` | Delete account |
|  | `GET` | `/api/v1/user/invitations` | List workspace invitations |

## Development
- **Migrations**: `migrations/schema.sql` – applied automatically on server start.
- **Rate limiting**: Configurable via `internal/middleware/rate_limit.go`.
- **Testing**: Use the provided `curl` scripts or go test suites.

## License
MIT – feel free to fork and extend.
