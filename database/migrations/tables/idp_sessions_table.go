package tables

import "github.com/causon-mikolorenz/unified-access-backend/database/migrations"

var IdpSessionsMigration = migrations.TableMigration{
	TableName: "idp_sessions",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-idp-sessions-table",
			SQL: `
			CREATE TABLE IF NOT EXISTS idp_sessions (
				session_id VARCHAR(255) PRIMARY KEY,
				user_id BINARY(16) NOT NULL,
				ip_address VARCHAR(45) NOT NULL,
				user_agent VARCHAR(512) NOT NULL,
				created_at TIMESTAMP DEFAULT NOW(),
				expires_at TIMESTAMP NOT NULL,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				INDEX idx_session_expiry (expires_at)
			);`,
		},
	},
}
