# UI Automation

A comprehensive web UI automation testing platform with integrated documentation generation capabilities.

## Features

### Authentication
- User authentication system with plain username + password
- Initial version supports basic credential-based access

### Project Management
- Organize test procedures into projects
- Owner-based access control
- Soft delete support for data retention

### Test Procedures
- Create and manage test procedures with JSON-based steps
- **Explicit versioning**: User-controlled version creation
- In-place updates for iterative development
- Version history tracking for audit trails
- Each test run references a specific immutable procedure version

### Test Run Management
- Track test execution with lifecycle management (pending → running → passed/failed/skipped)
- Attach multiple assets (images, videos, documents, binaries) to test runs
- Automatic asset storage in local filesystem (future: S3, GCS support)
- File upload with security controls (100MB limit, path traversal protection)
- Complete audit trail with timestamps

### Automated Test Generation
- Convert manual test procedures to automated tests
- Support for Selenium and Playwright test generation
- LLM-powered test script generation
- Execute and run generated tests directly from the service

### Documentation Generation
- Aggregate collected images from test runs
- Automatically generate comprehensive documentation for web UI usage
- Transform test evidence into user-facing documentation

## Getting Started

### Backend Server

The Go backend provides user management, authentication, and a foundation for the features described above.

#### Prerequisites

- Go 1.21+
- MySQL 8.0+
- Make (optional, for convenience)

#### Quick Start

1. Install dependencies:
```bash
make install-deps
```

2. Create configuration:
```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your database settings
```

3. Start MySQL database:
```bash
docker run -d \
  --name ui-automation-db \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=ui_automation \
  mysql:8
```

4. Run migrations:
```bash
make migrate-up
```

5. Start the server:
```bash
make run
```

The server will be available at `http://localhost:8080`.

#### Running Tests

```bash
make test
```

### API Endpoints

All authenticated endpoints require a session cookie obtained from login.

#### Public Endpoints
- `GET /health` - Health check

#### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login with credentials
- `POST /api/v1/auth/logout` - Logout

#### Users (Authenticated)
- `GET /api/v1/users` - List users (paginated)
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Soft delete user

#### Projects (Authenticated, Owner-Only)
- `GET /api/v1/projects` - List user's projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/{id}` - Get project details
- `PUT /api/v1/projects/{id}` - Update project
- `DELETE /api/v1/projects/{id}` - Soft delete project

#### Test Procedures (Authenticated, Project Owner-Only)
- `GET /api/v1/projects/{project_id}/procedures` - List procedures
- `POST /api/v1/projects/{project_id}/procedures` - Create procedure
- `GET /api/v1/projects/{project_id}/procedures/{id}` - Get procedure
- `PUT /api/v1/projects/{project_id}/procedures/{id}` - Update procedure (in-place)
- `DELETE /api/v1/projects/{project_id}/procedures/{id}` - Delete procedure
- `POST /api/v1/projects/{project_id}/procedures/{id}/versions` - Create new version
- `GET /api/v1/projects/{project_id}/procedures/{id}/versions` - Get version history

#### Test Runs (Authenticated)
- `GET /api/v1/procedures/{procedure_id}/runs` - List runs for procedure
- `POST /api/v1/procedures/{procedure_id}/runs` - Create test run
- `GET /api/v1/runs/{run_id}` - Get run details
- `PUT /api/v1/runs/{run_id}` - Update run notes
- `POST /api/v1/runs/{run_id}/start` - Start test run
- `POST /api/v1/runs/{run_id}/complete` - Complete test run

#### Test Run Assets (Authenticated)
- `POST /api/v1/runs/{run_id}/assets` - Upload asset (multipart/form-data)
- `GET /api/v1/runs/{run_id}/assets` - List assets for run
- `GET /api/v1/runs/{run_id}/assets/{asset_id}` - Download asset
- `DELETE /api/v1/runs/{run_id}/assets/{asset_id}` - Delete asset

See detailed API documentation and curl examples below.

## Architecture

### Backend

The backend follows enterprise patterns for maintainability and extensibility:

- **Interface-based design** for testability
- **Repository pattern** with Store interfaces
- **Setter pattern** for flexible updates
- **Dependency injection** via constructor functions
- **Context propagation** through all layers
- **Cookie-based session** management with automatic cleanup

### Project Structure

