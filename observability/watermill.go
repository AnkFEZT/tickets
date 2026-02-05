package observability

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TracingPublisherDecorator struct {
	message.Publisher
}

func (t TracingPublisherDecorator) Publish(topic string, messages ...*message.Message) error {
	for i := range messages {
		otel.GetTextMapPropagator().Inject(messages[i].Context(), propagation.MapCarrier(messages[i].Metadata))
	}
	return t.Publisher.Publish(topic, messages...)
}
