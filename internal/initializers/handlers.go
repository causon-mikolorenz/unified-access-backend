package initializers

import (
	v1 "github.com/causon-mikolorenz/unified-access-backend/internal/api/v1"
	"github.com/causon-mikolorenz/unified-access-backend/internal/repository"
	"github.com/jmoiron/sqlx"
)

type HandlerContainer struct {
	AuthHandler *v1.AuthHandler
}

func InitializeHandlers(db *sqlx.DB) *HandlerContainer {
	// Initialize Repositories
	authRepo := repository.NewAuthCodeRepository(db)

	// Create Handlers
	authHandler := &v1.AuthHandler{
		Repo:       authRepo,
		PrivateKey: PrivKey,
	}

	// Return Handlers
	return &HandlerContainer{
		AuthHandler: authHandler,
	}
}
