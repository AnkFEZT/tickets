package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/jmoiron/sqlx"
)

func NewPostgresSubscriber(db *sqlx.DB, logger watermill.LoggerAdapter) *sql.Subscriber {
	config := sql.SubscriberConfig{
		PollInterval:   time.Millisecond * 100,
		SchemaAdapter:  sql.DefaultPostgreSQLSchema{},
		OffsetsAdapter: sql.DefaultPostgreSQLOffsetsAdapter{},
	}

	subscriber, err := sql.NewSubscriber(db, config, logger)
	if err != nil {
		panic(fmt.Errorf("failed to create new watermill sql subscriber: %w", err))
	}

	return subscriber
}

func InitializeSchema(db *sqlx.DB) error {
	logger := watermill.NewSlogLogger(log.FromContext(context.Background()))
	sqlSub := NewPostgresSubscriber(db, logger)
	return sqlSub.SubscribeInitialize(outboxTopic)
}
