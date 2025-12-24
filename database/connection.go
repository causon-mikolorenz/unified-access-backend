package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func ConnectAdminToDB() (*sql.DB, error) {
	// Implementation for connecting to the database as admin for migration
	var err error
	config := mysql.Config{
		User:                 os.Getenv("ADMIN_SQL_USER"),
		Passwd:               os.Getenv("ADMIN_SQL_PASSWORD"),
		Net:                  "	tcp",
		Addr:                 os.Getenv("MYSQL_ADDRESS"),
		DBName:               os.Getenv("MYSQL_DB_NAME"),
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	return db, nil
}

func ConnectToDB() (*sql.DB, error) {
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

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		fmt.Errorf("Error pinging the database: ", err)
		return nil, err
	}

	return db, nil
}
