package main

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/causon-mikolorenz/unified-access-backend/database"
	v1 "github.com/causon-mikolorenz/unified-access-backend/internal/api/v1"
	"github.com/causon-mikolorenz/unified-access-backend/internal/middleware"
	"github.com/causon-mikolorenz/unified-access-backend/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	// 2. Load RSA Keys from Disk (Crucial for Asymmetric Signing)
	privKey, pubKey := loadRSAKeys()

	// 3. Database Migration Flag
	doMigrate := flag.Bool("migrate", false, "Run database migration first")
	flag.Parse()

	if *doMigrate {
		adminDatabase, err := database.ConnectAdminToDB()
		if err != nil {
			log.Fatalf("Error starting database with admin credentials: %v", err)
		}
		defer adminDatabase.Close()

		database.RunAllMigrations(adminDatabase)
		fmt.Println("Database Migrated Successfully")
		return
	}

	// 4. Connect to App Database
	appDB, err := database.ConnectToDB()
	if err != nil {
		log.Fatalf("Application Database failed to start: %v", err)
	}
	defer appDB.Close()

	// 5. Initialize Repositories and Handlers
	// We inject the DB into the repo, and the repo + private key into the handler
	authRepo := repository.NewAuthCodeRepository(appDB)
	authHandler := &v1.AuthHandler{
		Repo:       authRepo,
		PrivateKey: privKey, // Used in /token for signing JWTs
	}

	// 6. Start the Background Janitor (Cleanup expired codes)
	database.StartJanitor(appDB, 10) // 10 minutes interval

	// 7. Initialize Gin and Routes
	r := gin.Default()

	// Register API v1 Group
	v1Group := r.Group("/api/v1")
	{
		// PUBLIC ROUTES: Auth (Login & Token Exchange)
		authRoutes := v1Group.Group("/auth")
		{
			authRoutes.POST("/login", authHandler.LoginAndAuthorize)
			authRoutes.POST("/token", authHandler.ExchangeToken)
		}

		// PROTECTED ROUTES: Admin (Requires RS256 Validation + RBAC)
		adminRoutes := v1Group.Group("/admin")
		adminRoutes.Use(middleware.AuthorizeRBAC(pubKey, "admin"))
		{
			adminRoutes.GET("/status", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "IdP is operational and healthy"})
			})
		}
	}

	fmt.Println("Backend is running on :8080!")
	r.Run(":8080")
}

// Helper function to keep main.go clean
func loadRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey) {
	privBytes, err := os.ReadFile("certs/private.pem")
	if err != nil {
		log.Fatal("Could not read private key: ", err)
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		log.Fatal("Could not parse private key: ", err)
	}

	pubBytes, err := os.ReadFile("certs/public.pem")
	if err != nil {
		log.Fatal("Could not read public key: ", err)
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		log.Fatal("Could not parse public key: ", err)
	}

	return privKey, pubKey
}
