# my-finances-tracker

## OpenApi spec setup

Make sure *swaggo* is installed, you can do this by running the following commands:

```bash
go get -u github.com/swaggo/swag
go install github.com/swaggo/swag/cmd/swag@latest
```

To setup the openapi specification run the following command:

```bash
swag init -g cmd/server/main.go -o docs
```

Or simply use the Makefile:

```bash
make swagger
```

This will generate the openapi swagger documentation and store it in the docs folder. Make sure you mount this folder in some way so that your api can serve it.

## Migrations

In development mode migrations are automatically run on startup together with creating the database if it doesn't exist yet.

To install goose run:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

To create a new migration run the following command:
```bash
goose -dir ./migrations/<MIGRATION_TYPE> create <MIGRATION_NAME> sql
```

Or use the Makefile:
```bash
make migrate-create name=<MIGRATION_NAME>
```

## Makefile Commands

This project includes a Makefile for common development tasks. Run `make help` to see all available commands:

- `make build` - Build the application binary
- `make run` - Build and run the application
- `make dev` - Run with hot reload (requires air)
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run linting (fmt, vet, golangci-lint)
- `make swagger` - Generate Swagger documentation
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback last migration
- `make migrate-create name=<name>` - Create new migration
- `make install-tools` - Install all development tools
- `make clean` - Clean build artifacts

### First Time Setup

1. Install development tools:
   ```bash
   make install-tools
   ```

2. Create your `.env` file:
   ```bash
   make env
   ```
   Then edit `.env` with your local configuration.

3. Start development with hot reload:
   ```bash
   make dev
   ```