package outbox

import (
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
)

func AddForwarderHandler(
	subscriber message.Subscriber,
	publisher message.Publisher,
	router *message.Router,
	logger watermill.LoggerAdapter,
) {

	config := forwarder.Config{
		ForwarderTopic: outboxTopic,
		Router:         router,
		Middlewares:    []message.HandlerMiddleware{loggerMW},
	}

	_, err := forwarder.NewForwarder(subscriber, publisher, logger, config)
	if err != nil {
		panic(err)
	}
}

func loggerMW(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		log.FromContext(msg.Context()).With(
			"message_id", msg.UUID,
			"payload", string(msg.Payload),
			"metadata", msg.Metadata,
		).Info("Forwarding message")

		return h(msg)
	}
}
