package tables

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var UsersMigration = migrations.TableMigration{
	TableName: "users",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-users-table",
			SQL: `CREATE TABLE IF NOT EXISTS users (
				id BINARY(16) PRIMARY KEY,
				username VARCHAR(255) NOT NULL UNIQUE,
				first_name VARCHAR(50),
				middle_name VARCHAR(50),
				last_name VARCHAR(50),
				email VARCHAR(100) NOT NULL UNIQUE,
				password_hash VARCHAR(255) NOT NULL,
				status ENUM(
					'active', 
					'inactive', 
					'suspended', 
					'deleted'
				) DEFAULT 'active',
				created_at TIMESTAMP DEFAULT NOW(),
				updated_at TIMESTAMP DEFAULT NOW() ON UPDATE NOW(),
				deleted_at TIMESTAMP NULL
			);`,
		},
	},
}
