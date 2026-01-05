package tables

import "github.com/causon-mikolorenz/unified-access-backend/database/migrations"

var AuthorizationCodesMigration = migrations.TableMigration{
	TableName: "authorization_codes",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-authorization-codes-table",
			SQL: `
			CREATE TABLE IF NOT EXISTS authorization_codes (
				code VARCHAR(255) PRIMARY KEY,
				client_id BINARY(16) NOT NULL,
				user_id BINARY(16) NOT NULL,
				redirect_uri VARCHAR(2048) NOT NULL,
				expires_at TIMESTAMP NOT NULL,
				used_at TIMESTAMP,
				FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				INDEX idx_code_expiry (expires_at)
			);`,
		},
	},
}
