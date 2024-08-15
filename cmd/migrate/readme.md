This updated version:

Keeps your existing database connection and migration source setup.
Adds command-line flags for specifying the migration direction and number of steps.
Implements the logic to run migrations up or down based on the provided flags.
Checks and displays the migration version before and after running migrations.

To use this updated migration tool:

To run all "up" migrations:
Copygo run cmd/migrate/main.go -direction=up

To run a specific number of "up" migrations:
Copygo run cmd/migrate/main.go -direction=up -steps=2

To revert all migrations:
Copygo run cmd/migrate/main.go -direction=down

To revert a specific number of migrations:
Copygo run cmd/migrate/main.go -direction=down -steps=1

To just check the current migration version without running any migrations:
Copygo run cmd/migrate/main.go


This implementation maintains your existing setup while adding the functionality to run migrations up or down. It also provides more detailed output about the migration process, including the version before and after migration.