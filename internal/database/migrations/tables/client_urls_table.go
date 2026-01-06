package tables

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var ClientUrlsMigration = migrations.TableMigration{
	TableName: "client_urls",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-client_urls-table",
			SQL: `
			CREATE TABLE IF NOT EXISTS client_urls (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				client_id BINARY(16),
				redirect_url VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT NOW(),
				FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
				INDEX idx_client_lookup (client_id)
			);`,
		},
	},
}
