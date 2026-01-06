package tables

import "github.com/causon-mikolorenz/unified-access-backend/database/migrations"

var ScopesMigration = migrations.TableMigration{
	TableName: "scopes",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-scopes-table",
			SQL: `
			CREATE TABLE IF NOT EXISTS scopes (
				id INT AUTO_INCREMENT PRIMARY KEY,
				scope_name VARCHAR(50) NOT NULL UNIQUE,
				description VARCHAR(255)
			);`,
		},
	},
}
