package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lithammer/shortuuid/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tickets/adapters"
	"tickets/db"
	ticketsHttp "tickets/http"

	_ "github.com/lib/pq"
)

func waitForHttpServer(t *testing.T) {
	t.Helper()

	condition := func(t *assert.CollectT) {
		resp, err := http.Get("http://localhost:8080/health")
		if !assert.NoError(t, err) {
			return
		}
		defer resp.Body.Close()

		if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
			return
		}
	}

	require.EventuallyWithT(t, condition, time.Second*10, time.Millisecond*50)
}

func sendTicketsStatus(t *testing.T, req ticketsHttp.TicketsStatusRequest, idempotencyKey string) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := shortuuid.New()

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func sendTicketRefund(t *testing.T, ticketID string) {
	t.Helper()

	httpReq, err := http.NewRequest(
		http.MethodPut,
		"http://localhost:8080/ticket-refund/"+ticketID,
		nil,
	)
	require.NoError(t, err)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	resp.Body.Close()
}

func assertReceiptForTicketIssued(t *testing.T, receiptsService *adapters.ReceiptsServiceStub, ticket ticketsHttp.TicketStatusRequest) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			issuedReceipts := receiptsService.IssuedReceiptsCount()
			parentT.Log("issued receipts", issuedReceipts)

			if !assert.Greater(t, issuedReceipts, 0, "no receipts issued") {
				return
			}

			receipt, ok := receiptsService.FindIssuedReceipt(ticket.TicketID)
			if !assert.True(t, ok, "receipt for ticket %s not found", ticket.TicketID) {
				return
			}

			assert.Equal(t, ticket.TicketID, receipt.TicketID)
			assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
			assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
		},
		10*time.Second,
		100*time.Millisecond,
	)
}

func assertReceiptForTicketVoided(t *testing.T, receiptsService *adapters.ReceiptsServiceStub, ticketID string) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			voidedReceipts := receiptsService.VoidedReceiptsCount()
			parentT.Log("voided receipts", voidedReceipts)

			if !assert.Greater(t, voidedReceipts, 0, "no receipts voided") {
				return
			}

			voidedReceipt, ok := receiptsService.FindVoidedReceipt(ticketID)
			if !assert.True(t, ok, "voided receipt for ticket %s not found", ticketID) {
				return
			}

			assert.Equal(t, ticketID, voidedReceipt.TicketID)
			assert.NotEmpty(t, voidedReceipt.IdempotencyKey, "idempotency key should be set")
		},
		10*time.Second,
		100*time.Millisecond,
	)
}

func assertRowToSheetAdded(t *testing.T, spreadsheetsAPI *adapters.SpreadsheetsAPIStub, ticket ticketsHttp.TicketStatusRequest, sheetName string) bool {
	t.Helper()

	condition := func(t *assert.CollectT) {
		if !assert.True(t, spreadsheetsAPI.HasSheet(sheetName), "sheet %s not found", sheetName) {
			return
		}

		ticketRow, ok := spreadsheetsAPI.FindRowByTicketID(sheetName, ticket.TicketID)
		if !assert.True(t, ok, "ticket row not found in sheet %s", sheetName) {
			return
		}

		expectedRow := []string{
			ticket.TicketID,
			ticket.CustomerEmail,
			ticket.Price.Amount,
			ticket.Price.Currency,
		}

		assert.Equal(t, expectedRow, ticketRow)
	}

	return assert.EventuallyWithT(t, condition, 10*time.Second, 100*time.Millisecond)
}

func assertTicketStoredInRepository(t *testing.T, testDB *sqlx.DB, ticket ticketsHttp.TicketStatusRequest) {
	t.Helper()

	tickets := db.NewTicketRepository(testDB)

	condition := func(t *assert.CollectT) {
		foundTickets, err := tickets.FindAll(context.Background())
		if !assert.NoError(t, err, "failed to find tickets") {
			return
		}

		for _, foundTicket := range foundTickets {
			if foundTicket.TicketID == ticket.TicketID {
				return // ticket found
			}
		}

		t.Errorf("ticket with ID %s not found in repository", ticket.TicketID)
	}

	assert.EventuallyWithT(t, condition, 10*time.Second, 100*time.Millisecond)
}

