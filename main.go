package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/causon-mikolorenz/unified-access-backend/internal/api/v1"
	"github.com/causon-mikolorenz/unified-access-backend/internal/auth"
	"github.com/causon-mikolorenz/unified-access-backend/internal/database"
	"github.com/causon-mikolorenz/unified-access-backend/internal/initializers"
	"github.com/causon-mikolorenz/unified-access-backend/internal/repository"
	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	godotenv.Load()

	// 2. Load RSA Keys (Private for signing, Public for middleware)
	initializers.LoadRSAKeys()

	// 3. Database Migration Flag logic
	doMigrate := flag.Bool("migrate", false, "Run database migration first")
	flag.Parse()

	if *doMigrate {
		// 1. Connect with Admin privileges for schema changes
		adminDatabase, err := database.ConnectAdminToDB()
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		defer adminDatabase.Close()

		// 2. Run Tables Migrations
		database.RunAllMigrations(adminDatabase)
		fmt.Println("Schema Migrations Successful.")

		// 3. Setup Repository for Seeding
		// We use the adminDatabase connection here to ensure we have write perms
		userRepo := repository.NewUserRepository(adminDatabase)

		// 4. Prepare Admin Data
		adminDatabase.Exec("DELETE FROM users WHERE email = 'example@example.com'")

		newAdminID := uuid.New()
		rawPassword := "admin123"
		hashedPassword, _ := auth.HashSecret(rawPassword)

		adminUser := &models.User{
			ID:           newAdminID[:],
			Username:     "Admin",
			Email:        "example@example.com",
			PasswordHash: hashedPassword,
			Roles:        []string{"idp:admin"},
		}

		if err := userRepo.CreateUser(adminUser); err != nil {
			log.Printf("❌ Seeding Failed: %v", err)
		} else {
			fmt.Println("✅ Admin re-seeded with password: admin123")
		}

		testClientID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		hashedSecret, _ := auth.HashSecret("test-secret-123")

		redirectsJSON, _ := json.Marshal([]string{"http://localhost:5173/callback"})
		grantsJSON, _ := json.Marshal([]string{"authorization_code", "refresh_token"})

		// Call the procedure
		_, err = adminDatabase.Exec(
			"CALL RegisterClient(?, ?, ?, ?, ?)",
			testClientID[:],
			"React Dev Client",
			hashedSecret,
			redirectsJSON,
			grantsJSON,
		)

		if err != nil {
			log.Printf("Note: Client seeding skipped or failed: %v", err)
		} else {
			fmt.Println("✅ Test Client 'React Dev Client' registered via procedure.")
		}
	}

	// 4. Connect to App Database
	appDB, err := database.ConnectToDB()
	if err != nil {
		log.Fatalf("Application Database failed to start: %v", err)
	}
	defer appDB.Close()

	// 5. Initialize Layers
	handlerContainer := initializers.InitializeHandlers(appDB)

	// 6. Setup Signal Context for Graceful Shutdown
	// This context will cancel when you press Ctrl+C or when Docker stops the container
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start Background Janitor
	database.StartJanitor(ctx, appDB, 10*time.Minute)

	// 7. Initialize Gin Router
	r := gin.Default()

	// 8. Setup CORS
	r.Use(*&handlerContainer.CORS)

	// 9. Serve static images
	r.Static("/public", "./public")

	// 10. Handle Routes
	v1Group := r.Group("/api/v1")
	v1.MapRoutes(v1Group, *handlerContainer)

	// 11. Configure & Start HTTP Server
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

	// 12. Graceful Shutdown Sequence
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
