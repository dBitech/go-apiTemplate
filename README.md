# Go API Template

A production-ready Go API template following best practices and dBi Technologies API guidelines.

## Features

- Clean architecture pattern
- Structured logging with Zap
- Configuration via environment variables, YAML, and command-line flags
- Prometheus metrics for monitoring
- OpenTelemetry for distributed tracing
- Health check endpoints
- Graceful shutdown
- Containerization with Docker
- CI/CD with GitHub Actions
- Automatic dependency updates with Dependabot

## Getting Started

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for running containerized services)
- Make (optional, for using Makefile commands)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/dBiTech/go-apiTemplate.git
cd go-apiTemplate
```

2. Install dependencies:

```bash
go mod download
```

3. Build the application:

```bash
make build
```

### Running the Application

#### Local Development

```bash
make run
```

#### Docker Compose

Run the application with its supporting services (Postgres, Prometheus, Grafana, Jaeger):

```bash
docker-compose up -d
```

### Configuration

Configuration is loaded from multiple sources in the following order (each overrides the previous):

1. Default values
2. Configuration file (config.yaml)
3. Environment variables
4. Command-line flags

Environment variables are prefixed with `APP_` and use underscore notation:

```
APP_SERVER_HOST=0.0.0.0
APP_SERVER_PORT=8080
APP_LOGGING_LEVEL=debug
```

### API Endpoints

| Endpoint               | Method | Description             | Authentication |
|------------------------|--------|-------------------------|---------------|
| /health                | GET    | Health check            | None          |
| /health/liveness       | GET    | Liveness probe          | None          |
| /health/readiness      | GET    | Readiness probe         | None          |
| /metrics               | GET    | Prometheus metrics      | None          |
| /swagger               | GET    | Swagger UI              | None          |
| /api/v1/hello          | GET    | Hello world endpoint    | None          |
| /api/v1/examples       | GET    | List examples           | None          |
| /api/v1/examples       | POST   | Create example          | None          |
| /api/v1/examples/{id}  | GET    | Get example by ID       | None          |
| /api/v1/examples/{id}  | PUT    | Update example by ID    | None          |
| /api/v1/examples/{id}  | DELETE | Delete example by ID    | None          |
| /api/v1/protected/jwt  | GET    | JWT Protected resources | JWT           |
| /api/v1/protected/oauth2 | GET  | OAuth2 Protected resources | OAuth2     |
| /api/v1/me             | GET    | User profile with JWT   | JWT           |
| /api/v1/me/oauth2      | GET    | User profile with OAuth2| OAuth2        |

## Development

### API Documentation

This API template includes Swagger/OpenAPI integration for self-documenting APIs:

- **Swagger UI**: Browse the interactive API documentation at `/swagger/index.html` when the server is running.
- **OpenAPI Specification**: JSON format available at `/swagger/doc.json`.
- **Annotations**: API endpoints are documented using Swag annotations. See example in `handlers.go`.

To update the Swagger documentation after changing annotations:

```bash
go run github.com/swaggo/swag/cmd/swag init -g cmd/api/main.go -o docs
```

### Authentication

This API template includes two types of authentication:

#### JWT Authentication

JWT (JSON Web Token) authentication is implemented for securing API endpoints. To use JWT authentication:

1. Include a bearer token in your request:

   ```bash
   Authorization: Bearer <your-jwt-token>
   ```

2. The token must be signed with the configured secret key and include required scopes.

3. For development, you can obtain a token by setting up a client that calls the auth methods directly.

#### OAuth2 Authentication

OAuth2 authentication is also supported for securing API endpoints. The flow is as follows:

1. Configure the OAuth2 provider details in the config.yaml file.
2. Direct users to the authorization URL to obtain an authorization code.
3. Exchange the authorization code for an access token.
4. Include the access token in requests:

   ```bash
   Authorization: Bearer <oauth2-access-token>
   ```

Protected endpoints will verify the token with the OAuth2 provider and check required scopes.

### Project Structure

```text
├── cmd                  # Application entry points
│   └── api              # API server entry point
├── internal             # Private application code
│   ├── api              # API server implementation
│   ├── config           # Configuration handling
│   ├── handlers         # HTTP handlers
│   ├── middleware       # HTTP middleware
│   ├── models           # Data models
│   ├── repository       # Data access layer
│   └── service          # Business logic layer
├── pkg                  # Public libraries
│   ├── health           # Health check utilities
│   ├── logger           # Logging utilities
│   ├── metrics          # Metrics utilities
│   └── telemetry        # Telemetry utilities
├── test                 # Test utilities and integration tests
├── deploy               # Deployment configurations
│   ├── prometheus       # Prometheus configuration
│   ├── grafana          # Grafana configuration
│   └── otel             # OpenTelemetry configuration
└── .github              # GitHub workflows and configurations
    └── workflows        # GitHub Actions workflows
```

### Running Tests

```bash
make test
```

### Running Linters

```bash
make lint
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- [Chi Router](https://github.com/go-chi/chi)
- [Zap Logger](https://github.com/uber-go/zap)
- [Viper](https://github.com/spf13/viper)
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
