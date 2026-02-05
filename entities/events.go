package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
)

type MessageHeader struct {
	ID             string    `json:"id"`
	PublishedAt    time.Time `json:"published_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

func NewMessageHeaderWithIdempotencyKey(idempotencyKey string) MessageHeader {
	return MessageHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: idempotencyKey,
	}
}

func NewMessageHeader() MessageHeader {
	return MessageHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: shortuuid.New(),
	}
}

type TicketBookingConfirmed_v1 struct {
	Header        MessageHeader `json:"header"`
	BookingID     string        `json:"booking_id"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         Money         `json:"price"`
}

type TicketBookingCanceled_v1 struct {
	Header        MessageHeader `json:"header"`
	TicketID      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         Money         `json:"price"`
}

type TicketPrinted_v1 struct {
	Header   MessageHeader `json:"header"`
	TicketID string        `json:"ticket_id"`
	FileName string        `json:"file_name"`
}

type BookingMade_v1 struct {
	Header          MessageHeader `json:"header"`
	NumberOfTickets int           `json:"number_of_tickets"`
	BookingID       uuid.UUID     `json:"booking_id"`
	CustomerEmail   string        `json:"customer_email"`
	ShowID          uuid.UUID     `json:"show_id"`
}

type TicketReceiptIssued_v1 struct {
	Header        MessageHeader `json:"header"`
	TicketID      string        `json:"ticket_id"`
	ReceiptNumber string        `json:"receipt_number"`
	IssuedAt      time.Time     `json:"issued_at"`
}

type TicketRefunded_v1 struct {
	Header   MessageHeader `json:"header"`
	TicketID string        `json:"ticket_id"`
}

type DataLakeEvent struct {
	EventID      string    `db:"event_id"`
	PublishedAt  time.Time `db:"published_at"`
	EventName    string    `db:"event_name"`
	EventPayload []byte    `db:"event_payload"`
}
