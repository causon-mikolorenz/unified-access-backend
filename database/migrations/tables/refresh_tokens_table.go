package tables

import "github.com/causon-mikolorenz/unified-access-backend/database/migrations"

var RefreshTokensMigration = migrations.TableMigration{
	TableName: "refresh_tokens",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-refresh-tokens-table",
			SQL: `
			CREATE TABLE IF NOT EXISTS refresh_tokens (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				token VARCHAR(255) NOT NULL UNIQUE,
				client_id BINARY(16) NOT NULL,
				user_id BINARY(16) NOT NULL,
				expires_at TIMESTAMP NOT NULL,
				revoked_at TIMESTAMP,
				FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				INDEX idx_token_lookup (token),
				INDEX idx_expiry_lookup (expires_at)
			);`,
		},
	},
}
