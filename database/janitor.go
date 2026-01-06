package database

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

func StartJanitor(db *sqlx.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)

	// Run in a separate goroutine
	go func() {
		for {
			select {
			case <-ticker.C:
				cleanExpiredRecords(db)
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
