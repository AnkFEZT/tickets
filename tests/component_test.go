package tests_test

import (
	"testing"
)

func TestComponent(t *testing.T) {
	fixtures := SetupComponentTest(t)

	t.Run("ticket_booking", func(t *testing.T) {
		testTicketBooking(t, fixtures)
	})

	t.Run("ticket_cancellation", func(t *testing.T) {
		testTicketCancellation(t, fixtures)
	})

	t.Run("ticket_booking_idempotency", func(t *testing.T) {
		testTicketBookingIdempotency(t, fixtures)
	})

	t.Run("booking_exceeds_limit_returns_400", func(t *testing.T) {
		testBookingExceedsLimitReturns400(t, fixtures)
	})

	t.Run("multiple_bookings_until_exhausted", func(t *testing.T) {
		testMultipleBookingsUntilExhausted(t, fixtures)
	})

	t.Run("ticket_refund_voids_receipt", func(t *testing.T) {
		testTicketRefundVoidsReceipt(t, fixtures)
	})

	t.Run("ticket_refund_idempotency", func(t *testing.T) {
		testTicketRefundIdempotency(t, fixtures)
	})
}
