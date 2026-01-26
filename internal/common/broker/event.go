package broker

import (
	"context"
	"encoding/json"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const (
	EventOrderCreated = "order.created"
	EventOrderPaid    = "order.paid"
)

type RoutingType string

const (
	FanOut RoutingType = "fan-out"
	Direct RoutingType = "direct"
)

type PublishEventRequest struct {
	Channel  *amqp.Channel
	Routing  RoutingType
	Queue    string
	Exchange string
	Body     any
}

func PublishEvent(ctx context.Context, p PublishEventRequest) (err error) {
	_, dLog := logging.WhenEventPublish(ctx, p)
	defer dLog(nil, &err)

	if err = checkParam(p); err != nil {
		return err
	}
	switch p.Routing {
	case FanOut:
		return fanOut(ctx, p)
	case Direct:
		return directQueue(ctx, p)
	default:
		logrus.WithContext(ctx).Panicf("unsupported routing type: %s", string(p.Routing))
	}
	return nil
}

func checkParam(p PublishEventRequest) error {
	if p.Channel == nil {
		return errors.New("nil channel")
	}
	return nil
}

func directQueue(ctx context.Context, p PublishEventRequest) (err error) {
	_, err = p.Channel.QueueDeclare(p.Queue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	jsonBody, err := json.Marshal(p.Body)
	if err != nil {
		return err
	}
	return doPublish(ctx, p.Channel, p.Exchange, p.Queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         jsonBody,
		DeliveryMode: amqp.Persistent,
		Headers:      InjectRabbitMQHeaders(ctx),
	})
}

func doPublish(ctx context.Context, channel *amqp.Channel, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
	if err := channel.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg); err != nil {
		logging.Warnf(ctx, nil, "_publish_event_failed || exchange=%s || key=%s || msg=%v", exchange, key, msg)
		return errors.Wrap(err, "publish event error")
	}
	return nil
}

func fanOut(ctx context.Context, p PublishEventRequest) error {
	jsonBody, err := json.Marshal(p.Body)
	if err != nil {
		return err
	}
	return doPublish(ctx, p.Channel, p.Exchange, "", false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         jsonBody,
		DeliveryMode: amqp.Persistent,
		Headers:      InjectRabbitMQHeaders(ctx),
	})
}
