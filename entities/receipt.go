package entities

import "time"

type IssueReceiptRequest struct {
	IdempotencyKey string
	TicketID       string
	Price          Money
}

type IssueReceiptResponse struct {
	ReceiptNumber string
	IssuedAt      time.Time
}
