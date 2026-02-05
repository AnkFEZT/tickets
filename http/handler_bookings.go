package http

import (
	"errors"
	"fmt"
	"net/http"
	"tickets/db"
	"tickets/entities"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type PostBookTicketsRequest struct {
	ShowID          uuid.UUID `json:"show_id"`
	NumberOfTickets int       `json:"number_of_tickets"`
	CustomerEmail   string    `json:"customer_email"`
}

type PostBookTicketsResponse struct {
	BookingID uuid.UUID `json:"booking_id"`
}

func (h Handler) PostBookTickets(c echo.Context) error {
	var request PostBookTicketsRequest

	if err := c.Bind(&request); err != nil {
		return err
	}

	if request.NumberOfTickets < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "number of tickets must be greater than 0")
	}

	bookingID := uuid.New()

	booking := entities.Booking{
		BookingID:       bookingID,
		ShowID:          request.ShowID,
		NumberOfTickets: request.NumberOfTickets,
		CustomerEmail:   request.CustomerEmail,
	}

	if err := h.bookings.AddBooking(c.Request().Context(), booking); err != nil {
		if errors.Is(err, db.ErrNotEnoughSeats) {
			return echo.NewHTTPError(http.StatusBadRequest, "not enough seats available")
		}
		return fmt.Errorf("failed to add booking: %w", err)
	}

	return c.JSON(http.StatusCreated, PostBookTicketsResponse{BookingID: bookingID})
}
