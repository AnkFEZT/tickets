package tests_test

import (
	"testing"

	"tickets/entities"
	ticketsHttp "tickets/http"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testTicketBooking(t *testing.T, fixtures *TestFixtures) {
	ticket := ticketsHttp.TicketStatusRequest{
		TicketID: uuid.NewString(),
		Status:   "confirmed",
		Price: entities.Money{
			Amount:   "50.30",
			Currency: "GBP",
		},
		CustomerEmail: "email@example.com",
	}

	idempotencyKey := uuid.NewString()
	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{ticket},
	}, idempotencyKey)

	assertTicketStoredInRepository(t, fixtures.DB, ticket)
	assertTicketPrinted(t, fixtures.FileAPI, ticket.TicketID)
	assertReceiptForTicketIssued(t, fixtures.ReceiptsService, ticket)
	assertRowToSheetAdded(t, fixtures.SpreadsheetsAPI, ticket, "tickets-to-print")
}

func testTicketBookingIdempotency(t *testing.T, fixtures *TestFixtures) {
	ticket := ticketsHttp.TicketStatusRequest{
		TicketID: uuid.NewString(),
		Status:   "confirmed",
		Price: entities.Money{
			Amount:   "75.00",
			Currency: "EUR",
		},
		CustomerEmail: "idempotent@example.com",
	}

	idempotencyKey := uuid.NewString()

	// Send the same request 3 times with the same idempotency key
	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{ticket},
	}, idempotencyKey)

	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{ticket},
	}, idempotencyKey)

	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{ticket},
	}, idempotencyKey)

	assertTicketStoredInRepository(t, fixtures.DB, ticket)
	assertTicketPrinted(t, fixtures.FileAPI, ticket.TicketID)

	_, found := fixtures.ReceiptsService.FindIssuedReceipt(ticket.TicketID)
	assert.True(t, found, "receipt should be issued exactly once (idempotency)")

	_, found = fixtures.SpreadsheetsAPI.FindRowByTicketID("tickets-to-print", ticket.TicketID)
	assert.True(t, found, "row should be added to sheet exactly once (idempotency)")

	_, found = fixtures.FileAPI.FindPutCallByFileID(ticket.TicketID + "-ticket.html")
	assert.True(t, found, "file should be uploaded exactly once (idempotency)")
}
