package tests_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testBookingExceedsLimitReturns400(t *testing.T, fixtures *TestFixtures) {
	showID := uuid.New()
	deadNationID := uuid.New()

	createShow(t, fixtures.DB, showID, deadNationID, 5, "Test Show")

	statusCode, body := bookTickets(t, showID, 10, "test@example.com")

	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Contains(t, string(body), "not enough seats available")

	assert.Equal(t, 0, getBookedTicketsCount(t, fixtures.DB, showID))
}

func testMultipleBookingsUntilExhausted(t *testing.T, fixtures *TestFixtures) {
	showID := uuid.New()
	deadNationID := uuid.New()

	createShow(t, fixtures.DB, showID, deadNationID, 10, "Test Show")

	statusCode, _ := bookTickets(t, showID, 5, "test1@example.com")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, 5, getBookedTicketsCount(t, fixtures.DB, showID))

	statusCode, _ = bookTickets(t, showID, 5, "test2@example.com")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, 10, getBookedTicketsCount(t, fixtures.DB, showID))

	statusCode, body := bookTickets(t, showID, 1, "test3@example.com")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Contains(t, string(body), "not enough seats available")

	assert.Equal(t, 10, getBookedTicketsCount(t, fixtures.DB, showID))
}
