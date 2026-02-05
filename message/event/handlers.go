package event

import (
	"context"
	"fmt"
	"log/slog"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type FileAPI interface {
	UploadFile(ctx context.Context, fileID, fileContent string) error
}

type ReceiptsService interface {
	IssueReceipt(context.Context, entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error)
}

type TicketRepository interface {
	Add(ctx context.Context, ticket entities.Ticket) error
	Remove(ctx context.Context, ticketID string) error
	FindAll(context.Context) ([]entities.Ticket, error)
}

type ShowRepository interface {
	ShowByID(ctx context.Context, showID uuid.UUID) (entities.Show, error)
}

type DeadNationAPI interface {
	BookTicket(ctx context.Context, request entities.DeadNationBooking) error
}

type Handlers struct {
	fileAPI         FileAPI
	spreadsheetsAPI SpreadsheetsAPI
	receiptsService ReceiptsService
	tickets         TicketRepository
	shows           ShowRepository
	deadnation      DeadNationAPI
	eventBus        *cqrs.EventBus
}

func NewHandlers(
	fileAPI FileAPI,
	spreadsheetsAPI SpreadsheetsAPI,
	receiptsService ReceiptsService,
	tickets TicketRepository,
	shows ShowRepository,
	deadnation DeadNationAPI,
	eventBus *cqrs.EventBus,
) Handlers {
	return Handlers{fileAPI, spreadsheetsAPI, receiptsService, tickets, shows, deadnation, eventBus}
}

func (h Handlers) IssueReceipt(ctx context.Context, e *entities.TicketBookingConfirmed_v1) error {
	slog.Info("issuing receipt")

	request := entities.IssueReceiptRequest{
		IdempotencyKey: e.Header.IdempotencyKey,
		TicketID:       e.TicketID,
		Price:          e.Price,
	}

	response, err := h.receiptsService.IssueReceipt(ctx, request)
	if err != nil {
		return err
	}

	receiptIssued := entities.TicketReceiptIssued_v1{
		Header:        entities.NewMessageHeader(),
		TicketID:      e.TicketID,
		ReceiptNumber: response.ReceiptNumber,
		IssuedAt:      response.IssuedAt,
	}

	return h.eventBus.Publish(ctx, receiptIssued)
}

func (h Handlers) StoreTicket(ctx context.Context, e *entities.TicketBookingConfirmed_v1) error {
	slog.Info("storing ticket", "ticket_id", e.TicketID)

	ticket := entities.Ticket{
		TicketID:      e.TicketID,
		Price:         e.Price,
		CustomerEmail: e.CustomerEmail,
	}

	return h.tickets.Add(ctx, ticket)
}

func (h Handlers) AppendToTracker(ctx context.Context, e *entities.TicketBookingConfirmed_v1) error {
	slog.Info("Appending ticket to the tracker")

	row := []string{
		e.TicketID,
		e.CustomerEmail,
		e.Price.Amount,
		e.Price.Currency,
	}

	return h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-print", row)
}

func (h Handlers) TicketRefundToSheet(ctx context.Context, e *entities.TicketBookingCanceled_v1) error {
	slog.Info("Adding ticket refund to sheet")

	row := []string{
		e.TicketID,
		e.CustomerEmail,
		e.Price.Amount,
		e.Price.Currency,
	}

	return h.spreadsheetsAPI.AppendRow(ctx, "tickets-to-refund", row)
}

func (h Handlers) RemoveCanceledTicket(ctx context.Context, e *entities.TicketBookingCanceled_v1) error {
	slog.Info("removing ticket", "ticket_id", e.TicketID)

	return h.tickets.Remove(ctx, e.TicketID)
}

func (h Handlers) PrintTicket(ctx context.Context, e *entities.TicketBookingConfirmed_v1) error {
	slog.Info("creating ticket file", "ticket_id", e.TicketID)

	fileID := e.TicketID + "-ticket.html"
	fileContent := fmt.Sprintf("<html><body>Ticket ID: %s, Price: %s %s</body></html>", e.TicketID, e.Price.Amount, e.Price.Currency)

	if err := h.fileAPI.UploadFile(ctx, fileID, fileContent); err != nil {
		return err
	}

	ticketPrinted := entities.TicketPrinted_v1{
		Header:   entities.NewMessageHeader(),
		TicketID: e.TicketID,
		FileName: fileID,
	}

	return h.eventBus.Publish(ctx, ticketPrinted)
}

func (h Handlers) BookPlaceInDeadNation(ctx context.Context, e *entities.BookingMade_v1) error {
	slog.Info("booking ticket on Dead Nation", "booking_id", e.BookingID)

	show, err := h.shows.ShowByID(ctx, e.ShowID)
	if err != nil {
		return fmt.Errorf("failed to get show: %w", err)
	}

	request := entities.DeadNationBooking{
		BookingID:         e.BookingID,
		DeadNationEventID: show.DeadNationID,
		NumberOfTickets:   e.NumberOfTickets,
		CustomerEmail:     e.CustomerEmail,
	}

	return h.deadnation.BookTicket(ctx, request)
}
