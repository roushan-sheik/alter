# GoTickets API

GoTickets is a robust backend service for a ticket booking platform, implemented in Go. It provides comprehensive functionality for user authentication, event lifecycle management, and ticket reservations. The service is built on top of the Echo web framework and utilizes GORM for data persistence.

## Live Environment

- **BASE URL:** https://altar.softvencealpha.com/
- **Swagger Docs:** https://altar.softvencealpha.com/swagger/index.html

## Architecture

This project strictly adheres to a Domain-Driven Design (DDD) layered architecture to ensure a clean separation of concerns, high maintainability, and testability.

The application is modularized within the `internal/` directory:

### Directory Layout

```text
.
├── cmd/                        # Main application entry point
│   └── main.go                 
├── internal/                   # Private application and library code
│   ├── auth/                   # JWT authentication and authorization utilities
│   ├── config/                 # Environment configuration and database initialization
│   ├── domain/                 # Core business domains
│   │   ├── booking/            # Ticket reservation and booking logic
│   │   ├── event/              # Event management logic
│   │   └── user/               # User identity and profile management
│   ├── httpresponse/           # Standardized HTTP error and response handling
│   ├── middlewares/            # HTTP middleware components (e.g., Auth, Logging)
│   └── server/                 # HTTP server configuration and bootstrap
├── go.mod                      # Module dependencies
└── go.sum                      # Module checksums
```

### Domain Module Structure

Each domain within `internal/domain/` (e.g., `booking`, `user`) implements a standardized layered pattern:

1. **Handler (`handler.go`)**: The transport layer. Responsible for request parsing, input validation, and HTTP response formatting. Handlers delegate core processing to the Service layer.
2. **Service (`service.go`)**: The business logic layer. Encapsulates domain-specific rules and orchestrates data flow between Handlers and Repositories.
3. **Repository (`repository.go`)**: The data access layer. Abstracts all database interactions, allowing the Service layer to remain agnostic of the underlying storage implementation.
4. **Entity (`entity.go`)**: Database schema definitions, utilized by GORM for ORM mapping and migrations.
5. **DTO (`dto/`)**: Data Transfer Objects defining strict API contracts for request and response payloads, ensuring internal entities are not directly exposed.
6. **Register (`register.go`)**: Dependency injection and route registration for the specific domain module.

## Technology Stack

- **Language:** Go (Golang)
- **Web Framework:** [Echo v5](https://echo.labstack.com/)
- **ORM:** [GORM](https://gorm.io/)
- **Authentication:** JWT (JSON Web Tokens)

## Configuration

The application utilizes environment variables for configuration. Create a `.env` file in the root directory and copy the content of `.env.example` file into `.env` file.

```bash
cp .env.example .env
```

Fill in the required values in `.env` before running the application.

## Running the Application

### Option 1: Local Development (Recommended for development)

Requires Go and PostgreSQL installed on your machine.

```bash
# Install air for hot reload (one-time setup)
go install github.com/air-verse/air@latest

# Start the development server with hot reload
air
```

The server starts at `http://localhost:5000`.


### Option 2: Docker — Production Build

Builds an optimized, minimal binary image (~15MB). Use this for staging and production deployments.

```bash
docker compose up --build
```

### API Documentation

Once the server is running, visit the interactive Swagger UI:

```
http://localhost:5000/swagger/index.html
```

### Regenerating Swagger Docs

After modifying handler annotations, regenerate the documentation files by running:

```bash
swag init -g cmd/api_info.go --parseDependency --parseInternal
```

*This repository serves as a reference implementation for structuring scalable Go applications.*

## License

This project is licensed under the terms of the MIT license.
