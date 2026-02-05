package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"tickets/entities"
)

type DataLake struct {
	db *sqlx.DB
}

func NewDataLake(db *sqlx.DB) DataLake {
	if db == nil {
		panic("db is nil")
	}

	return DataLake{db: db}
}

func (s DataLake) StoreEvent(
	ctx context.Context,
	eventID string,
	eventHeader entities.MessageHeader,
	eventName string,
	payload []byte,
) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO events (event_id, published_at, event_name, event_payload) VALUES ($1, $2, $3, $4)`,
		eventID,
		eventHeader.PublishedAt,
		eventName,
		payload,
	)

	var postgresError *pq.Error
	if errors.As(err, &postgresError) && postgresError.Code.Name() == "unique_violation" {
		// handling re-delivery
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not store %s event in data lake: %w", eventID, err)
	}

	return nil
}

func (s DataLake) GetEvents(ctx context.Context) ([]entities.DataLakeEvent, error) {
	var events []entities.DataLakeEvent
	err := s.db.SelectContext(ctx, &events, "SELECT * FROM events ORDER BY published_at ASC")
	if err != nil {
		return nil, fmt.Errorf("could not get events from data lake: %w", err)
	}

	return events, nil
}

func (s DataLake) HasEvents(ctx context.Context) (bool, error) {
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM events")
	if err != nil {
		return false, fmt.Errorf("could not check events count: %w", err)
	}

	return count > 0, nil
}
