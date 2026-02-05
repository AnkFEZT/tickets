package entities

type AppendToTrackerPayload struct {
	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

type IssueReceiptPayload struct {
	TicketID string `json:"ticket_id"`
	Price    Money  `json:"price"`
}
