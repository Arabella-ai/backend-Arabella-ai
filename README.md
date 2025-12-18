# Arabella Backend

AI Video Generation Platform - Backend API

## Architecture

This project follows Clean Architecture principles:

```
backend/
├── cmd/                    # Application entry points
│   ├── api/                # Main API server
│   └── swagger/            # Swagger documentation generator
├── config/                 # Configuration management
├── internal/
│   ├── domain/             # Business entities and interfaces
│   │   ├── entity/         # Domain entities
│   │   ├── repository/     # Repository interfaces
│   │   └── service/        # Service interfaces
│   ├── infrastructure/      # External implementations
│   │   ├── auth/           # Authentication (JWT, Google OAuth)
│   │   ├── cache/          # Redis cache
│   │   ├── database/       # PostgreSQL connection
│   │   ├── provider/       # AI video generation providers
│   │   ├── queue/          # Job queue (Redis)
│   │   ├── repository/    # Repository implementations
│   │   └── worker/         # Background workers
│   ├── interface/          # External interfaces
│   │   ├── http/           # HTTP handlers and middleware
│   │   └── websocket/      # WebSocket handlers
│   └── usecase/            # Business logic (use cases)
├── database/
│   └── schema/             # Database schema files
├── migrations/              # Database migrations
└── scripts/                 # Utility scripts
```

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Configure environment variables (see `.env.example`):
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Run migrations:
```bash
./scripts/run-migrations.sh
```

4. Build and run:
```bash
go build -o bin/api ./cmd/api
./bin/api
```

## Environment Variables

See `.env.example` for all required environment variables.

**Important**: Never commit `.env` files or secrets to git.

## Database

Database migrations are stored in `migrations/` directory.
Schema files are in `database/schema/`.

## API Documentation

Swagger documentation is available at `/swagger/index.html` in development mode.

## License

MIT

