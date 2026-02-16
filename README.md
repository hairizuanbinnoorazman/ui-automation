# UI Automation

A comprehensive web UI automation testing platform with integrated documentation generation capabilities.

## Features

### Authentication
- User authentication system with plain username + password
- Initial version supports basic credential-based access

### Test Procedures
- Create and manage multiple test procedures for UI testing
- Current focus: Web automation testing
- Aggregate test procedures together as projects for better organization

### Test Documentation
- Upload notes and images for each test run
- Provides proof and documentation of web automation execution
- Track test execution history with visual evidence

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

#### Public Endpoints
- `GET /health` - Health check
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login with credentials
- `POST /api/v1/auth/logout` - Logout

#### Protected Endpoints (Require Authentication)
- `GET /api/v1/users` - List users (with pagination)
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Soft delete user

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
├── user/                     # User domain
├── session/                  # Session management
├── database/                 # Database & migrations
├── logger/                   # Logging abstraction
└── testutil/                 # Test utilities
```

### Database Schema

The system is built with extensibility in mind. Current tables:
- **users** - User accounts with authentication
- **projects** - Test procedure organization (ready for future use)
- **test_procedures** - Test steps and metadata (ready for future use)
- **test_runs** - Execution history (ready for future use)

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

log:
  level: info  # debug, info, warn, error
```

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
