package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/causon-mikolorenz/unified-access-backend/docs"
	v1 "github.com/causon-mikolorenz/unified-access-backend/internal/api/v1"
	"github.com/causon-mikolorenz/unified-access-backend/internal/database"
	"github.com/causon-mikolorenz/unified-access-backend/internal/initializers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Unified Access Identity Provider API
// @version 1.0
// @description Identity Provider for the Unified School System Capstone.
// @termsOfService http://swagger.io/terms/
// @contact.name Miko Lorenz Causon
// @contact.email causonmikolorenz@gmail.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
func main() {
	godotenv.Load()
	initializers.LoadRSAKeys()

	doMigrate := flag.Bool("migrate", false, "Run database migration first")
	flag.Parse()

	if *doMigrate {
		initializers.MigrateAndSeed()
	}

	appDB, err := database.ConnectToDB()
	if err != nil {
		log.Fatalf("[Main] App Database Connection Failed: %v", err)
	}
	defer appDB.Close()

	h := initializers.InitializeHandlers(appDB)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	database.StartJanitor(ctx, appDB, 10*time.Minute)

	r := gin.Default()
	r.Use(h.CORS)
	r.Static("/public", "./public")

	// Register Swagger UI route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1Group := r.Group("/api/v1")
	v1.MapRoutes(v1Group, *h)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Main] Server Listen Error: %v", err)
		}
	}()

	fmt.Println("Backend is operational on :8080")

	<-ctx.Done()

	stop()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[Main] Forced Shutdown Error: %v", err)
	}

	log.Println("Server exited")
}
