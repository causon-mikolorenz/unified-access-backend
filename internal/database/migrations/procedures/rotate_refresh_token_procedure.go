package procedures

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var RotateRefreshTokenProcedure = migrations.MigrationPart{
    Name: "rotate-refresh-token-procedure",
    SQL: `
        DROP PROCEDURE IF EXISTS RotateRefreshToken;
        CREATE PROCEDURE RotateRefreshToken(
            IN p_oldToken VARCHAR(255),
            IN p_newToken VARCHAR(255),
            IN p_newExpiresAt TIMESTAMP
        )
        BEGIN
            DECLARE v_userId BINARY(16);
            DECLARE v_clientId BINARY(16);
            DECLARE v_revokedAt TIMESTAMP;
            DECLARE v_expiresAt TIMESTAMP;

            -- Exit handler for unexpected system errors
            DECLARE EXIT HANDLER FOR SQLEXCEPTION
            BEGIN
                ROLLBACK;
                RESIGNAL;
            END;

            START TRANSACTION;

            -- 1. Look up the old token and lock the row
            -- If not found, MySQL will throw an error or we handle v_userId being NULL
            SELECT user_id, client_id, revoked_at, expires_at 
            INTO v_userId, v_clientId, v_revokedAt, v_expiresAt
            FROM refresh_tokens 
            WHERE token = p_oldToken FOR UPDATE;

            IF v_userId IS NULL THEN
                ROLLBACK;
                SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Token not found.';
            END IF;

            -- 2. Security Check: Detection of Replay Attack
            IF v_revokedAt IS NOT NULL THEN
                -- REPLAY ATTACK: Revoke everything for this user/client pair
                UPDATE refresh_tokens 
                SET revoked_at = NOW() 
                WHERE user_id = v_userId AND client_id = v_clientId AND revoked_at IS NULL;
                
                INSERT INTO audit_logs (user_id, action, details)
                VALUES (v_userId, 'token_replay_attack', 
                        CONCAT('Replay detected for token replaced by ', IFNULL(p_newToken, 'unknown')));
                
                -- COMMIT the punishment so it isn't rolled back
                COMMIT; 
                
                SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Security Breach: Token reuse detected.';
            END IF;

            -- 3. Check for Expiration
            IF v_expiresAt < NOW() THEN
                -- No need to revoke the whole chain for an expiration, just roll back
                ROLLBACK;
                SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Refresh token has expired.';
            END IF;

            -- 4. Successful Exchange
            -- Mark old as revoked and link it to the new one
            UPDATE refresh_tokens 
            SET revoked_at = NOW(), replaced_by = p_newToken 
            WHERE token = p_oldToken;

            -- Insert the new token
            INSERT INTO refresh_tokens (token, client_id, user_id, expires_at)
            VALUES (p_newToken, v_clientId, v_userId, p_newExpiresAt);

            -- 5. Audit the rotation
            INSERT INTO audit_logs (user_id, action, details)
            VALUES (v_userId, 'token_rotation', 'Token rotated successfully.');

            COMMIT;
        END;`,
}