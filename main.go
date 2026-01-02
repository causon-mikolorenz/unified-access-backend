package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/causon-mikolorenz/unified-access-backend/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Create flag for migration
	doMigrate := flag.Bool("migrate", false, "Run database migration first")
	flag.Parse()

	if *doMigrate {
		// Connect to the database with admin creds
		adminDatabase, err := database.ConnectAdminToDB()
		// Check for errors
		if err != nil {
			log.Fatalf(
				"Error starting database with admin credentials: %v",
				err,
			)
		}
		// Close before exiting
		defer adminDatabase.Close()

		// Do the migrations
		err = database.RunAllMigrations(adminDatabase)
		fmt.Println("Database Migrated Successfully")

		// Exit after migrating
		return
	}

	// Connect to the database
	appDB, err := database.ConnectToDB()
	if err != nil {
		log.Fatalf("Application Database failed to start: %v", err)
	}

	// Close the Application DB when the program exits
	defer appDB.Close()

	fmt.Println("Backend is running!")

	r := gin.Default()
	r.Run(":8080")
}
