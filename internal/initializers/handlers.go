package initializers

import (
	v1 "github.com/causon-mikolorenz/unified-access-backend/internal/api/v1"
	"github.com/causon-mikolorenz/unified-access-backend/internal/middleware"
	"github.com/causon-mikolorenz/unified-access-backend/internal/repository"
	"github.com/jmoiron/sqlx"
)

func InitializeHandlers(db *sqlx.DB) *v1.Handlers {
	authRepo := repository.NewAuthCodeRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	clientRepo := repository.NewClientRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)

	mw := &middleware.Middleware{ClientRepo: clientRepo}

	return &v1.Handlers{
		AuthHandler: &v1.AuthHandler{
			Repo:        authRepo,
			SessionRepo: sessionRepo,
			PrivateKey:  PrivKey,
		},
		ClientHandler: &v1.ClientHandler{
			Repo:       clientRepo,
			PrivateKey: PrivKey,
		},
		RoleHandler: &v1.RoleHandler{
			Repo: roleRepo,
		},
		UserHandler: &v1.UserHandler{
			Repo: userRepo,
		},
		PubKey: PubKey,
		CORS:   mw.CORSMiddleware(),
	}
}
