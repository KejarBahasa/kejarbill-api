package bootstrap

import (
	authHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/auth/handler"
	authRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/auth/repository"
	authServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/auth/service"

	userHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/handler"
	userRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"
	userServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/service"

	expenseHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/handler"
	expenseRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/repository"
	expenseServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/service"

	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	ledgerServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/config"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/logger"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"
	redisConn "github.com/KejarBahasa/kejarbill-api/internal/shared/redis"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

type Dependency struct {
	Config         *config.Config
	DB             *pgxpool.Pool
	Redis          *goredis.Client
	PasetoMaker    *security.PasetoMaker
	SessionStore   *security.SessionStore
	AuthMiddleware *middleware.AuthMiddleware

	AuthHandler *authHandlerPkg.AuthHandler

	UserHandler *userHandlerPkg.UserHandler

	ExpenseHandler *expenseHandlerPkg.ExpenseHandler
}

func BuildDependency() (*Dependency, error) {
	cfg := config.LoadConfig()
	logger.Init(cfg.AppEnv, cfg.AppName)

	db := database.NewPostgres(cfg)
	rdb := redisConn.NewRedis(cfg.RedisAddr, cfg.RedisPassword)

	pasetoMaker, err := security.NewPasetoMaker(cfg.PasetoSecretKey)
	if err != nil {
		return nil, err
	}

	sessionStore := security.NewSessionStore(rdb)

	authRepo := authRepoPkg.NewAuthRepository(db)
	userRepo := userRepoPkg.NewUserRepository(db)
	ledgerRepo := ledgerRepoPkg.NewLedgerRepository()
	expenseRepo := expenseRepoPkg.NewExpenseRepository()

	authMiddleware := middleware.NewAuthMiddleware(pasetoMaker, authRepo)

	authService := authServicePkg.NewAuthService(authRepo, pasetoMaker, sessionStore, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
	userService := userServicePkg.NewUserService(userRepo)
	ledgerService := ledgerServicePkg.NewLedgerService()
	expenseService := expenseServicePkg.NewExpenseService(db, expenseRepo, ledgerRepo, ledgerService)

	authHandler := authHandlerPkg.NewAuthHandler(authService)
	userHandler := userHandlerPkg.NewUserHandler(userService)
	expenseHandler := expenseHandlerPkg.NewExpenseHandler(expenseService)

	return &Dependency{
		Config:         cfg,
		DB:             db,
		Redis:          rdb,
		PasetoMaker:    pasetoMaker,
		SessionStore:   sessionStore,
		AuthMiddleware: authMiddleware,

		AuthHandler: authHandler,

		UserHandler: userHandler,

		ExpenseHandler: expenseHandler,
	}, nil
}
