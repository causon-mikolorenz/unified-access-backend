package tables

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var ClientsMigration = migrations.TableMigration{
	TableName: "clients",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-clients-table",
			SQL: `
			CREATE TABLE IF NOT EXISTS clients (
				id BINARY(16) PRIMARY KEY,
				client_name VARCHAR(100) NOT NULL,
				client_secret VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT NOW(),
				updated_at TIMESTAMP DEFAULT NOW() ON UPDATE NOW(),
				deleted_at TIMESTAMP NULL
			);`,
		},
		{
			ID: "add-url-columns",
			SQL: `
				ALTER TABLE clients
				ADD COLUMN base_url VARCHAR(255) NOT NULL,
				ADD COLUMN redirect_uri VARCHAR(255) NOT NULL,
				ADD COLUMN logout_uri VARCHAR(255) NOT NULL;
			`,
		},
		{
			ID: "add-other-identifier-columns",
			SQL: `
				ALTER TABLE clients
				ADD COLUMN abbreviation VARCHAR(10) UNIQUE NOT NULL,
				ADD COLUMN description TEXT,
				ADD COLUMN image_location VARCHAR(255);
			`,
		},
	},
}
