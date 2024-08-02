- go get github.com/lib/pq
- go get github.com/spf13/viper
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
- 


Project structure explanation

cmd/: Contains the main entry points for the application

api/main.go: Starts the HTTP server for the API
seedadmin/main.go: Seeds the initial admin user in the database


internal/: Core application code, following clean architecture principles

config/: Manages application configuration
delivery/http/: Handles HTTP requests and responses
domain/: Defines core business entities and interfaces
repository/: Manages data persistence and retrieval
usecase/: Implements business logic
server/: Sets up and initializes the server


migrations/: Contains database migration files
pkg/: Shared packages and utilities
scripts/: Utility scripts for development and deployment