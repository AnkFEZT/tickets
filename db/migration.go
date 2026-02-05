package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"tickets/entities"
)

type bookingMade_v0 struct {
	Header entities.MessageHeader `json:"header"`

	NumberOfTickets int `json:"number_of_tickets"`

	BookingID uuid.UUID `json:"booking_id"`

	CustomerEmail string    `json:"customer_email"`
	ShowID        uuid.UUID `json:"show_id"`
}

type ticketBookingConfirmed_v0 struct {
	Header entities.MessageHeader `json:"header"`

	TicketID      string         `json:"ticket_id"`
	CustomerEmail string         `json:"customer_email"`
	Price         entities.Money `json:"price"`

	BookingID string `json:"booking_id"`
}

type ticketReceiptIssued_v0 struct {
	Header entities.MessageHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	ReceiptNumber string `json:"receipt_number"`

	IssuedAt time.Time `json:"issued_at"`
}

type ticketPrinted_v0 struct {
	Header entities.MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}

type ticketRefunded_v0 struct {
	Header entities.MessageHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}

func unmarshalDataLakeEvent[T any](event entities.DataLakeEvent) (*T, error) {
	eventInstance := new(T)

	err := json.Unmarshal(event.EventPayload, &eventInstance)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event %s: %w", event.EventName, err)
	}

	return eventInstance, nil
}

func migrateEvent(ctx context.Context, event entities.DataLakeEvent, rm OpsBookingReadModel) error {
	switch event.EventName {
	case "BookingMade_v0":
		bookingMade, err := unmarshalDataLakeEvent[bookingMade_v0](event)
		if err != nil {
			return err
		}

		return rm.OnBookingMade(ctx, &entities.BookingMade_v1{
			Header:          bookingMade.Header,
			NumberOfTickets: bookingMade.NumberOfTickets,
			BookingID:       bookingMade.BookingID,
			CustomerEmail:   bookingMade.CustomerEmail,
			ShowID:          bookingMade.ShowID,
		})
	case "TicketBookingConfirmed_v0":
		bookingConfirmedEvent, err := unmarshalDataLakeEvent[ticketBookingConfirmed_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketBookingConfirmed(ctx, &entities.TicketBookingConfirmed_v1{
			Header:        bookingConfirmedEvent.Header,
			TicketID:      bookingConfirmedEvent.TicketID,
			CustomerEmail: bookingConfirmedEvent.CustomerEmail,
			Price:         bookingConfirmedEvent.Price,
			BookingID:     bookingConfirmedEvent.BookingID,
		})
	case "TicketReceiptIssued_v0":
		receiptIssuedEvent, err := unmarshalDataLakeEvent[ticketReceiptIssued_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketReceiptIssued(ctx, &entities.TicketReceiptIssued_v1{
			Header:        receiptIssuedEvent.Header,
			TicketID:      receiptIssuedEvent.TicketID,
			ReceiptNumber: receiptIssuedEvent.ReceiptNumber,
			IssuedAt:      receiptIssuedEvent.IssuedAt,
		})
	case "TicketPrinted_v0":
		ticketPrintedEvent, err := unmarshalDataLakeEvent[ticketPrinted_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketPrinted(ctx, &entities.TicketPrinted_v1{
			Header:   ticketPrintedEvent.Header,
			TicketID: ticketPrintedEvent.TicketID,
			FileName: ticketPrintedEvent.FileName,
		})
	case "TicketRefunded_v0":
		ticketRefundedEvent, err := unmarshalDataLakeEvent[ticketRefunded_v0](event)
		if err != nil {
			return err
		}

		return rm.OnTicketRefunded(ctx, &entities.TicketRefunded_v1{
			Header:   ticketRefundedEvent.Header,
			TicketID: ticketRefundedEvent.TicketID,
		})
	default:
		// Skip unknown events (they may be events we don't need for this read model)
		slog.Info("Skipping unknown event during migration", "event_name", event.EventName)
		return nil
	}
}

func MigrateOpsReadModel(ctx context.Context, dataLake DataLake, rm OpsBookingReadModel) {
	slog.Info("Starting Ops read model migration, waiting for events table to be populated...")

	// Wait for events table to have data
	for {
		select {
		case <-ctx.Done():
			slog.Info("Migration cancelled")
			return
		default:
		}

		hasEvents, err := dataLake.HasEvents(ctx)
		if err != nil {
			slog.Error("Error checking events table", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if hasEvents {
			break
		}

		time.Sleep(1 * time.Second)
	}

	slog.Info("Events table populated, starting migration...")

	events, err := dataLake.GetEvents(ctx)
	if err != nil {
		slog.Error("Failed to get events from data lake", "error", err)
		return
	}

	slog.Info("Migrating events", "count", len(events))

	for i, event := range events {
		if err := migrateEvent(ctx, event, rm); err != nil {
			slog.Error("Failed to migrate event",
				"event_id", event.EventID,
				"event_name", event.EventName,
				"error", err,
			)
			continue
		}

		if (i+1)%100 == 0 {
			slog.Info("Migration progress", "processed", i+1, "total", len(events))
		}
	}

	slog.Info("Ops read model migration completed", "total_events", len(events))
}
