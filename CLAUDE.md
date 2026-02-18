# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go backend + Elm frontend web application for UI automation testing with test procedure versioning, test run management, and asset storage. The system tracks manual and automated test execution with full audit trails.

## Development Commands

### Backend (Go)
```bash
# Build the backend binary
make build

# Run the backend server (builds first)
make run

# Run all tests with race detection
make test

# Apply database migrations
make migrate-up

# Rollback last migration
make migrate-down

# Install/update dependencies
make install-deps
```

### Frontend (Elm)
```bash
cd frontend

# Build the Elm application
elm make src/App.elm --output=elm.js

# Development with live reload (requires elm-live)
elm-live src/App.elm --open -- --output=elm.js

# Production build with optimizations
make build-prod

# Serve the built application
make serve  # Uses Python HTTP server on port 8000
```

### Docker Compose (Full Stack)

```bash
# First-time setup: Build Elm locally
cd frontend
elm make src/App.elm --output=elm.js
cd ..

# Start entire stack
docker compose up

# Or use Make target (auto-builds elm.js if missing)
make docker-dev

# After making changes to Elm source:
cd frontend
elm make src/App.elm --output=elm.js
# Refresh browser - changes are immediately visible
```

The docker-compose setup mounts locally-built `elm.js` into the frontend container. This allows instant feedback without rebuilding the container.

**Architecture:**
- Backend runs on internal port 8080
- Frontend (Nginx) exposes port 8080 and proxies to backend
- Access application at http://localhost:8080

**For standalone container builds** (without Docker Compose):
```bash
cd frontend
docker build -t ui-automation-frontend .
docker run -p 8080:80 ui-automation-frontend
```

The Dockerfile still builds Elm from source for self-contained deployments.

### Integration Testing
```bash
# Run full integration test suite (requires running backend)
./test_integration.sh
```

## Architecture Patterns

### Backend Structure

The backend follows **domain-driven design** with these consistent patterns across all domains:

**Package Structure** (user/, project/, testprocedure/, testrun/):
- `model.go` - Domain entity with JSON tags
- `store.go` - Store interface (repository pattern)
- `mysql.go` - MySQL implementation of Store
- `setters.go` - Optional field updates via setter pattern
- `*_test.go` - Table-driven tests

**Key Principles**:
- **Interface-based design**: All persistence uses Store interfaces for testability
- **Repository pattern**: Each domain has its own Store interface
- **Setter pattern**: Updates use `Set*()` functions to modify only specified fields
- **Dependency injection**: Stores constructed with `New*Store(db *sql.DB)` functions
- **Context propagation**: All Store methods accept `context.Context` as first parameter
- **Soft deletes**: Entities have `deleted_at` timestamp, not hard deleted

### Frontend Architecture

**Elm Application** using The Elm Architecture (TEA):
- `App.elm` - Main entry point with routing and page composition
- `Types.elm` - All domain types with JSON encoders/decoders
- `API.elm` - HTTP client for backend communication
- `Pages/*.elm` - Individual pages with Model/Msg/update/view

**Routing**: URL-based navigation
- `/` → Login page
- `/projects` → Projects list
- `/projects/{id}/procedures` → Test procedures
- `/procedures/{id}/runs` → Test runs

**API Communication**: Frontend expects backend at `http://localhost:8080/api/v1`

**Frontend Change Policy**: Avoid changes to `frontend/index.html` unless absolutely necessary. All UI work should be done in Elm modules.

### Session Management

Cookie-based sessions managed by `session/` package:
- Session data stored in-memory (no database persistence)
- Automatic cleanup of expired sessions
- Configuration via `config.yaml` (cookie_name, cookie_secret, duration, secure flag)

### Storage Abstraction

The `storage/` package provides a unified interface for file storage:
- Currently: Local filesystem implementation
- Future: S3, GCS support planned
- Assets stored at `./uploads/test-runs/{run_id}/{asset_type}/{filename}`

## Database Schema Key Relationships

```
users (id)
  ↓ owner_id
projects (id)
  ↓ project_id
test_procedures (id) → supports versioning
  ↓ test_procedure_id
test_runs (id)
  ↓ test_run_id
test_run_assets (id)
```

### Test Procedure Versioning

**Critical concept**: Test procedures use **explicit versioning** (not automatic):

- **In-place updates** (`PUT /procedures/{id}`): Modifies procedure without creating new version
  - Use for iterative development, typo fixes
  - Doesn't affect existing test runs (they capture procedure state at creation)

- **Explicit version creation** (`POST /procedures/{id}/versions`): Creates immutable copy
  - Creates new row with incremented `version` number
  - Original procedure marked `is_latest=false`, new version `is_latest=true`
  - Both versions linked via `parent_id` (version 2+ points to version 1)
  - API lists only show latest version unless explicitly requesting history

**Database columns for versioning**:
- `version` - Integer version number (1, 2, 3, ...)
- `is_latest` - Boolean flag indicating current version
- `parent_id` - Self-referential foreign key to version 1 (nullable)

## Configuration

The application loads config from `config.yaml` with environment variable overrides.

