package http

import (
	"net/http"
	"tickets/entities"

	"github.com/labstack/echo/v4"
)

func (h Handler) PutTicketRefund(c echo.Context) error {
	ticketID := c.Param("ticket_id")

	if ticketID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ticket_id is required")
	}

	// Use ticketID as idempotency key to ensure refund is idempotent
	idempotencyKey := "refund-" + ticketID

	cmd := &entities.RefundTicket{
		Header:   entities.NewMessageHeaderWithIdempotencyKey(idempotencyKey),
		TicketID: ticketID,
	}

	if err := h.commandBus.Send(c.Request().Context(), cmd); err != nil {
		return err
	}

	return c.NoContent(http.StatusAccepted)
}
