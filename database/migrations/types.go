package migrations

type MigrationStep struct {
	ID  string // Unique ID for this specific change (e.g., "users-create", "users-add-phone")
	SQL string
}

type TableMigration struct {
	TableName string
	Steps     []MigrationStep
}
