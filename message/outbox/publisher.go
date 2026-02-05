package outbox

import (
	"context"
	"fmt"

	"tickets/observability"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
)

func NewPublisherForDB(ctx context.Context, tx *sqlx.Tx) (message.Publisher, error) {
	var publisher message.Publisher

	logger := watermill.NewSlogLogger(log.FromContext(ctx))

	config := sql.PublisherConfig{
		SchemaAdapter: sql.DefaultPostgreSQLSchema{},
	}
	publisher, err := sql.NewPublisher(tx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create outbox publisher: %w", err)
	}

	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}
	publisher = observability.TracingPublisherDecorator{Publisher: publisher}

	publisher = forwarder.NewPublisher(publisher, forwarder.PublisherConfig{
		ForwarderTopic: outboxTopic,
	})
	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}
	publisher = observability.TracingPublisherDecorator{Publisher: publisher}

	return publisher, nil
}
