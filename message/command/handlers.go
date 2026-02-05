package command

import (
	"context"
	"log/slog"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type ReceiptsService interface {
	VoidReceipt(ctx context.Context, ticketID, idempotencyKey string) error
}

type PaymentsService interface {
	RefundPayment(ctx context.Context, ticketID, idempotencyKey string) error
}

type Handlers struct {
	receiptsService ReceiptsService
	paymentsService PaymentsService
	events          *cqrs.EventBus
}

func NewHandlers(receiptsService ReceiptsService, paymentsService PaymentsService, eventBus *cqrs.EventBus) Handlers {
	return Handlers{
		receiptsService: receiptsService,
		paymentsService: paymentsService,
		events:          eventBus,
	}
}

func (h Handlers) RefundTicketHandler(ctx context.Context, cmd *entities.RefundTicket) error {
	slog.Info("refunding ticket", "ticket_id", cmd.TicketID)

	err := h.receiptsService.VoidReceipt(ctx, cmd.TicketID, cmd.Header.IdempotencyKey)
	if err != nil {
		return err
	}

	err = h.paymentsService.RefundPayment(ctx, cmd.TicketID, cmd.Header.IdempotencyKey)
	if err != nil {
		return err
	}

	ticketRefunded := entities.TicketRefunded_v1{
		Header:   cmd.Header,
		TicketID: cmd.TicketID,
	}

	if err := h.events.Publish(ctx, ticketRefunded); err != nil {
		return err
	}

	slog.Info("ticket refunded successfully", "ticket_id", cmd.TicketID)
	return nil
}
