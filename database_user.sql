CREATE USER 'idp_admin'@'%' IDENTIFIED BY '<idpsqladmin="q1Q!w2W@">';

GRANT ALL PRIVILEGES ON identity_db.* TO 'idp_admin'@'%';
FLUSH PRIVILEGES;

CREATE USER 'idp_svc'@'%' IDENTIFIED BY '<sqlidpsvc="3306user">';

GRANT SELECT, INSERT, UPDATE, EXECUTE ON identity_db.* TO 'idp_svc'@'%';
FLUSH PRIVILEGES;