```
ui-automation/
├── cmd/backend/              # Application entry point
│   └── handlers/            # HTTP request handlers
├── user/                     # User domain
├── project/                  # Project domain
├── testprocedure/           # Test procedure domain (with versioning)
├── testrun/                 # Test run domain (with assets)
├── storage/                 # Blob storage abstraction
├── session/                 # Session management
├── database/                # Database & migrations
├── logger/                  # Logging abstraction
└── testutil/                # Test utilities
```

### Database Schema

The system uses a fully implemented relational schema:
- **users** - User accounts with authentication
- **projects** - Project organization (owner_id → user.id)
- **test_procedures** - Test steps with versioning (project_id → project.id)
  - Versioning columns: version, is_latest, parent_id
- **test_runs** - Execution history (test_procedure_id → test_procedure.id)
- **test_run_assets** - Asset metadata (test_run_id → test_run.id)

## API Reference

### Authentication Flow

1. **Register**: Create account and receive session cookie
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"email":"test@example.com","username":"testuser","password":"password123"}'
```

2. **Login**: Authenticate and receive session cookie
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"email":"test@example.com","password":"password123"}'
```

3. **Use Protected Endpoints**: Include cookie in requests
```bash
curl http://localhost:8080/api/v1/users -b cookies.txt
```

4. **Logout**: Clear session
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout -b cookies.txt -c cookies.txt
```

### Configuration

Configuration is loaded from `config.yaml` with environment variable overrides.

```yaml
server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 15s
  write_timeout: 15s

database:
  host: localhost
  port: 3306
  user: root
  password: password
  database: ui_automation
  max_open_conns: 25
  max_idle_conns: 5

session:
  cookie_name: session_id
  cookie_secret: change-this-secret-in-production-min-32-chars
  duration: 24h
  secure: false  # Set to true in production (HTTPS)

storage:
  type: local  # "local" (future: "s3", "gcs")
  base_dir: ./uploads

log:
  level: info  # debug, info, warn, error
```

### Detailed API Examples

#### Complete Workflow Example

```bash
# 1. Register and login
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"email":"user@example.com","username":"user","password":"password123"}'

curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"email":"user@example.com","password":"password123"}'

# 2. Create a project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"name":"My Test Project","description":"Testing project"}' | jq

# 3. List projects
curl -X GET "http://localhost:8080/api/v1/projects?limit=10&offset=0" \
  -b cookies.txt | jq

# 4. Create a test procedure
curl -X POST http://localhost:8080/api/v1/projects/1/procedures \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name":"Login Test",
    "description":"Test user login",
    "steps":[
      {"action":"navigate","url":"https://example.com/login"},
      {"action":"type","selector":"#username","value":"testuser"},
      {"action":"type","selector":"#password","value":"pass123"},
      {"action":"click","selector":"#login-btn"}
    ]
  }' | jq

# 5. Create a test run
curl -X POST http://localhost:8080/api/v1/procedures/1/runs \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"notes":"First test execution"}' | jq

# 6. Start the test run
curl -X POST http://localhost:8080/api/v1/runs/1/start \
  -H "Content-Type: application/json" \
  -b cookies.txt | jq

# 7. Upload a screenshot
curl -X POST http://localhost:8080/api/v1/runs/1/assets \
  -b cookies.txt \
  -F "file=@screenshot.png" \
  -F "asset_type=image" \
  -F "description=Login page screenshot" | jq

# 8. Complete the test run
curl -X POST http://localhost:8080/api/v1/runs/1/complete \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"status":"passed","notes":"All steps completed successfully"}' | jq

# 9. List assets for the run
curl -X GET http://localhost:8080/api/v1/runs/1/assets \
  -b cookies.txt | jq

# 10. Download an asset
curl -X GET http://localhost:8080/api/v1/runs/1/assets/1 \
  -b cookies.txt \
  -o downloaded_screenshot.png
```

#### Test Procedure Versioning Example

```bash
# 1. Create initial test procedure
curl -X POST http://localhost:8080/api/v1/projects/1/procedures \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name":"API Test",
    "description":"Version 1",
    "steps":[{"action":"api_call","endpoint":"/users"}]
  }' | jq

# 2. Update test procedure (in-place, doesn't create new version)
curl -X PUT http://localhost:8080/api/v1/projects/1/procedures/1 \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"description":"Version 1 - Updated"}' | jq

# 3. Create a test run with v1
curl -X POST http://localhost:8080/api/v1/procedures/1/runs \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"notes":"Run with v1"}' | jq

