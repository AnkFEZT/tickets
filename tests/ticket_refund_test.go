package tests_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testTicketRefundVoidsReceipt(t *testing.T, fixtures *TestFixtures) {
	ticketID := uuid.NewString()

	sendTicketRefund(t, ticketID)

	assertReceiptForTicketVoided(t, fixtures.ReceiptsService, ticketID)
	assertPaymentRefunded(t, fixtures.PaymentsService, ticketID)
}

func testTicketRefundIdempotency(t *testing.T, fixtures *TestFixtures) {
	ticketID := uuid.NewString()

	sendTicketRefund(t, ticketID)
	sendTicketRefund(t, ticketID)
	sendTicketRefund(t, ticketID)

	_, found := fixtures.ReceiptsService.FindVoidedReceipt(ticketID)
	assert.True(t, found, "receipt should be voided exactly once (idempotency)")

	_, found = fixtures.PaymentsService.FindRefundedPayment(ticketID)
	assert.True(t, found, "payment should be refunded exactly once (idempotency)")
}
