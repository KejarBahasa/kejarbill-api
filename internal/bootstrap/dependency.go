package bootstrap

import (
	authHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/auth/handler"
	authRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/auth/repository"
	authServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/auth/service"

	balanceHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/handler"
	balanceServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"

	expenseHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/handler"
	expenseRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/repository"
	expenseServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/service"

	groupHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/handler"
	groupRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/repository"
	groupServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group/service"

	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"

	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	ledgerServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"

	userHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/handler"
	userRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"
	userServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/service"

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

	GroupHandler *groupHandlerPkg.GroupHandler

	ExpenseHandler *expenseHandlerPkg.ExpenseHandler

	BalanceHandler *balanceHandlerPkg.BalanceHandler
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
	groupRepo := groupRepoPkg.NewGroupRepository()
	groupParticipantRepo := groupParticipantRepoPkg.NewGroupParticipantRepository()
	expenseRepo := expenseRepoPkg.NewExpenseRepository()

	authMiddleware := middleware.NewAuthMiddleware(pasetoMaker, authRepo)

	authService := authServicePkg.NewAuthService(authRepo, pasetoMaker, sessionStore, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
	userService := userServicePkg.NewUserService(userRepo)
	groupService := groupServicePkg.NewGroupService(db, groupRepo, groupParticipantRepo)
	ledgerService := ledgerServicePkg.NewLedgerService()
	expenseService := expenseServicePkg.NewExpenseService(db, expenseRepo, ledgerRepo, ledgerService, groupRepo, groupParticipantRepo)
	balanceService := balanceServicePkg.NewBalanceService(db, ledgerRepo, groupRepo, groupParticipantRepo)

	authHandler := authHandlerPkg.NewAuthHandler(authService)
	userHandler := userHandlerPkg.NewUserHandler(userService)
	groupHandler := groupHandlerPkg.NewGroupHandler(groupService)
	expenseHandler := expenseHandlerPkg.NewExpenseHandler(expenseService)
	balanceHandler := balanceHandlerPkg.NewBalanceHandler(balanceService)

	return &Dependency{
		Config:         cfg,
		DB:             db,
		Redis:          rdb,
		PasetoMaker:    pasetoMaker,
		SessionStore:   sessionStore,
		AuthMiddleware: authMiddleware,

		AuthHandler: authHandler,

		UserHandler: userHandler,

		GroupHandler: groupHandler,

		ExpenseHandler: expenseHandler,

		BalanceHandler: balanceHandler,
	}, nil
}
