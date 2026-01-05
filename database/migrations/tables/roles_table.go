package tables

import "github.com/causon-mikolorenz/unified-access-backend/database/migrations"

var RolesMigration = migrations.TableMigration{
	TableName: "roles",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-roles-table",
			SQL: `CREATE TABLE IF NOT EXISTS roles (
				id INT AUTO_INCREMENT PRIMARY KEY,
				role_name VARCHAR(50) NOT NULL UNIQUE,
				description VARCHAR(255),
				created_at TIMESTAMP DEFAULT NOW(),
				updated_at TIMESTAMP DEFAULT NOW() ON UPDATE NOW(),
				deleted_at TIMESTAMP NULL
			);`,
		},
	},
}
