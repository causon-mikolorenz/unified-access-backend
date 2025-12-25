package migrations

type Migration struct {
	Name string
	SQL  string
}

var Tables = []Migration{
	{
		Name: "Create Users Table",
		SQL: `CREATE TABLE IF NOT EXISTS users (
			id BINARY(16) PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			first_name VARCHAR(50),
			last_name VARCHAR(50),
			email VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			status ENUM(
				'active', 
				'inactive', 
				'suspended', 
				'deleted'
			) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL
		);`,
	},
	{
		Name: "Create Roles Table",
		SQL: `CREATE TABLE IF NOT EXISTS roles (
			id INT AUTO_INCREMENT PRIMARY KEY,
			role_name VARCHAR(50) NOT NULL UNIQUE,
			description VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL
		);`,
	},
	{
		Name: "Create UserRoles Table",
		SQL: `CREATE TABLE IF NOT EXISTS user_roles (
			user_id BINARY(16),
			role_id INT,
			assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, role_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
			INDEX idx_role_lookup (role_id)
		);`,
	},
	{
		Name: "Create AuditLogs Table",
		SQL: `CREATE TABLE IF NOT EXISTS audit_logs (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BINARY(16),
			action VARCHAR(100) NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			details TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
			INDEX idx_user_action (user_id, action)
		);`,
	},
	{
		// Serrvice Providers
		Name: "Create Clients Table",
		SQL: `CREATE TABLE IF NOT EXISTS clients (
			id BINARY(16) PRIMARY KEY,
			client_name VARCHAR(100) NOT NULL,
			client_secret VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL
		);`,
	},
	{
		// Redirect URLs after authentication
		Name: "Create ClientUrls Table",
		SQL: `CREATE TABLE IF NOT EXISTS client_urls (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			client_id BINARY(16),
			redirect_url VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
			INDEX idx_client_lookup (client_id)
		);`,
	},
	{
		// Authorization Codes for Authentication
		Name: "Create AuthorizationCodes Table",
		SQL: `CREATE TABLE IF NOT EXISTS authorization_codes (
            code VARCHAR(255) PRIMARY KEY,
            client_id BINARY(16) NOT NULL,
            user_id BINARY(16) NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            used BOOLEAN DEFAULT FALSE,
            FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );`,
	},
	{
		Name: "Create RefreshTokens Table",
		SQL: `CREATE TABLE IF NOT EXISTS refresh_tokens (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            token VARCHAR(255) NOT NULL UNIQUE,
            client_id BINARY(16) NOT NULL,
            user_id BINARY(16) NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            revoked BOOLEAN DEFAULT FALSE,
            FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
            INDEX idx_token_lookup (token)
        );`,
	},
	{
		Name: "Create Scopes Table",
		SQL: `CREATE TABLE IF NOT EXISTS scopes (
            id INT AUTO_INCREMENT PRIMARY KEY,
            scope_name VARCHAR(50) NOT NULL UNIQUE,
            description VARCHAR(255)
        );`,
	},
	{
		Name: "Create ClientGrantTypes Table",
		SQL: `CREATE TABLE IF NOT EXISTS client_grant_types (
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
}
