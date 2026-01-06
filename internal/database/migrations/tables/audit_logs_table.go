package tables

import "github.com/causon-mikolorenz/unified-access-backend/database/migrations"

var AuditLogsMigration = migrations.TableMigration{
	TableName: "audit_logs",
	Steps: []migrations.MigrationStep{
		{
			ID: "create-audit-logs-table",
			SQL: `
			CREATE TABLE IF NOT EXISTS audit_logs (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				user_id BINARY(16),
				action VARCHAR(100) NOT NULL,
				timestamp TIMESTAMP DEFAULT NOW(),
				details TEXT,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
				INDEX idx_user_action (user_id, action)
			);`,
		},
	},
}
