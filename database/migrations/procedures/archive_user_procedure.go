package procedures

import "github.com/causon-mikolorenz/unified-access-backend/database/migrations"

var ArchiveUserProcedure = migrations.MigrationPart{
	Name: "create-archive/soft-delete-user-procedure",
	SQL: `
		DROP PROCEDURE IF EXISTS ArchiveUser;
		CREATE PROCEDURE ArchiveUser(IN userId BINARY(16))
		BEGIN
			DECLARE user_exists INT DEFAULT 0;

			-- Exit handler for rollback
			DECLARE EXIT HANDLER FOR SQLEXCEPTION
			BEGIN
				ROLLBACK;
				RESIGNAL; -- Pass the error back to the Go application
			END;

			START TRANSACTION;

			-- Check if the user exists and isn't already deleted
			SELECT COUNT(*) INTO user_exists 
			FROM users 
			WHERE id = userId AND deleted_at IS NULL;

			IF user_exists > 0 THEN
				-- 1. Soft delete the user
				UPDATE users
				SET deleted_at = NOW(), status = 'deleted' 
				WHERE id = userId;

				-- 2. Invalidate all active sessions/tokens
				UPDATE refresh_tokens
				SET expires_at = NOW()
				WHERE user_id = userId AND expires_at > NOW();

				UPDATE authorization_codes
				SET expires_at = NOW()
				WHERE user_id = userId AND expires_at > NOW();
				
				-- 3. Log the action
				INSERT INTO audit_logs (user_id, action, details)
				VALUES (
					userId, 
					'archive_user', 
					CONCAT('User ', HEX(userId), ' was successfully archived.')
				);
			ELSE
				-- Optional: Throw an error if user doesn't exist
				SIGNAL SQLSTATE '45000' 
				SET MESSAGE_TEXT = 'User not found or already archived';
			END IF;

			COMMIT;
		END;
	`,
}
