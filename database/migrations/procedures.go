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
}
