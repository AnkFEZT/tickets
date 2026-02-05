package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/payments"
)

type PaymentsServiceClient struct {
	// we are not mocking this client: it's pointless to use interface here
	clients *clients.Clients
}

func NewPaymentsServiceClient(clients *clients.Clients) *PaymentsServiceClient {
	if clients == nil {
		panic("NewPaymentsServiceClient: clients is nil")
	}

	return &PaymentsServiceClient{clients: clients}
}

func (c PaymentsServiceClient) RefundPayment(ctx context.Context, ticketID, idempotencyKey string) error {
	resp, err := c.clients.Payments.PutRefundsWithResponse(ctx, payments.PaymentRefundRequest{
		// we use TicketID as the payment reference
		PaymentReference: ticketID,
		Reason:           "customer requested refund",
		DeduplicationId:  &idempotencyKey,
	})

	if err != nil {
		return fmt.Errorf("failed to post refund: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusCreated:
		return nil
	default:
		return fmt.Errorf("unexpected status code for POST payments-api/refunds: %d", resp.StatusCode())
	}
}
