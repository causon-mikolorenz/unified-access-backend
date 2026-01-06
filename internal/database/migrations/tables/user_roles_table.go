package tables

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var UserRolesMigration = migrations.TableMigration{
	TableName: "user_roles",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-user-roles-table",
			SQL: `CREATE TABLE IF NOT EXISTS user_roles (
				user_id BINARY(16),
				role_id INT,
				assigned_at TIMESTAMP DEFAULT NOW(),
				PRIMARY KEY (user_id, role_id),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
				INDEX idx_role_lookup (role_id)
			);`,
		},
	},
}
