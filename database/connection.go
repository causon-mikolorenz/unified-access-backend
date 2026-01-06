package database

import (
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func ConnectAdminToDB() (*sqlx.DB, error) {
	// Implementation for connecting to the database as admin for migration
	var err error
	config := mysql.Config{
		User:                 os.Getenv("ADMIN_SQL_USER"),
		Passwd:               os.Getenv("ADMIN_SQL_PASSWORD"),
		Net:                  "tcp",
		Addr:                 os.Getenv("MYSQL_ADDRESS"),
		DBName:               os.Getenv("MYSQL_DB_NAME"),
		AllowNativePasswords: true,
		ParseTime:            true,
		MultiStatements:      true,
	}

	db, err := sqlx.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("Error connecting to database: %w", err)
	}

	return db, nil
}

func ConnectToDB() (*sqlx.DB, error) {
	// Implementation for connecting to the database as the application service
	var err error
	config := mysql.Config{
		User:                 os.Getenv("APP_USER"),
		Passwd:               os.Getenv("APP_PASSWORD"),
		Net:                  "tcp",
		Addr:                 os.Getenv("MYSQL_ADDRESS"),
		DBName:               os.Getenv("MYSQL_DB_NAME"),
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	db, err := sqlx.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("Error connecting to database: %w", err)
	}

	maxConnections := 25
	maxLifetime := 5 * time.Minute

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxConnections)
	db.SetConnMaxLifetime(maxLifetime)

	if err = db.Ping(); err != nil {
		db.Close()

		return nil, fmt.Errorf("Error pinging the database at %s: %w", config.Addr, err)
	}

	return db, nil
}
