package initializers

import (
	"fmt"
	"log"
	"os"

	"github.com/causon-mikolorenz/unified-access-backend/internal/auth"
	"github.com/causon-mikolorenz/unified-access-backend/internal/database"
	"github.com/causon-mikolorenz/unified-access-backend/internal/repository"
	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/google/uuid"
)

/**
 * Migrate handles the database schema updates and initial data seeding.
 * It uses administrative privileges to ensure schema changes are permitted.
 */
func MigrateAndSeed() {
	adminDatabase, err := database.ConnectAdminToDB()
	if err != nil {
		log.Fatalf("[Migrate] What failed: %v", err)
	}
	defer adminDatabase.Close()

	database.RunAllMigrations(adminDatabase)
	fmt.Println("Database migration completed successfully.")

	userRepo := repository.NewUserRepository(adminDatabase)

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	adminUser := os.Getenv("ADMIN_USERNAME")

	adminDatabase.Exec("DELETE FROM users WHERE email = ?", adminEmail)

	newAdminID := uuid.New()
	hashedPassword, _ := auth.HashSecret(adminPass)

	user := &models.User{
		ID:           newAdminID[:],
		Username:     adminUser,
		Email:        adminEmail,
		PasswordHash: hashedPassword,
		Roles:        []string{"idp:admin"},
	}

	if err := userRepo.CreateUser(user); err != nil {
		log.Printf("[Migrate] Admin seeding failed: %v", err)
	} else {
		fmt.Printf("Admin user %s seeded successfully.\n", adminEmail)
	}

	// Fetch client magic values from environment
	cID := os.Getenv("CLIENT_ID")
	cSecret := os.Getenv("CLIENT_SECRET")
	cName := os.Getenv("CLIENT_NAME")
	cCallback := os.Getenv("CLIENT_CALLBACK")
	cBase := os.Getenv("CLIENT_BASE_URL")
	cAbbreviation := os.Getenv("CLIENT_ABBREVIATION")

	parsedID, _ := uuid.Parse(cID)
	hashedClientSecret, _ := auth.HashSecret(cSecret)
	grants := []string{"authorization_code", "refresh_token"}

	// Initialize the struct to avoid nil pointer
	client := &models.Client{
		ID:            parsedID[:],
		ClientName:    cName,
		Abbreviation:  cAbbreviation,
		ClientSecret:  hashedClientSecret,
		BaseUrl:       cBase,
		RedirectUri:   cCallback,
		LogoutUri:     cBase,
		Description:   "",
		ImageLocation: "",
	}

	clientRepo := repository.NewClientRepository(adminDatabase)
	err = clientRepo.CreateClient(client, grants)
	if err != nil {
		log.Printf("[Migrate] Client seeding failed: %v", err)
	} else {
		fmt.Printf("Client %s registered successfully.\n", cName)
	}
}
