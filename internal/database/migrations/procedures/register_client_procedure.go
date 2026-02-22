package procedures

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var RegisterClientProcedure = migrations.MigrationPart{
	Name: "register-client-procedure",
	SQL: `
        DROP PROCEDURE IF EXISTS RegisterClient;
        CREATE PROCEDURE RegisterClient(
            IN p_clientId BINARY(16),
            IN p_clientName VARCHAR(100),
            IN p_clientSecretHash VARCHAR(255),
            IN p_redirectUrisJson JSON,
            IN p_grantTypesJson JSON
        )
        BEGIN
            -- Exit handler for atomic rollback
            DECLARE EXIT HANDLER FOR SQLEXCEPTION
            BEGIN
                ROLLBACK;
                RESIGNAL;
            END;

            START TRANSACTION;

            -- 1. Insert the Client Identity
            INSERT INTO clients (id, client_name, client_secret)           
            VALUES (p_clientId, p_clientName, p_clientSecretHash);

            -- 2. Map & Bulk Insert Redirect URIs
            INSERT INTO client_urls (client_id, redirect_uri)
            SELECT p_clientId, jt.url
            FROM JSON_TABLE(
                p_redirectUrisJson, 
                "$[*]" COLUMNS(url VARCHAR(2048) PATH "$")
            ) AS jt;

            -- 3. Map & Bulk Insert Allowed Grant Types
            -- (Aligned column names to grant_name for keyword safety)
            INSERT INTO client_grant_types (client_id, grant_type)
            SELECT p_clientId, jt.grant_name
            FROM JSON_TABLE(
                p_grantTypesJson, 
                "$[*]" COLUMNS(grant_name VARCHAR(50) PATH "$")
            ) AS jt;

            -- 4. Audit Trail (Uses p_clientId)
            INSERT INTO audit_logs (user_id, action, details)
            VALUES (
                NULL, 
                'register_client', 
                CONCAT(
					'Client ', 
					BINARY_TO_UUID(p_clientId), 
					' was registered.'
				)
            );

            COMMIT;
        END;
    `,
}
