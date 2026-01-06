package procedures

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var CreateUserProcedure = migrations.MigrationPart{
	Name: "create-user-procedure",
	SQL: `
		DROP PROCEDURE IF EXISTS CreateUser;
		CREATE PROCEDURE CreateUser(
			IN p_userId BINARY(16),
			IN p_username VARCHAR(255),
			IN p_firstName VARCHAR(50),
			IN p_middleName VARCHAR(50),
			IN p_lastName VARCHAR(50),
			IN p_userEmail VARCHAR(100),
			IN p_userPasswordHash VARCHAR(255),
			IN p_rolesJson JSON -- This will be '["admin", "user"]'
		)
		BEGIN
			DECLARE EXIT HANDLER FOR SQLEXCEPTION
			BEGIN
				ROLLBACK;
				RESIGNAL;
			END;

			START TRANSACTION;

			-- 1. Insert the User
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
				p_userId, 
				p_username, 
				p_firstName, 
				p_middleName, 
				p_lastName, 
				p_userEmail, 
				p_userPasswordHash
			);
			
			-- 2. Map JSON role names to IDs and Insert into user_roles
			-- We use JSON_TABLE to turn the JSON array into a virtual SQL table
			INSERT INTO user_roles (user_id, role_id)
			SELECT p_userId, r.id
			FROM roles r
			WHERE r.role_name IN (
				SELECT jt.role_name 
				FROM JSON_TABLE(
					p_rolesJson, 
					"$[*]" COLUMNS(role_name VARCHAR(50) PATH "$")
				) AS jt
			);

			-- 3. Audit Log
			INSERT INTO audit_logs (user_id, action, details)
			VALUES (
				p_userId, 
				'create_user', 
				CONCAT('Account ', BINARY_TO_UUID(p_userId), ' created.')
			);

			COMMIT;
		END;
	`,
}
