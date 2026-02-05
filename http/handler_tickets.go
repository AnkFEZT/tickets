package http

import (
	"fmt"
	"net/http"
	"tickets/entities"

	"github.com/labstack/echo/v4"
)

type TicketsStatusRequest struct {
	Tickets []TicketStatusRequest `json:"tickets"`
}

type TicketStatusRequest struct {
	TicketID      string         `json:"ticket_id"`
	BookingID     string         `json:"booking_id"`
	Status        string         `json:"status"`
	Price         entities.Money `json:"price"`
	CustomerEmail string         `json:"customer_email"`
}

func (h Handler) PostTicketsStatus(c echo.Context) error {
	idempotencyKey := c.Request().Header.Get("Idempotency-Key")

	if idempotencyKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Idempotency-Key header is required")
	}

	var request TicketsStatusRequest

	if err := c.Bind(&request); err != nil {
		return err
	}

	for _, t := range request.Tickets {
		if t.Status == "confirmed" {
			e := &entities.TicketBookingConfirmed_v1{
				Header:        entities.NewMessageHeaderWithIdempotencyKey(idempotencyKey),
				BookingID:     t.BookingID,
				TicketID:      t.TicketID,
				CustomerEmail: t.CustomerEmail,
				Price:         t.Price,
			}

			if err := h.eventBus.Publish(c.Request().Context(), e); err != nil {
				return err
			}

			continue
		}

		if t.Status == "canceled" {
			e := &entities.TicketBookingCanceled_v1{
				Header:        entities.NewMessageHeaderWithIdempotencyKey(idempotencyKey),
				TicketID:      t.TicketID,
				CustomerEmail: t.CustomerEmail,
				Price:         t.Price,
			}

			if err := h.eventBus.Publish(c.Request().Context(), e); err != nil {
				return err
			}

			continue
		}

		return fmt.Errorf("unknown ticket status: %s", t.Status)
	}

	return c.NoContent(http.StatusOK)
}

func (h Handler) GetTickets(c echo.Context) error {
	tickets, err := h.tickets.FindAll(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tickets)
}
