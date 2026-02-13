package procedures

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var CreateUserProcedure = migrations.MigrationPart{
	Name: "create-user-procedure",
	SQL: `
        DROP PROCEDURE IF EXISTS CreateUser;
        
        -- We define the procedure's default collation here
        CREATE PROCEDURE CreateUser(
            IN p_userId BINARY(16),
            IN p_username VARCHAR(255),
            IN p_firstName VARCHAR(50),
            IN p_middleName VARCHAR(50),
            IN p_lastName VARCHAR(50),
            IN p_userEmail VARCHAR(100),
            IN p_userPasswordHash VARCHAR(255),
            IN p_rolesJson JSON
        )
        BEGIN
            DECLARE EXIT HANDLER FOR SQLEXCEPTION
            BEGIN
                ROLLBACK;
                RESIGNAL;
            END;

            START TRANSACTION;

            -- 1. Use COLLATE explicitly in the check to resolve the 'Illegal Mix'
            IF NOT EXISTS (
                SELECT 1 FROM users 
                WHERE email COLLATE utf8mb4_unicode_ci = p_userEmail COLLATE utf8mb4_unicode_ci
                OR username COLLATE utf8mb4_unicode_ci = p_username COLLATE utf8mb4_unicode_ci
            ) THEN
                INSERT INTO users (
                    id, username, first_name, middle_name, last_name, email, password_hash
                )
                VALUES (
                    p_userId, p_username, p_firstName, p_middleName, p_lastName, p_userEmail, p_userPasswordHash
                );  
            
                -- 2. Insert roles with explicit collation casting
                INSERT INTO user_roles (user_id, role_id)
                SELECT p_userId, r.id
                FROM roles r
                WHERE r.role_name COLLATE utf8mb4_unicode_ci IN (
                    SELECT jt.role_name COLLATE utf8mb4_unicode_ci
                    FROM JSON_TABLE(
                        p_rolesJson, 
                        "$[*]" COLUMNS(role_name VARCHAR(50) PATH "$")
                    ) AS jt
                );

                -- 3. Audit Log
                INSERT INTO audit_logs (user_id, action, details)
                VALUES (
                    p_userId, 
                    'USER_CREATION', 
                    CONCAT('Account ', p_username, ' created.')
                );

                COMMIT;
            ELSE
                ROLLBACK;
            END IF;
        END;
    `,
}
