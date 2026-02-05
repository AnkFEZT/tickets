package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"tickets/entities"
	"tickets/message/event"
	"tickets/message/outbox"
)

var ErrNotEnoughSeats = errors.New("not enough seats available")

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) BookingRepository {
	if db == nil {
		panic("db is nil")
	}

	return BookingRepository{db: db}
}

func (b BookingRepository) AddBooking(ctx context.Context, booking entities.Booking) error {
	updateFn := func(ctx context.Context, tx *sqlx.Tx) error {
		availableSeats, err := getAvailableSeats(ctx, tx, booking.ShowID)
		if err != nil {
			return err
		}

		alreadyBookedSeats, err := getAlreadyBookedSeats(ctx, tx, booking.ShowID)
		if err != nil {
			return err
		}

		if availableSeats-alreadyBookedSeats < booking.NumberOfTickets {
			return ErrNotEnoughSeats
		}

		if err := insertBooking(ctx, tx, booking); err != nil {
			return err
		}

		if err := publishBookingMadeEvent(ctx, tx, booking); err != nil {
			return err
		}

		return nil
	}

	return updateInTx(ctx, b.db, sql.LevelSerializable, updateFn)
}

func getAvailableSeats(ctx context.Context, tx *sqlx.Tx, showID uuid.UUID) (int, error) {
	var availableSeats int
	err := tx.GetContext(ctx, &availableSeats, `
		SELECT
			number_of_tickets AS available_seats
		FROM
			shows
		WHERE
			show_id = $1
	`, showID)
	if err != nil {
		return 0, fmt.Errorf("could not get available seats: %w", err)
	}
	return availableSeats, nil
}

func getAlreadyBookedSeats(ctx context.Context, tx *sqlx.Tx, showID uuid.UUID) (int, error) {
	var alreadyBookedSeats int
	err := tx.GetContext(ctx, &alreadyBookedSeats, `
		SELECT
			COALESCE(SUM(number_of_tickets), 0) AS already_booked_seats
		FROM
			bookings
		WHERE
			show_id = $1
	`, showID)
	if err != nil {
		return 0, fmt.Errorf("could not get already booked seats: %w", err)
	}
	return alreadyBookedSeats, nil
}

func insertBooking(ctx context.Context, tx *sqlx.Tx, booking entities.Booking) error {
	insertSql := `
		INSERT INTO
			bookings (booking_id, show_id, number_of_tickets, customer_email)
		VALUES
			(:booking_id, :show_id, :number_of_tickets, :customer_email)
	`
	_, err := tx.NamedExecContext(ctx, insertSql, booking)
	if err != nil {
		return fmt.Errorf("could not add booking: %w", err)
	}
	return nil
}

func publishBookingMadeEvent(ctx context.Context, tx *sqlx.Tx, booking entities.Booking) error {
	publisher, err := outbox.NewPublisherForDB(ctx, tx)
	if err != nil {
		return fmt.Errorf("could not create event bus: %w", err)
	}

	bus := event.NewEventBus(publisher)

	e := &entities.BookingMade_v1{
		Header:          entities.NewMessageHeader(),
		NumberOfTickets: booking.NumberOfTickets,
		BookingID:       booking.BookingID,
		CustomerEmail:   booking.CustomerEmail,
		ShowID:          booking.ShowID,
	}

	return bus.Publish(ctx, e)
}
