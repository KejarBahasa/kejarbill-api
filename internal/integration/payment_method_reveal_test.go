package integration_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	paymentMethodConstants "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/constants"
	paymentMethodRepoPkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/repository"
	paymentMethodServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/service"
	settlementHandlerPkg "github.com/KejarBahasa/kejarbill-api/internal/module/settlement/handler"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (f *recipientMethodFixture) addEncryptedMethod(t *testing.T, visibility, accountNumber string) string {
	t.Helper()
	methodID := uuid.NewString()
	encryption := newTestEncryption(t)
	ciphertext, err := encryption.Encrypt(accountNumber)
	if err != nil {
		t.Fatalf("encrypt account number: %v", err)
	}
	mustExec(t, context.Background(), `
		INSERT INTO payment_methods (id, user_id, method_type, provider_name, account_number, visibility, status)
		VALUES ($1, $2, 'bank', 'Reveal Bank', $3, $4, 'active')
	`, methodID, f.recipientID, ciphertext, visibility)
	return methodID
}

func TestPaymentMethodRevealAndMasking(t *testing.T) {
	f := newRecipientMethodFixture(t)
	methodID := f.addEncryptedMethod(t, paymentMethodConstants.VisibilityGroupMember, "1234567890")

	methods, err := f.service.GetRecipientPaymentMethods(context.Background(), f.requesterID, f.groupID, f.recipientParticipant)
	if err != nil {
		t.Fatalf("GetRecipientPaymentMethods() error = %v", err)
	}
	var masked string
	for _, method := range methods {
		if method.ID == methodID && method.MaskedAccountNumber != nil {
			masked = *method.MaskedAccountNumber
		}
	}
	if masked != utils.MaskAccountNumber("1234567890") {
		t.Fatalf("masked account number = %q, want %q", masked, utils.MaskAccountNumber("1234567890"))
	}

	revealed, err := f.service.RevealRecipientPaymentMethod(context.Background(), f.requesterID, f.groupID, f.recipientParticipant, methodID)
	if err != nil {
		t.Fatalf("RevealRecipientPaymentMethod() error = %v", err)
	}
	if revealed != "1234567890" {
		t.Fatalf("revealed account number = %q, want plaintext", revealed)
	}
}

func TestPaymentMethodRevealDebtorOnlyRequiresOutstandingDebt(t *testing.T) {
	f := newRecipientMethodFixture(t)
	methodID := f.addEncryptedMethod(t, paymentMethodConstants.VisibilityDebtorOnly, "9876543210")

	_, err := f.service.RevealRecipientPaymentMethod(context.Background(), f.requesterID, f.groupID, f.recipientParticipant, methodID)
	if !errors.Is(err, paymentMethodConstants.ErrPaymentMethodNotFound) {
		t.Fatalf("error before debt = %v, want not found", err)
	}

	f.makeRequesterOwe(t)
	revealed, err := f.service.RevealRecipientPaymentMethod(context.Background(), f.requesterID, f.groupID, f.recipientParticipant, methodID)
	if err != nil {
		t.Fatalf("RevealRecipientPaymentMethod() after debt error = %v", err)
	}
	if revealed != "9876543210" {
		t.Fatalf("revealed account number = %q, want plaintext", revealed)
	}
}

func TestPaymentMethodRevealRejectsUnauthorizedAndMissingMethods(t *testing.T) {
	f := newRecipientMethodFixture(t)
	methodID := f.addEncryptedMethod(t, paymentMethodConstants.VisibilityGroupMember, "1111222233")

	_, err := f.service.RevealRecipientPaymentMethod(context.Background(), uuid.NewString(), f.groupID, f.recipientParticipant, methodID)
	if !errors.Is(err, expenseConstants.ErrForbiddenGroupAccess) {
		t.Fatalf("unauthorized error = %v, want forbidden group access", err)
	}
	_, err = f.service.RevealRecipientPaymentMethod(context.Background(), f.requesterID, f.groupID, f.recipientParticipant, uuid.NewString())
	if !errors.Is(err, paymentMethodConstants.ErrPaymentMethodNotFound) {
		t.Fatalf("missing method error = %v, want not found", err)
	}
}

func TestPaymentMethodRevealDecryptFailureIsNotReturned(t *testing.T) {
	f := newRecipientMethodFixture(t)
	methodID := uuid.NewString()
	mustExec(t, context.Background(), `
		INSERT INTO payment_methods (id, user_id, method_type, provider_name, account_number, visibility, status)
		VALUES ($1, $2, 'bank', 'Broken Bank', 'not-valid-ciphertext', 'group_members', 'active')
	`, methodID, f.recipientID)

	_, err := f.service.RevealRecipientPaymentMethod(context.Background(), f.requesterID, f.groupID, f.recipientParticipant, methodID)
	if err == nil {
		t.Fatal("expected decrypt error")
	}
}

func TestPaymentMethodRevealSetsNoStoreHeaders(t *testing.T) {
	f := newRecipientMethodFixture(t)
	methodID := f.addEncryptedMethod(t, paymentMethodConstants.VisibilityGroupMember, "1234567890")
	handler := settlementHandlerPkg.NewSettlementHandler(f.service)
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(security.ContextUserID, f.requesterID)
		return c.Next()
	})
	app.Post("/v1/groups/:group_id/participants/:participant_id/payment-methods/:payment_method_id/reveal", handler.RevealRecipientPaymentMethod)

	request := httptest.NewRequest(http.MethodPost, "/v1/groups/"+f.groupID+"/participants/"+f.recipientParticipant+"/payment-methods/"+methodID+"/reveal", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	if response.Header.Get("Cache-Control") != "no-store, private" {
		t.Fatalf("cache-control = %q, want no-store, private", response.Header.Get("Cache-Control"))
	}
	if response.Header.Get("Pragma") != "no-cache" {
		t.Fatalf("pragma = %q, want no-cache", response.Header.Get("Pragma"))
	}
}

func TestPaymentMethodListDoesNotExposePlaintext(t *testing.T) {
	f := newRecipientMethodFixture(t)
	f.addEncryptedMethod(t, paymentMethodConstants.VisibilityGroupMember, "1234567890")
	service := paymentMethodServicePkg.NewPaymentMethodService(
		integrationPool,
		newTestEncryption(t),
		paymentMethodRepoPkg.NewPaymentMethodRepository(),
	)

	methods, err := service.FindMyPaymentMethods(context.Background(), f.recipientID)
	if err != nil {
		t.Fatalf("FindMyPaymentMethods() error = %v", err)
	}
	for _, method := range methods {
		if method.MaskedAccountNumber != nil && *method.MaskedAccountNumber == "1234567890" {
			t.Fatal("list response exposed plaintext account number")
		}
	}
}
