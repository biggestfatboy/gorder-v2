package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/broker"
	"github.com/biggestfatboy/gorder-v2/common/genproto/orderpb"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/biggestfatboy/gorder-v2/payment/app"
	"github.com/biggestfatboy/gorder-v2/payment/app/command"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type Consumer struct {
	app app.Application
}

func NewConsumer(app app.Application) *Consumer {
	return &Consumer{app: app}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
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
	tr := otel.Tracer("rabbitMQ")
	ctx, span := tr.Start(ctx, fmt.Sprintf("rabbitMQ.%s.consume", q.Name))
	defer span.End()

	logging.Infof(ctx, nil, "Payment receive a message from %s, msg=%v", q.Name, string(msg.Body))
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

	o := &orderpb.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "fail to unmarshal msg body to order")
		return
	}
	if _, err = c.app.CreatePayment.Handle(ctx, command.CreatePayment{Order: o}); err != nil {
		err = errors.Wrap(err, "fail to create payment")
		if err = broker.HandlerRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error, error handling retry, messageID=%s, err=%v", msg.MessageId, err)
		}
		return
	}
	span.AddEvent("payment.create")
}
