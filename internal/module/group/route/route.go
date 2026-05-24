package route

import (
	expenseHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/expense/handler"
	"github.com/KejarBahasa/kejarbill-api/internal/module/group/handler"
	groupMemberHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/handler"
	groupParticipantHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/handler"
	balanceHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/ledger/handler"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"

	"github.com/gofiber/fiber/v3"
)

func GroupRoute(
	api fiber.Router,
	authMiddleware *middleware.AuthMiddleware,
	groupHandler *handler.GroupHandler,
	groupMemberHandler *groupMemberHandlerPkg.GroupMemberHandler,
	groupParticipantHandler *groupParticipantHandlerPkg.GroupParticipantHandler,
	balanceHandler *balanceHandlerPkg.BalanceHandler,
	expenseHandler *expenseHandlerPkg.ExpenseHandler,
) {
	group := api.Group("/groups", authMiddleware.Protected)

	group.Post("/", groupHandler.Create)

	group.Post("/:group_id/members", groupMemberHandler.AddMember)
	group.Post("/:group_id/members/bulk", groupMemberHandler.AddMemberBulk)

	group.Get("/:group_id/participants", groupParticipantHandler.GetByGroupID)
	group.Post("/:group_id/participants/guests", groupParticipantHandler.CreateGuestParticipants)

	group.Get("/:group_id/balances", balanceHandler.GetGroupBalances)

	group.Get("/:group_id/expenses", expenseHandler.GetByGroupID)
}
