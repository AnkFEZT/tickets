package db

import (
	"fmt"
	"tickets/message/outbox"

	"github.com/jmoiron/sqlx"
)

func InitializeDatabaseSchema(db *sqlx.DB) error {
	initScript := `
		CREATE TABLE IF NOT EXISTS tickets (
			ticket_id UUID PRIMARY KEY,
			price_amount NUMERIC(10, 2) NOT NULL,
			price_currency CHAR(3) NOT NULL,
			customer_email VARCHAR(255) NOT NULL,
			deleted_at TIMESTAMP NULL
		);

		CREATE TABLE IF NOT EXISTS read_model_ops_bookings (
			booking_id UUID PRIMARY KEY,
			payload JSONB NOT NULL
		);

		CREATE TABLE IF NOT EXISTS shows (
			show_id UUID PRIMARY KEY,
			dead_nation_id UUID NOT NULL,
			number_of_tickets INT NOT NULL,
			start_time TIMESTAMP NOT NULL,
			title VARCHAR(255) NOT NULL,
			venue VARCHAR(255) NOT NULL,

			UNIQUE (dead_nation_id)
		);

		CREATE TABLE IF NOT EXISTS bookings (
			booking_id UUID PRIMARY KEY,
			show_id UUID NOT NULL,
			number_of_tickets INT NOT NULL,
			customer_email VARCHAR(255) NOT NULL,
			FOREIGN KEY (show_id) REFERENCES shows(show_id)
		);

		CREATE TABLE IF NOT EXISTS events (
			event_id UUID PRIMARY KEY,
			published_at TIMESTAMP NOT NULL,
			event_name VARCHAR(255) NOT NULL,
			event_payload JSONB NOT NULL
		);
	`

	if _, err := db.Exec(initScript); err != nil {
		return fmt.Errorf("could not initialize database schema: %w", err)
	}

	if err := outbox.InitializeSchema(db); err != nil {
		return fmt.Errorf("could not initialize outbox schema: %w", err)
	}

	return nil
}
