package procedures

import "github.com/causon-mikolorenz/unified-access-backend/database/migrations"

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

            -- Exit handler for atomic rollback
            DECLARE EXIT HANDLER FOR SQLEXCEPTION
            BEGIN
                ROLLBACK;
                RESIGNAL;
            END;

            START TRANSACTION;

            -- 1. Look up the old token and lock the row
            SELECT user_id, client_id, revoked_at, expires_at 
            INTO v_userId, v_clientId, v_revokedAt, v_expiresAt
            FROM refresh_tokens 
            WHERE token = p_oldToken FOR UPDATE;

            -- 2. Security Check: Detection of Replay Attack
            -- If the token was already revoked, someone is trying to reuse a "burned" token.
            IF v_revokedAt IS NOT NULL THEN
                -- REPLAY ATTACK DETECTED: Revoke all tokens for this specific session
                UPDATE refresh_tokens 
                SET revoked_at = NOW() 
                WHERE user_id = v_userId AND client_id = v_clientId;
                
                -- Audit the breach attempt
                INSERT INTO audit_logs (user_id, action, details)
                VALUES (v_userId, 'token_replay_attack', 
                        CONCAT(
							'Replay detected for client ', 
							HEX(v_clientId), 
							'. Entire chain revoked.'));
                
                SIGNAL SQLSTATE '45000' 
                SET MESSAGE_TEXT = 'Security Breach: Token reuse detected.';
            END IF;

            -- 3. Check for Expiration
            IF v_expiresAt < NOW() THEN
                SIGNAL SQLSTATE '45000' 
                SET MESSAGE_TEXT = 'Refresh token has expired.';
            END IF;

            -- 4. Successful Exchange: Burn the old, Issue the new
            UPDATE refresh_tokens 
            SET revoked_at = NOW(), replaced_by = p_newToken 
            WHERE token = p_oldToken;

            INSERT INTO refresh_tokens (token, client_id, user_id, expires_at)
            VALUES (p_newToken, v_clientId, v_userId, p_newExpiresAt);

            -- 5. Audit the rotation
            INSERT INTO audit_logs (user_id, action, details)
            VALUES (
				v_userId, 
				'token_rotation', 
				'Refresh token rotated successfully.'
			);

            COMMIT;
        END;`,
}
