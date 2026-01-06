package main

import (
	"context"
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	godotenv.Load()

	// 2. Load RSA Keys (Private for signing, Public for middleware)
	privKey, pubKey := loadRSAKeys()

	// 3. Database Migration Flag logic
	doMigrate := flag.Bool("migrate", false, "Run database migration first")
	flag.Parse()

	if *doMigrate {
		adminDatabase, _ := database.ConnectAdminToDB()
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

	// 5. Initialize Layers
	authRepo := repository.NewAuthCodeRepository(appDB)
	authHandler := &v1.AuthHandler{
		Repo:       authRepo,
		PrivateKey: privKey,
	}

	// 6. Setup Signal Context for Graceful Shutdown
	// This context will cancel when you press Ctrl+C or when Docker stops the container
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start Background Janitor
	database.StartJanitor(ctx, appDB, 10*time.Minute)

	// 7. Initialize Gin Router & Routes
	r := gin.Default()

	v1Group := r.Group("/api/v1")
	{
		// Public Auth Endpoints
		auth := v1Group.Group("/auth")
		{
			auth.POST("/login", authHandler.LoginAndAuthorize)
			auth.POST("/token", authHandler.ExchangeToken)
		}

		// Protected Admin Endpoints (Using the pre-loaded pubKey)
		admin := v1Group.Group("/admin")
		admin.Use(middleware.AuthorizeRBAC(pubKey, "admin"))
		{
			admin.GET("/status", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "IdP is operational"})
			})
		}
	}

	// 8. Configure & Start HTTP Server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r, // Gin is the handler here
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	fmt.Println("Backend is running on :8080!")

	// 9. Graceful Shutdown Sequence
	<-ctx.Done() // Wait for SIGINT/SIGTERM

	stop()
	log.Println("Shutting down gracefully... (timeout 5s)")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

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
