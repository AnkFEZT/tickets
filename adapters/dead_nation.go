package adapters

import (
	"context"
	"fmt"
	"net/http"

	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/dead_nation"
)

type DeadNationClient struct {
	clients *clients.Clients
}

func NewDeadNationClient(clients *clients.Clients) *DeadNationClient {
	if clients == nil {
		panic("NewDeadNationClient: clients is nil")
	}

	return &DeadNationClient{clients: clients}
}

func (c DeadNationClient) BookTicket(ctx context.Context, request entities.DeadNationBooking) error {
	resp, err := c.clients.DeadNation.PostTicketBookingWithResponse(ctx, dead_nation.PostTicketBookingRequest{
		BookingId:       request.BookingID,
		EventId:         request.DeadNationEventID,
		NumberOfTickets: request.NumberOfTickets,
		CustomerAddress: request.CustomerEmail, // translation: our CustomerEmail -> their CustomerAddress
	})
	if err != nil {
		return fmt.Errorf("failed to post ticket booking: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to post ticket booking: unexpected status code %d", resp.StatusCode())
	}

	return nil
}