**Environment variable pattern**: `SECTION_KEY` (uppercase with underscores)
- Example: `database.host` → `DATABASE_HOST`
- Example: `session.cookie_secret` → `SESSION_COOKIE_SECRET`

**Critical settings**:
- `session.cookie_secret` - Must be min 32 chars in production
- `session.secure` - Set to `true` for HTTPS (production)
- `storage.base_dir` - Base path for uploaded assets

## Handler/Route Structure

Handlers in `cmd/backend/handlers/`:
- Follow pattern: `{domain}_handlers.go` (e.g., `user_handlers.go`)
- Each handler receives injected Store interfaces via constructor
- Authentication middleware checks session cookies
- Owner-based authorization for projects and procedures

**Handler dependencies**: All handlers depend on Store interfaces, not concrete implementations. Tests can inject mock stores.

## Authorization Requirements

**Every handler that operates on a project-owned resource MUST verify ownership before doing any work.** This is a hard requirement — missing it is a security vulnerability.

### Two middleware layers

1. **`AuthMiddleware`** — applied to all `/api/v1` routes. Validates the session cookie and puts `UserID` into the request context. All protected routes have this.

2. **`ProjectAuthorizationMiddleware`** — applied only to the `projectRouter` subrouter (`/api/v1/projects/{id}`). Checks that the authenticated user owns that project. **This only covers routes registered on `projectRouter`, not on `apiRouter`.**

### Routes registered directly on `apiRouter` do NOT get project auth automatically

Routes like `/procedures/{id}/...` or `/projects/{project_id}/procedures/...` are registered on `apiRouter`, which only carries `AuthMiddleware`. They must enforce ownership themselves inside the handler.

### Pattern: handler-level ownership check for procedure routes

When a handler operates on a test procedure by ID (and is not under `projectRouter`), call `checkProcedureOwnership` before doing any real work:

```go
func (h *TestProcedureHandler) MyHandler(w http.ResponseWriter, r *http.Request) {
    id, ok := parseUUIDOrRespond(w, r, "id", "test procedure")
    if !ok {
        return
    }

    // REQUIRED: verify the caller owns the project this procedure belongs to
    if !h.checkProcedureOwnership(w, r, id) {
        return
    }

    // ... rest of handler
}
```

`checkProcedureOwnership` (`cmd/backend/handlers/testprocedure.go`):
1. Gets `UserID` from context (returns 401 if missing)
2. Fetches the procedure to get its `ProjectID` (returns 404 if not found)
3. Fetches the project to get its `OwnerID` (returns 404 if not found)
4. Returns 403 if `OwnerID != UserID`

`TestProcedureHandler` must have `projectStore project.Store` injected (see constructor) for this to work.

### Checklist when adding a new route

- [ ] Is the route on `projectRouter`? → ownership is handled by middleware, no extra work needed.
- [ ] Is the route on `apiRouter` and operates on a procedure? → call `checkProcedureOwnership`.
- [ ] Is the route on `apiRouter` and operates on a different resource (e.g. test run)? → add an equivalent ownership check tracing back to the owning project.
- [ ] Never register a state-mutating route (POST/PUT/DELETE) on `apiRouter` without an explicit ownership check in the handler.

## Adding New Features

### New Domain Entity
1. Create `{domain}/` package with model.go, store.go, mysql.go, setters.go
2. Add migration files: `database/migrations/000xxx_create_{domain}_table.{up,down}.sql`
3. Run `make migrate-up`
4. Create handlers in `cmd/backend/handlers/{domain}_handlers.go`
5. Register routes in `cmd/backend/serve.go`
6. Add Elm types to `frontend/src/Types.elm`
7. Add API functions to `frontend/src/API.elm`
8. Create page module in `frontend/src/Pages/{Domain}.elm`

### Database Migration
```bash
# Create new migration (manual, no tool)
# Files must follow pattern: 000xxx_{description}.{up,down}.sql
# Place in database/migrations/

# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## Testing Strategy

**Backend**: Table-driven tests using `testutil.SetupTestDB()` for isolated database tests
**Frontend**: Elm's type system prevents many runtime errors; manual testing in browser
**Integration**: Shell script (`test_integration.sh`) exercises full API workflow

When writing backend tests:
- Use `{domain}/common_test.go` for shared test setup
- Always use `t.Parallel()` for concurrent test execution
- Create fresh database instance per test with testutil
- Mock external dependencies (storage, sessions) when appropriate

## Common Gotchas

1. **Frontend CORS**: Backend must allow frontend origin. Check CORS middleware in `serve.go`.

2. **Session cookies**: Sessions stored in memory, lost on server restart. Not suitable for multi-instance deployments without session store backend.

3. **JSON field names**: Backend uses snake_case in JSON tags, Elm decoders must match exactly.

4. **Test procedure version references**: Test runs reference specific procedure versions by ID. Don't assume latest version is being tested.

5. **Asset file paths**: Storage package sanitizes filenames to prevent path traversal. Don't bypass this security check.

6. **Soft deletes**: Queries must filter `deleted_at IS NULL` except when specifically requesting deleted records.
