package database

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

func StartJanitor(ctx context.Context, db *sqlx.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		log.Println("[Janitor] Started background cleanup task")
		for {
			select {
			case <-ticker.C:
				cleanExpiredRecords(db)
			case <-ctx.Done(): // Listen for the shutdown signal
				log.Println("[Janitor] Shutting down background task...")
				return
			}
		}
	}()
}

func cleanExpiredRecords(db *sqlx.DB) {
	// 1. Delete expired auth codes
	query := "DELETE FROM auth_codes WHERE expires_at < NOW()"
	result, err := db.Exec(query)
	if err != nil {
		log.Printf("[Janitor] Error cleaning auth_codes: %v", err)
		return
	}

	rowsDeleted, _ := result.RowsAffected()

	// 2. Log the cleanup to audit_logs if any rows were removed
	if rowsDeleted > 0 {
		log.Printf("[Janitor] Successfully purged %d expired auth codes.",
			rowsDeleted)

		auditQuery := `INSERT INTO audit_logs (action, details, created_at) 
					   VALUES (?, ?, NOW())`
		_, err = db.Exec(
			auditQuery,
			"DB_CLEANUP",
			"Purged expired authorization codes",
		)
		if err != nil {
			log.Printf("[Janitor] Failed to log cleanup to audit_logs: %v", err)
		}
	}
}
