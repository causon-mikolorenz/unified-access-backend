package procedures

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var UpdateUserPasswordProcedure = migrations.MigrationPart{
	Name: "update-user-password-procedure",
	SQL: `
        DROP PROCEDURE IF EXISTS UpdateUserPassword;
        CREATE PROCEDURE UpdateUserPassword(
            IN p_userId BINARY(16),
            IN p_newPasswordHash VARCHAR(255)
        )
        BEGIN
            DECLARE EXIT HANDLER FOR SQLEXCEPTION
            BEGIN
                ROLLBACK;
                RESIGNAL;
            END;

            START TRANSACTION;

            -- 1. Update password hash
            UPDATE users
            SET password_hash = p_newPasswordHash, updated_at = NOW()
            WHERE id = p_userId AND status != 'deleted';

            -- 2. SECURITY KILL-SWITCH: Invalidate ALL active states
            -- This forces the user to log in again on all devices
            
            -- Invalidate Browser Sessions
            UPDATE idp_sessions 
            SET expires_at = NOW() 
            WHERE user_id = p_userId AND expires_at > NOW();

            -- Invalidate Refresh Tokens (Global Sign-out)
            UPDATE refresh_tokens
            SET revoked_at = NOW() 
            WHERE user_id = p_userId AND revoked_at IS NULL;

            -- Invalidate pending Authorization Codes
            UPDATE authorization_codes
            SET expires_at = NOW()
            WHERE user_id = p_userId AND expires_at > NOW();

            -- 3. Audit Log
            INSERT INTO audit_logs (user_id, action, details)
            VALUES (
                p_userId, 
                'password_change', 
                CONCAT(
                    'Password updated and all sessions revoked for user ', 
                    BIN_TO_UUID(p_userId)
                )
            );
            COMMIT;
        END;
    `,
}
