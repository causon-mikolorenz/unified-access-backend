package tables

import "github.com/causon-mikolorenz/unified-access-backend/database/migrations"

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
	},
}
