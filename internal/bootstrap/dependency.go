package bootstrap

import (
	"log"

	activityHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/activity/handler"
	activityServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/activity/service"

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

	groupMemberHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/handler"
	groupMemberRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/repository"
	groupMemberServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/service"

	groupParticipantHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/handler"
	groupParticipantRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/repository"
	groupParticipantServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/service"

	ledgerRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/repository"
	ledgerServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/service"

	paymentMethodHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/handler"
	paymentMethodRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/repository"
	paymentMethodServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/service"

	settlementHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/handler"
	settlementRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/repository"
	settlementServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/service"

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
	Encryption     *security.Encryption
	AuthMiddleware *middleware.AuthMiddleware

	AuthHandler *authHandlerPkg.AuthHandler

	UserHandler *userHandlerPkg.UserHandler

	GroupHandler *groupHandlerPkg.GroupHandler

	GroupMemberHandler *groupMemberHandlerPkg.GroupMemberHandler

	GroupParticipantHandler *groupParticipantHandlerPkg.GroupParticipantHandler

	ExpenseHandler *expenseHandlerPkg.ExpenseHandler

	ActivityHandler *activityHandlerPkg.ActivityHandler

	BalanceHandler *balanceHandlerPkg.BalanceHandler

	SettlementHandler *settlementHandlerPkg.SettlementHandler

	PaymentMethodHandler *paymentMethodHandlerPkg.PaymentMethodHandler
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

	encryption, err := security.NewEncryption(cfg.PaymentMethodEncryptionKey)
	if err != nil {
		log.Fatalf("failed to initialize encryption: %v", err)
	}

	authRepo := authRepoPkg.NewAuthRepository(db)
	userRepo := userRepoPkg.NewUserRepository(db)
	ledgerRepo := ledgerRepoPkg.NewLedgerRepository()
	groupRepo := groupRepoPkg.NewGroupRepository()
	groupMemberRepo := groupMemberRepoPkg.NewGroupMemberRepository()
	groupParticipantRepo := groupParticipantRepoPkg.NewGroupParticipantRepository()
	expenseRepo := expenseRepoPkg.NewExpenseRepository()
	settlementRepo := settlementRepoPkg.NewSettlementRepository()
	paymentMethodRepo := paymentMethodRepoPkg.NewPaymentMethodRepository()

	authMiddleware := middleware.NewAuthMiddleware(pasetoMaker, authRepo)

	authService := authServicePkg.NewAuthService(authRepo, pasetoMaker, sessionStore, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
	userService := userServicePkg.NewUserService(userRepo)
	groupService := groupServicePkg.NewGroupService(db, groupRepo, groupMemberRepo, groupParticipantRepo, userRepo)
	groupParticipantService := groupParticipantServicePkg.NewGroupParticipantService(db, groupRepo, groupMemberRepo, groupParticipantRepo)
	groupMemberService := groupMemberServicePkg.NewGroupMemberService(db, groupRepo, groupMemberRepo, groupParticipantRepo, userRepo)
	ledgerService := ledgerServicePkg.NewLedgerService()
	expenseService := expenseServicePkg.NewExpenseService(db, expenseRepo, ledgerRepo, ledgerService, groupRepo, groupMemberRepo, groupParticipantRepo)
	activityService := activityServicePkg.NewActivityService(db, expenseRepo, settlementRepo, groupRepo, groupMemberRepo)
	balanceService := balanceServicePkg.NewBalanceService(db, ledgerRepo, groupRepo, groupMemberRepo)
	settlementService := settlementServicePkg.NewSettlementService(db, settlementRepo, ledgerRepo, groupRepo, groupMemberRepo)
	paymentMethodService := paymentMethodServicePkg.NewPaymentMethodService(db, encryption, paymentMethodRepo)

	authHandler := authHandlerPkg.NewAuthHandler(authService)
	userHandler := userHandlerPkg.NewUserHandler(userService)
	groupHandler := groupHandlerPkg.NewGroupHandler(groupService)
	groupMemberHandler := groupMemberHandlerPkg.NewGroupMemberHandler(groupMemberService)
	groupParticipantHandler := groupParticipantHandlerPkg.NewGroupParticipantHandler(groupParticipantService)
	expenseHandler := expenseHandlerPkg.NewExpenseHandler(expenseService)
	activityHandler := activityHandlerPkg.NewActivityHandler(activityService)
	balanceHandler := balanceHandlerPkg.NewBalanceHandler(balanceService)
	settlementHandler := settlementHandlerPkg.NewSettlementHandler(settlementService)
	paymentMethodHandler := paymentMethodHandlerPkg.NewPaymentMethodHandler(paymentMethodService)

	return &Dependency{
		Config:         cfg,
		Encryption:     encryption,
		DB:             db,
		Redis:          rdb,
		PasetoMaker:    pasetoMaker,
		SessionStore:   sessionStore,
		AuthMiddleware: authMiddleware,

		AuthHandler: authHandler,

		UserHandler: userHandler,

		GroupHandler: groupHandler,

		GroupMemberHandler: groupMemberHandler,

		GroupParticipantHandler: groupParticipantHandler,

		ExpenseHandler: expenseHandler,

		ActivityHandler: activityHandler,

		BalanceHandler: balanceHandler,

		SettlementHandler: settlementHandler,

		PaymentMethodHandler: paymentMethodHandler,
	}, nil
}
