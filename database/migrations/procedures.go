package migrations

var Procedures = []Migration{
	{
		Name: "Create Archive/Soft Delete User Procedure",
		SQL: `
		-- Drop the procedure first
		DROP PROCEDURE IF EXISTS ArchiveUser;
		-- Then create to become idempotent
		CREATE PROCEDURE ArchiveUser(IN userId BINARY(16))
		BEGIN
			DECLARE EXIT HANDLER FOR SQLEXCEPTION
			BEGIN
				ROLLBACK;
			END;

			START TRANSACTION;

			-- Archive user by setting deleted_at
			UPDATE users
			SET deleted_at = NOW(), status = 'deleted' 
			WHERE id = userId AND deleted_at IS NULL;
			-- I expire ang active refresh tokens ni user
			UPDATE refresh_tokens
			SET expires_at = NOW()
			WHERE user_id = userId AND expires_at > NOW();
			-- I expire yung auth codes na may connection si user
			UPDATE authorization_codes
			SET expires_at = NOW()
			WHERE user_id = userId AND expires_at > NOW();
			-- Ilagay sa audit logs
			INSERT INTO audit_logs (user_id, action, details)
			VALUES (
				userId, 
				'archive_user', 
				CONCAT('User ', HEX(userId), ' was archived.')
			);
			COMMIT;
		END;`,
	},
	{
		Name: "Create User Procedure",
		SQL: `
		-- Drop the procedure first
		DROP PROCEDURE IF EXISTS CreateUser;
		-- Then create to become idempotent
		CREATE PROCEDURE CreateUser(
			IN userId BINARY(16),
			IN username varchar(255),
			IN firstName VARCHAR(50),
			IN middleName VARCHAR(50),
			IN lastName VARCHAR(50),
			IN userEmail VARCHAR(100),
			IN userPasswordHash VARCHAR(255)
		)
		BEGIN
			DECLARE EXIT HANDLER FOR SQLEXCEPTION
			BEGIN
				ROLLBACK;
			END;

			START TRANSACTION;

			-- Create user by inserting into users table
			INSERT INTO users (
				id,
				username,
				first_name,
				middle_name,
				last_name,
				email,
				password_hash
			)
			VALUES (
				userId,
				username,
				firstName,
				middleName,
				lastName,
				userEmail,
				userPasswordHash
			);
			
			-- Ilagay sa audit logs
			INSERT INTO audit_logs (user_id, action, details)
			VALUES (
				userId, 
				'Create user', 
				CONCAT('User ', HEX(userId), ' was created.')
			);
			COMMIT;
		END;`,
	},
	{
		Name: "Update User Password Procedure",
		SQL: `
		-- Drop the procedure first
		DROP PROCEDURE IF EXISTS UpdateUserPassword;
		-- Then create to become idempotent
		CREATE PROCEDURE UpdateUserPassword(
			IN userId BINARY(16),
			IN newPasswordHash VARCHAR(255)
		)
		BEGIN
			DECLARE EXIT HANDLER FOR SQLEXCEPTION
			BEGIN
				ROLLBACK;
			END;

			START TRANSACTION;

			-- Update user password hash
			UPDATE users
			SET password_hash = newPasswordHash, updated_at = NOW()
			WHERE id = userId AND deleted_at IS NULL;
			-- Terminate all active sessions by expiring their refresh tokens
			UPDATE refresh_tokens
			SET expires_at = NOW()
			WHERE user_id = userId AND expires_at > NOW();
			-- Terminate all active auth codes
			UPDATE authorization_codes
			SET expires_at = NOW()
			WHERE user_id = userId AND expires_at > NOW();
			-- Ilagay sa audit logs
			INSERT INTO audit_logs (user_id, action, details)
			VALUES (
				userId, 
				'Update user password', 
				CONCAT('User ', HEX(userId), ' password was updated.')
			);
			COMMIT;
		END;`,
	},
	{
		Name: "Create Register Client (Service Provider) Procedure",
		SQL: `
		-- Drop the procedure first
		DROP PROCEDURE IF EXISTS RegisterClient;
		-- Then create to become idempotent
		CREATE PROCEDURE RegisterClient(
			IN clientId BINARY(16),
			IN clientName VARCHAR(100),
			IN clientSecretHash VARCHAR(255),
			IN redirectURIs JSON
		)
		BEGIN
			DECLARE EXIT HANDLER FOR SQLEXCEPTION
			BEGIN
				ROLLBACK;
			END;
			-- Declare loop variables
			DECLARE i INT DEFAULT 0;
    		DECLARE uri_count INT DEFAULT JSON_LENGTH(redirectURIs);

			START TRANSACTION;
			-- Insert new client (service provider)
			INSERT INTO clients (
				id,
				client_name,
				client_secret
			)			
			VALUES (
				clientId,
				clientName,
				clientSecretHash
			);

			-- Insert redirect URIs by looping through JSON
			while i < uri_count DO
				INSERT INTO client_urls (
					client_id,
					redirect_url
				)
				VALUES (
					clientId,
					JSON_UNQUOTE(
						JSON_EXTRACT(
							redirectURIs, 
							CONCAT('$[', i, ']')
						)
					)
				);
				SET i = i + 1;
			END WHILE;

			-- Ilagay sa audit logs
			INSERT INTO audit_logs (user_id, action, details)
			VALUES (
				NULL, 
				'Register client', 
				CONCAT('Client ', HEX(clientId), ' was registered.')
			);
			COMMIT;
		END;`,
	},
	{
		Name: "Create Authorization Code Exchange Procedure",
		SQL: `
		DROP PROCEDURE IF EXISTS ExchangeAuthorizationCode;
		CREATE PROCEDURE ExchangeAuthorizationCode(
			IN  p_code VARCHAR(255),
			IN  p_client_id BINARY(16),
			OUT p_user_id BINARY(16)
		)
		BEGIN
			DECLARE v_user_id BINARY(16) DEFAULT NULL;
			DECLARE EXIT HANDLER FOR SQLEXCEPTION 
			BEGIN 
				SET p_user_id = NULL; 
				ROLLBACK; 
			END;

			START TRANSACTION;
				SELECT user_id INTO v_user_id 
				FROM authorization_codes 
				WHERE code = p_code 
				AND client_id = p_client_id
				AND used = FALSE 
				AND expires_at > NOW()
				FOR UPDATE;

				IF v_user_id IS NOT NULL THEN
					UPDATE authorization_codes 
					SET used = TRUE 
					WHERE code = p_code;

					INSERT INTO audit_logs (user_id, action, details)
					VALUES (
						v_user_id, 
						'auth_code_exchange', 
						'Code successfully exchanged.'
					);
					
					SET p_user_id = v_user_id;
				ELSE
					SET p_user_id = NULL;
				END IF;
			COMMIT;

			SELECT p_user_id AS user_id;
		END;`,
	},
	{
		Name: "Create Logout Stored Procedure",
		SQL: `
		-- Drop the procedure first
		DROP PROCEDURE IF EXISTS LogoutUser;
		-- Then create to become idempotent
		CREATE PROCEDURE LogoutUser(IN userId BINARY(16))
		BEGIN
			DECLARE EXIT HANDLER FOR SQLEXCEPTION
			BEGIN
				ROLLBACK;
			END;

			START TRANSACTION;

			-- I expire ang active refresh tokens ni user
			UPDATE refresh_tokens
			SET expires_at = NOW()
			WHERE user_id = userId AND expires_at > NOW();
			-- I expire yung auth codes na may connection si user
			UPDATE authorization_codes
			SET expires_at = NOW()
			WHERE user_id = userId AND expires_at > NOW();
			-- Ilagay sa audit logs
			INSERT INTO audit_logs (user_id, action, details)
			VALUES (
				userId, 
				'logout', 
				CONCAT('User ', HEX(userId), ' logged out.')
			);
			COMMIT;
		END;`,
	},
}
