package procedures

import "github.com/causon-mikolorenz/unified-access-backend/internal/database/migrations"

var GetUserForAuthProcedure = migrations.MigrationPart{
	Name: "get-user-for-auth-procedure",
	SQL: `
        DROP PROCEDURE IF EXISTS GetUserForAuth;
        CREATE PROCEDURE GetUserForAuth(IN p_email VARCHAR(255))
        BEGIN
            SELECT 
                u.id,
				u.email, 
				u.first_name,
				u.middle_name,
				u.last_name,
                u.username, 
                u.password_hash, 
                u.status,
                -- Return roles as a comma-separated string
                IFNULL((SELECT GROUP_CONCAT(r.role_name) 
                 FROM user_roles ur 
                 JOIN roles r ON ur.role_id = r.id 
                 WHERE ur.user_id = u.id), '') as roles
            FROM users u
            WHERE u.email = p_email AND u.status != 'deleted'
            LIMIT 1;
        END;`,
}
