package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/biggestfatboy/gorder-v2/common/broker"
	"github.com/biggestfatboy/gorder-v2/order/app"
	"github.com/biggestfatboy/gorder-v2/order/app/command"
	"github.com/biggestfatboy/gorder-v2/order/domain/order"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	app app.Application
}

func NewConsumer(app app.Application) *Consumer {
	return &Consumer{app: app}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(broker.EventOrderPaid, true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Warnf("fail to consume: queue=%s,err=%v", q.Name, err)
	}

	var forever chan struct{}
	go func() {
		for msg := range msgs {
			c.handleMessage(ch, msg, q)
		}
	}()
	<-forever
}

func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("rabbitMQ")
	_, span := t.Start(ctx, fmt.Sprintf("rabbitMQ.%s.consumer", q.Name))
	defer span.End()
	var err error
	defer func() {
		if err != nil {
			logging.Warnf(ctx, nil, "consume failed || from=%s || msg=%+v || err=%v", q.Name, msg, err)
			_ = msg.Nack(false, false)
		} else {
			logging.Infof(ctx, nil, "%s", "consume success")
			_ = msg.Ack(false)
		}
	}()
	logrus.Infof("order receive a message from %s, msg=%v", q.Name, string(msg.Body))

	o := &order.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "failed to unmarshall msg.body to domian.order")
		return
	}
	if _, err = c.app.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *order.Order) (*order.Order, error) {
			if err = o.IsPaid(); err != nil {
				return nil, err
			}
			return order, nil
		}}); err != nil {
		logging.Errof(ctx, nil, "failed to updating order, orderID = %v, err=%v", o.ID, err)

		if err = broker.HandlerRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error, error handling retry, messageID=%s || err=%v", msg.MessageId, err)
		}
		return
	}

	span.AddEvent("order.updated")
	logrus.Infof("Order consumer paid event success")
}
