package tables

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var ClientGrantTypesMigration = migrations.TableMigration{
	TableName: "client_grant_types",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-client-grant-types-table",
			SQL: `
			CREATE TABLE IF NOT EXISTS client_grant_types (
				client_id BINARY(16) NOT NULL,
				grant_type ENUM(
					'authorization_code', 
					'refresh_token', 
					'client_credentials'
				) NOT NULL,
				PRIMARY KEY (client_id, grant_type),
				FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
			);`,
		},
	},
}
