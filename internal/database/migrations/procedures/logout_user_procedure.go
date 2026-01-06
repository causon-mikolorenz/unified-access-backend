package procedures

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var LogoutUserProcedure = migrations.MigrationPart{
	Name: "logout-user-procedure",
	SQL: `
        DROP PROCEDURE IF EXISTS LogoutUser;
        CREATE PROCEDURE LogoutUser(IN p_userId BINARY(16))
        BEGIN
            DECLARE EXIT HANDLER FOR SQLEXCEPTION
            BEGIN
                ROLLBACK;
                RESIGNAL;
            END;

            START TRANSACTION;

            -- 1. Invalidate Browser Sessions
            -- This ensures the IdP no longer recognizes the browser cookie
            UPDATE idp_sessions 
            SET expires_at = NOW() 
            WHERE user_id = p_userId AND expires_at > NOW();

            -- 2. Invalidate Refresh Tokens
            -- Use revoked_at if you have it, or set expires_at to NOW()
            UPDATE refresh_tokens
            SET revoked_at = NOW()
            WHERE user_id = p_userId AND revoked_at IS NULL;

            -- 3. Invalidate Authorization Codes
            -- Prevents pending codes from being exchanged for tokens
            UPDATE authorization_codes
            SET expires_at = NOW()
            WHERE user_id = p_userId AND expires_at > NOW();

            -- 4. Audit Log
            INSERT INTO audit_logs (user_id, action, details)
            VALUES (
                p_userId, 
                'logout', 
                CONCAT(
					'User ', 
					BINARY_TO_UUID(p_userId), 
					' performed a global logout.'
				)
            );

            COMMIT;
        END;`,
}
