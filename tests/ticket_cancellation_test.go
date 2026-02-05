package tests_test

import (
	"testing"

	"tickets/entities"
	ticketsHttp "tickets/http"

	"github.com/google/uuid"
)

func testTicketCancellation(t *testing.T, fixtures *TestFixtures) {
	canceledTicket := ticketsHttp.TicketStatusRequest{
		TicketID: uuid.NewString(),
		Status:   "canceled",
		Price: entities.Money{
			Amount:   "100.00",
			Currency: "USD",
		},
		CustomerEmail: "canceled@example.com",
	}

	idempotencyKey := uuid.NewString()
	sendTicketsStatus(t, ticketsHttp.TicketsStatusRequest{
		Tickets: []ticketsHttp.TicketStatusRequest{canceledTicket},
	}, idempotencyKey)

	assertRowToSheetAdded(t, fixtures.SpreadsheetsAPI, canceledTicket, "tickets-to-refund")
}
