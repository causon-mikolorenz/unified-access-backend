package migrations

var Procedures = []Migration{
	{
		Name: "Create Archive/Soft Delete User Procedure",
		SQL: `
		-- Drop the procedure first
		DROP PROCEDURE IF EXISTS ArchiveUser;
		-- Then create to become idempotent
		CREATE PROCEDUREArchiveUser(IN userId BINARY(16))
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
			UPDATE auth_codes
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

			-- Archive user by setting deleted_at
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
}
