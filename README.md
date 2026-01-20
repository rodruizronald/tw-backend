# Job Board Application

A job board application that connects companies with job seekers, focusing on technology skills. The application provides RESTful APIs for managing job postings, companies, and technology skills with comprehensive documentation.

## Project Structure

```
.
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── models/            # Data models
│   └── repository/        # Database access layer
├── migrations/             # Database migration files
├── schema/                # Database schema definition
├── docs/                  # Generated Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── job/                   # Job-related handlers and logic
├── main.go                # Application entry point
└── go.mod                 # Go module definition
```

## Prerequisites & Dependencies

### System Requirements
- Go 1.24+
- PostgreSQL 12+

### Required Tools
- [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations
- [swaggo/swag](https://github.com/swaggo/swag) for API documentation generation

### Go Dependencies
- `github.com/gin-gonic/gin` - Web framework
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/swaggo/gin-swagger` - Swagger middleware
- `github.com/swaggo/files` - Swagger static files
- `golang-migrate/migrate` - Database migrations

## Installation & Setup

### Step 1: Install Migration Tools

**macOS:**
```bash
brew install golang-migrate
```

**Linux:**
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate
```

**Windows (using scoop):**
```bash
scoop install migrate
```

### Step 2: Install Go Dependencies

```bash
# Install Swagger CLI tool
go install github.com/swaggo/swag/cmd/swag@latest

# Install project dependencies
go get github.com/swaggo/swag@latest
go get github.com/gin-gonic/gin
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

### Step 3: Database Setup

**Create PostgreSQL database:**
```bash
createdb jobs_in_tech
```

**Run migrations:**
```bash
# Apply all migrations (migrate up)
migrate -path migrations -database "postgres://username:password@localhost:5432/jobs_in_tech?sslmode=disable" up

# Apply specific number of migrations
migrate -path migrations -database "postgres://username:password@localhost:5432/jobs_in_tech?sslmode=disable" up 2

# Rollback all migrations (migrate down)
migrate -path migrations -database "postgres://username:password@localhost:5432/jobs_in_tech?sslmode=disable" down

# Rollback specific number of migrations
migrate -path migrations -database "postgres://username:password@localhost:5432/jobs_in_tech?sslmode=disable" down 1

# Check current migration version
migrate -path migrations -database "postgres://username:password@localhost:5432/jobs_in_tech?sslmode=disable" version

# Force migration to specific version (use with caution)
migrate -path migrations -database "postgres://username:password@localhost:5432/jobs_in_tech?sslmode=disable" force 001
```

*Replace `username` and `password` with your PostgreSQL credentials.*

### Step 4: Generate Swagger Documentation

**Generate API documentation:**
```bash
swag init
```

This command will:
- Scan your code for swagger comments
- Generate the `docs/` folder with documentation files
- Create interactive API documentation

### Step 5: Run the Application

**Set environment variables:**
```bash
export DATABASE_URL="postgres://username:password@localhost:5432/jobs_in_tech?sslmode=disable"
export PORT=8080
```

**Start the server:**
```bash
go run main.go
```

The application will be available at:
- **API**: `http://localhost:8080/api/v1`
- **Swagger UI**: `http://localhost:8080/swagger/index.html`

## Database Models

The application uses the following data models:

- **Company**: Represents companies that post 
- **Job**: Represents job postings with details like title, description, requirements
- **Technology**: Represents technology skills (programming languages, frameworks, tools)
- **TechnologyAlias**: Alternative names for technologies (e.g., "JS" for "JavaScript")
- **JobTechnology**: Association between  and required technologies

## API Documentation

### Accessing Swagger UI

Once the application is running, visit `http://localhost:8080/swagger/index.html` to access the interactive API documentation where you can:

- View all available endpoints
- Test API calls directly from the browser
- See request/response schemas
- Understand authentication requirements

### API Endpoints Overview

The API provides endpoints for:
- **Companies**: Create, read, update, and delete company profiles
- ****: Manage job postings with full CRUD operations
- **Technologies**: Handle technology skills and their aliases
- **Job-Technology Relations**: Associate  with required technologies

## Development Workflow

### Adding New Database Migrations

Create a new migration file:
```bash
migrate create -ext sql -dir migrations -seq migration_name
```

This creates two files:
- `migrations/{timestamp}_migration_name.up.sql` - For applying changes
- `migrations/{timestamp}_migration_name.down.sql` - For rolling back changes

### Migration Management

**Apply migrations:**
```bash
# Apply all pending migrations
migrate -path migrations -database "your-database-url" up

# Apply only the next N migrations
migrate -path migrations -database "your-database-url" up 2

# Migrate to specific version
migrate -path migrations -database "your-database-url" goto 3
```

**Rollback migrations:**
```bash
# Rollback all migrations
migrate -path migrations -database "your-database-url" down

# Rollback the last N migrations
migrate -path migrations -database "your-database-url" down 1

# Rollback to specific version
migrate -path migrations -database "your-database-url" goto 1
```

**Check migration status:**
```bash
# Show current migration version
migrate -path migrations -database "your-database-url" version

# Show migration history (if supported by database)
migrate -path migrations -database "your-database-url" history
```

**Emergency migration fixes:**
```bash
# Force migration to specific version (use with extreme caution)
migrate -path migrations -database "your-database-url" force 001

# Drop database and recreate (development only)
migrate -path migrations -database "your-database-url" drop
```

### Migration Best Practices

1. **Always test migrations** on a copy of production data
2. **Write rollback scripts** for every migration
3. **Use transactions** where possible in your SQL files
4. **Backup your database** before running migrations in production
5. **Review migration order** - migrations run in numerical order

### Updating API Documentation

After modifying API endpoints or adding swagger comments:

1. Update swagger annotations in your handler functions
2. Regenerate documentation: `swag init`
3. Restart the application to see changes

### Common Development Tasks

**Environment-specific Swagger:**
```bash
# Disable swagger in production
export GIN_MODE=release
```

**Database Connection Testing:**
```bash
# Test database connectivity and show current version
migrate -path migrations -database "your-database-url" version

# Validate migration files
migrate -path migrations -database "your-database-url" validate
```

**Migration Troubleshooting:**
```bash
# If migrations are stuck, check dirty state
migrate -path migrations -database "your-database-url" version

# Fix dirty migration state (replace X with version number)
migrate -path migrations -database "your-database-url" force X

# Start fresh (development only - destroys all data)
migrate -path migrations -database "your-database-url" drop
migrate -path migrations -database "your-database-url" up
```

**Regenerate Documentation:**
```bash
# After updating swagger comments
swag init
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Gin framework mode | `debug` |

## Getting Started

1. Clone the repository
2. Follow the installation steps above
3. Set up your database and run migrations
4. Generate API documentation
5. Start the server and visit the Swagger UI
6. Begin developing your job board features

For questions or contributions, please refer to the project documentation or open an issue in the repository.
