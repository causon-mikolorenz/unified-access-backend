package database

import (
	"log"

	"github.com/causon-mikolorenz/unified-access-backend/database/migrations"
	"github.com/causon-mikolorenz/unified-access-backend/database/migrations/procedures"
	"github.com/causon-mikolorenz/unified-access-backend/database/migrations/tables"
	"github.com/jmoiron/sqlx"
)

func executeTableMigration(db *sqlx.DB, migration migrations.TableMigration) {
	for _, step := range migration.Steps {
		// 1. Check if this specific step ID has already been executed
		var count int
		query := "SELECT COUNT(*) FROM migration_history WHERE migration_id = ?"
		err := db.Get(&count, query, step.ID)
		if err != nil {
			// If the history table doesn't exist yet, we'll need to handle that
			// (usually by creating it at the very start of RunMigrations)
			log.Fatalf("Failed to check migration history: %v", err)
		}

		// 2. If count is 0, the step has never run
		if count == 0 {
			log.Printf("[MIGRATE] Applying step: %s", step.ID)

			// Execute the SQL (Create table, Alter table, etc.)
			db.MustExec(step.SQL)

			// 3. Record that this step is now finished
			db.MustExec("INSERT INTO migration_history (migration_id) VALUES (?)", step.ID)
		} else {
			log.Printf("[MIGRATE] Skipping completed step: %s", step.ID)
		}
	}
}

func RunAllMigrations(db *sqlx.DB) {
	log.Println("[MIGRATE] Initializing migration history...")
	db.MustExec(`CREATE TABLE IF NOT EXISTS migration_history (
        id INT AUTO_INCREMENT PRIMARY KEY,
        migration_id VARCHAR(255) NOT NULL UNIQUE,
        applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`)

	parentTables := []migrations.TableMigration{
		tables.UsersMigration,
		tables.RolesMigration,
		tables.ClientsMigration,
		tables.ScopesMigration,
		tables.AuditLogsMigration,
	}

	childTables := []migrations.TableMigration{
		tables.UserRolesMigration,
		tables.ClientUrlsMigration,
		tables.ClientGrantTypesMigration,
		tables.IdpSessionsMigration,
		tables.RefreshTokensMigration,
		tables.AuthorizationCodesMigration,
	}

	procedurePlan := []migrations.MigrationPart{
		procedures.CreateUserProcedure,
		procedures.RegisterClientProcedure,
		procedures.ArchiveUserProcedure,
		procedures.UpdateUserPasswordProcedure,
		procedures.LogoutUserProcedure,
		procedures.RotateRefreshTokenProcedure,
		procedures.GetUserForAuthProcedure,
	}

	allMigrations := append(parentTables, childTables...)
	for _, m := range allMigrations {
		executeTableMigration(db, m)
	}

	// 2. Deploy/Overwrite Procedures
	for _, proc := range procedurePlan {
		log.Printf("[MIGRATE] Deploying Procedure: %s", proc.Name)
		db.MustExec(proc.SQL)
	}

	log.Println("[MIGRATE] All tables and procedures are synchronized.")
}