func assertTicketPrinted(t *testing.T, fileAPI *adapters.FilesAPIStub, ticketID string) {
	t.Helper()

	condition := func(t *assert.CollectT) {
		calls := fileAPI.PutCallsCount()
		if !assert.Greater(t, calls, 0, "no files uploaded") {
			return
		}

		expectedFileID := ticketID + "-ticket.html"

		call, ok := fileAPI.FindPutCallByFileID(expectedFileID)
		if !assert.True(t, ok, "ticket file with ID %s not found", expectedFileID) {
			return
		}

		assert.Contains(t, call.FileContent, ticketID, "file content should contain ticket ID")
	}

	assert.EventuallyWithT(t, condition, 10*time.Second, 100*time.Millisecond)
}

func assertDeadNationBookingMade(t *testing.T, deadNation *adapters.DeadNationStub, bookingID, deadNationEventID string, numberOfTickets int, customerEmail string) {
	t.Helper()

	parsedBookingID, err := uuid.Parse(bookingID)
	require.NoError(t, err, "invalid booking ID")

	parsedDeadNationEventID, err := uuid.Parse(deadNationEventID)
	require.NoError(t, err, "invalid dead nation event ID")

	condition := func(t *assert.CollectT) {
		bookings := deadNation.BookingsCount()
		if !assert.Greater(t, bookings, 0, "no bookings made to Dead Nation") {
			return
		}

		booking, ok := deadNation.FindBookingByBookingID(bookingID)
		if !assert.True(t, ok, "booking with ID %s not found in Dead Nation stub", bookingID) {
			return
		}

		assert.Equal(t, parsedBookingID, booking.BookingID, "booking ID should match")
		assert.Equal(t, parsedDeadNationEventID, booking.DeadNationEventID, "dead nation event ID should match")
		assert.Equal(t, numberOfTickets, booking.NumberOfTickets, "number of tickets should match")
		assert.Equal(t, customerEmail, booking.CustomerEmail, "customer email should match")
	}

	assert.EventuallyWithT(t, condition, 10*time.Second, 100*time.Millisecond)
}

func createShow(t *testing.T, db *sqlx.DB, showID uuid.UUID, deadNationID uuid.UUID, numberOfTickets int, title string) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO shows (show_id, dead_nation_id, number_of_tickets, start_time, title, venue)
		VALUES ($1, $2, $3, NOW(), $4, 'Test Venue')`,
		showID, deadNationID, numberOfTickets, title,
	)
	require.NoError(t, err, "failed to create show")
}

func bookTickets(t *testing.T, showID uuid.UUID, numberOfTickets int, customerEmail string) (int, []byte) {
	t.Helper()

	req := ticketsHttp.PostBookTicketsRequest{
		ShowID:          showID,
		NumberOfTickets: numberOfTickets,
		CustomerEmail:   customerEmail,
	}

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/book-tickets",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp.StatusCode, body
}

func assertBookingCreated(t *testing.T, db *sqlx.DB, bookingID uuid.UUID, showID uuid.UUID, numberOfTickets int, customerEmail string) {
	t.Helper()

	countQuery := `
		SELECT COUNT(*) FROM bookings
		WHERE booking_id = $1 AND show_id = $2 AND number_of_tickets = $3 AND customer_email = $4
	`
	var count int
	err := db.Get(&count, countQuery, bookingID, showID, numberOfTickets, customerEmail)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "booking should exist in database")
}

func getBookedTicketsCount(t *testing.T, db *sqlx.DB, showID uuid.UUID) int {
	t.Helper()

	var count int
	err := db.Get(&count, `SELECT COALESCE(SUM(number_of_tickets), 0) FROM bookings WHERE show_id = $1`, showID)
	require.NoError(t, err)
	return count
}

func assertPaymentRefunded(t *testing.T, paymentsService *adapters.PaymentsServiceStub, ticketID string) {
	t.Helper()

	parentT := t

	condition := func(t *assert.CollectT) {
		refundedPayments := paymentsService.RefundedPaymentsCount()
		parentT.Log("refunded payments", refundedPayments)

		assert.Greater(t, refundedPayments, 0, "no payments refunded")
	}

	assert.EventuallyWithT(t, condition, 10*time.Second, 100*time.Millisecond)

	refundedPayment, ok := paymentsService.FindRefundedPayment(ticketID)
	require.Truef(t, ok, "refunded payment for ticket %s not found", ticketID)
	assert.Equal(t, ticketID, refundedPayment.TicketID)
	assert.NotEmpty(t, refundedPayment.IdempotencyKey, "idempotency key should be set")
}
