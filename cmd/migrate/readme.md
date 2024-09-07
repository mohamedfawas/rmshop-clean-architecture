This updated version:

Keeps your existing database connection and migration source setup.
Adds command-line flags for specifying the migration direction and number of steps.
Implements the logic to run migrations up or down based on the provided flags.
Checks and displays the migration version before and after running migrations.

To use this updated migration tool:

To run all "up" migrations:
`go run cmd/migrate/main.go -direction=up`

To run a specific number of "up" migrations:
`go run cmd/migrate/main.go -direction=up -steps=2`

To revert all migrations:
`go run cmd/migrate/main.go -direction=down`

To revert a specific number of migrations:
`go run cmd/migrate/main.go -direction=down -steps=1`

To just check the current migration version without running any migrations:
`go run cmd/migrate/main.go`
