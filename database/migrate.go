package database

import (
	"database/sql"
	"fmt"

	"github.com/causon-mikolorenz/unified-access-backend/database/migrations"
)

func RunAllMigrations(db *sql.DB) error {
	allMigrations := append(migrations.Tables, migrations.Procedures...)

	// Start Transaction
	transaction, err := db.Begin()
	// Check if there's an error
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	// Cycle through all migration calls
	for _, migration := range allMigrations {
		fmt.Printf("Executing %s\n", migration.Name)
		if _, err := transaction.Exec(migration.SQL); err != nil {
			return fmt.Errorf("Failed to execute %s: %w", migration.Name, err)
		}
	}

	// Commit afterwards
	return transaction.Commit()
}
