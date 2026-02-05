package db_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tickets/db"
	"tickets/entities"
)

func TestBookingsRepository_AddBooking_seats_limit(t *testing.T) {
	ctx := context.Background()

	testDB := getDBTest()

	err := db.InitializeDatabaseSchema(testDB)
	require.NoError(t, err)

	bookings := db.NewBookingRepository(testDB)
	shows := db.NewShowRepository(testDB)

	t.Run("overbooking", func(t *testing.T) {
		showID := uuid.New()

		err := shows.AddShow(ctx, entities.Show{
			ShowID:          showID,
			DeadNationID:    uuid.New(),
			NumberOfTickets: 2,
			StartTime:       time.Now().Add(time.Hour),
			Title:           "Example title",
			Venue:           "Exmaple vanue",
		})
		require.NoError(t, err)

		err = bookings.AddBooking(ctx, entities.Booking{
			BookingID:       uuid.New(),
			ShowID:          showID,
			NumberOfTickets: 2,
			CustomerEmail:   "foo@bar.com",
		})
		require.NoError(t, err)

		err = bookings.AddBooking(ctx, entities.Booking{
			BookingID:       uuid.New(),
			ShowID:          showID,
			NumberOfTickets: 2,
			CustomerEmail:   "foo@bar.com",
		})
		require.ErrorIs(t, err, db.ErrNotEnoughSeats)
	})

	t.Run("parallel_overbooking", func(t *testing.T) {
		showID := uuid.New()

		workersCount := 50
		workersErrs := make(chan error, workersCount)

		unlock := make(chan struct{})

		err := shows.AddShow(ctx, entities.Show{
			ShowID:          showID,
			DeadNationID:    uuid.New(),
			NumberOfTickets: 2,
			StartTime:       time.Now().Add(time.Hour),
			Title:           "Example title",
			Venue:           "Exmaple vanue",
		})
		require.NoError(t, err)

		wg := sync.WaitGroup{}
		wg.Add(workersCount)

		for i := 0; i < workersCount; i++ {
			go func() {
				defer wg.Done()

				// we are synchronizing goroutines to make sure that chance of overbooking is as high as possible
				<-unlock
				err = bookings.AddBooking(ctx, entities.Booking{
					BookingID:       uuid.New(),
					ShowID:          showID,
					NumberOfTickets: 2,
					CustomerEmail:   "foo@bar.com",
				})
				workersErrs <- err
			}()
		}
		close(unlock)

		wg.Wait()
		close(workersErrs)

		failedWorkers := 0
		succeededWorkers := 0
		errors := []error{}

		for err := range workersErrs {
			if err != nil {
				failedWorkers++
				errors = append(errors, err)
			} else {
				succeededWorkers++
			}
		}

		assert.Equal(t, 1, succeededWorkers)
		assert.Equal(t, workersCount-1, failedWorkers)

		if succeededWorkers == 0 {
			// all workers failed, let's print error
			for _, err := range errors {
				t.Log("error:", err)
			}
		}
	})
}
