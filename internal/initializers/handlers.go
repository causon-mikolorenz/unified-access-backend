package initializers

import (
	v1 "github.com/causon-mikolorenz/unified-access-backend/internal/api/v1"
	"github.com/causon-mikolorenz/unified-access-backend/internal/repository"
	"github.com/jmoiron/sqlx"
)

type HandlerContainer struct {
	AuthHandler   *v1.AuthHandler
	ClientHandler *v1.ClientHandler
	RoleHandler   *v1.RoleHandler
	UserHandler   *v1.UserHandler
}

func InitializeHandlers(db *sqlx.DB) *HandlerContainer {
	// Initialize Repositories
	authRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	clientRepo := repository.NewClientRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create Handlers
	authHandler := &v1.AuthHandler{
		Repo:        authRepo,
		SessionRepo: sessionRepo,
		PrivateKey:  PrivKey,
	}
	clientHandler := &v1.ClientHandler{
		Repo: clientRepo,
		PrivateKey: PrivKey,
	}
	roleHandler := &v1.RoleHandler{
		Repo: roleRepo,
	}
	userHandler := &v1.UserHandler{
		Repo: userRepo,
	}


	// Return Handlers
	return &HandlerContainer{
		AuthHandler: authHandler,
		ClientHandler: clientHandler,
		RoleHandler: roleHandler,
		UserHandler: userHandler,
	}
}