# 4. Explicitly create version 2 (immutable copy)
curl -X POST http://localhost:8080/api/v1/projects/1/procedures/1/versions \
  -H "Content-Type: application/json" \
  -b cookies.txt | jq

# 5. Get version history (shows v1 and v2)
curl -X GET http://localhost:8080/api/v1/projects/1/procedures/1/versions \
  -b cookies.txt | jq

# 6. Create test run with v2
curl -X POST http://localhost:8080/api/v1/procedures/2/runs \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"notes":"Run with v2"}' | jq

# Result: Both test runs reference their specific procedure versions
# - Run 1 references procedure ID 1 (v1)
# - Run 2 references procedure ID 2 (v2)
```

#### Asset Management Example

```bash
# Upload different asset types
curl -X POST http://localhost:8080/api/v1/runs/1/assets \
  -b cookies.txt \
  -F "file=@screenshot.png" \
  -F "asset_type=image" \
  -F "description=UI screenshot"

curl -X POST http://localhost:8080/api/v1/runs/1/assets \
  -b cookies.txt \
  -F "file=@recording.mp4" \
  -F "asset_type=video" \
  -F "description=Test execution video"

curl -X POST http://localhost:8080/api/v1/runs/1/assets \
  -b cookies.txt \
  -F "file=@logs.txt" \
  -F "asset_type=document" \
  -F "description=Test logs"

# List all assets
curl -X GET http://localhost:8080/api/v1/runs/1/assets \
  -b cookies.txt | jq

# Download specific asset
curl -X GET http://localhost:8080/api/v1/runs/1/assets/1 \
  -b cookies.txt \
  -o downloaded_file

# Delete asset
curl -X DELETE http://localhost:8080/api/v1/runs/1/assets/1 \
  -b cookies.txt | jq
```

### Versioning Behavior

The system uses **explicit versioning** for test procedures:

- **In-Place Updates** (`PUT /procedures/{id}`): Modifies the procedure without creating a new version
  - Use for iterative development and fixing typos
  - Doesn't affect existing test runs (they reference the snapshot at creation time)

- **Explicit Version Creation** (`POST /procedures/{id}/versions`): Creates an immutable copy
  - Use when you want to preserve history before major changes
  - New version gets incremented version number (v2, v3, etc.)
  - Only the latest version appears in procedure lists (is_latest=true)
  - Old versions remain accessible via version history

**Example Scenario:**
1. Create procedure v1 with 3 steps
2. Run test → references v1 (procedure ID 1)
3. Update v1 description (in-place) → still v1
4. Run test → references updated v1 (procedure ID 1)
5. Create version → v2 created (procedure ID 2)
6. Run test → references v2 (procedure ID 2)
7. View history → shows both v1 and v2 with their test runs

### Asset Upload Requirements

- **Max file size**: 100MB
- **Supported types**: image, video, binary, document
- **Format**: multipart/form-data
- **Required fields**:
  - `file`: The file to upload
  - `asset_type`: One of [image, video, binary, document]
- **Optional fields**:
  - `description`: Asset description
- **Storage**: Files stored in `./uploads/test-runs/{run_id}/{asset_type}/{filename}`
- **Security**: Path traversal protection, filename sanitization

## Development

### Make Commands

```bash
make build          # Build the binary
make run            # Build and run the server
make test           # Run all tests with race detection
make migrate-up     # Apply all pending migrations
make migrate-down   # Rollback last migration
make clean          # Remove build artifacts
make install-deps   # Download and tidy dependencies
```

### Adding New Features

The architecture makes it easy to extend:

1. **New domain models**: Follow the user package pattern (model.go, store.go, setters.go, mysql.go)
2. **New endpoints**: Add handlers in cmd/backend/handlers/
3. **Database changes**: Create new migration files in database/migrations/

## Production Deployment

### Security Checklist

- Set `session.secure: true` (requires HTTPS)
- Generate strong `session.cookie_secret` (min 32 characters)
- Use environment variables for sensitive data
- Enable firewall and restrict database access
- Set up SSL/TLS certificates
- Configure monitoring and alerting

### Building for Production

```bash
go build -o bin/backend \
  -ldflags="-X main.Version=1.0.0 -X main.Commit=$(git rev-parse HEAD) -X main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  cmd/backend/*.go
```

## Contributing

(Coming soon)
