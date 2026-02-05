package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h Handler) GetOpsBookings(c echo.Context) error {
	receiptIssueDate := c.QueryParam("receipt_issue_date")

	bookings, err := h.opsBookings.FindAll(c.Request().Context(), receiptIssueDate)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, bookings)
}

func (h Handler) GetOpsBookingByID(c echo.Context) error {
	bookingID := c.Param("id")

	if _, err := uuid.Parse(bookingID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid booking ID format")
	}

	booking, err := h.opsBookings.FindByID(c.Request().Context(), bookingID)
	if err != nil {
		errMsg := err.Error()
		if len(errMsg) >= 16 && errMsg[:16] == "booking not found" {
			return echo.NewHTTPError(http.StatusNotFound, "booking not found")
		}
		return err
	}

	return c.JSON(http.StatusOK, booking)
}
