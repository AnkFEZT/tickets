package message

import (
	"time"
	"tickets/observability"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewRedisPublisher(rdb *redis.Client, watermillLogger watermill.LoggerAdapter) message.Publisher {
	var pub message.Publisher
	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	pub = log.CorrelationPublisherDecorator{Publisher: pub}
	pub = observability.TracingPublisherDecorator{Publisher: pub}

	return pub
}

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     100,
		MinIdleConns: 10,
		PoolTimeout:  4 * time.Second,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}

func NewRedisSubscriber(rdb *redis.Client, consumerGroup string, watermillLogger watermill.LoggerAdapter) message.Subscriber {
	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: consumerGroup,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}
	return sub
}
