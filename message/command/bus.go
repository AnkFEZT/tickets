package command

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewCommandBus(pub message.Publisher) *cqrs.CommandBus {
	bus, err := cqrs.NewCommandBusWithConfig(pub, cqrs.CommandBusConfig{
		Marshaler: marshaler,
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return fmt.Sprintf("commands.%s", params.CommandName), nil
		},
	})

	if err != nil {
		panic(err)
	}

	return bus
}
