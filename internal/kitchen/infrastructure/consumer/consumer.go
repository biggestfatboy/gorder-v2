package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/broker"
	"github.com/biggestfatboy/gorder-v2/common/genproto/orderpb"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"time"
)

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*orderpb.Item
}

type OrderService interface {
	UpdateOrder(ctx context.Context, order *orderpb.Order) error
}

type Consumer struct {
	orderGRPC OrderService
}

func NewConsumer(orderGRPC OrderService) *Consumer {
	return &Consumer{orderGRPC: orderGRPC}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	if err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil); err != nil {
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
	var err error
	logrus.Infof("Kitchen receive a message from %s, msg=%v", q.Name, string(msg.Body))
	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	tr := otel.Tracer("rabbitMQ")
	mqctx, span := tr.Start(ctx, fmt.Sprintf("rabbitMQ.%s.consume", q.Name))
	defer func() {
		span.End()
		if err != nil {
			_ = msg.Nack(false, false)
		} else {
			_ = msg.Ack(false)
		}
	}()
	o := &Order{}
	if err := json.Unmarshal(msg.Body, o); err != nil {
		logrus.Infof("failed to unmarshall msg to order, err=%v", err)
		return
	}
	if o.Status != "paid" {
		err = errors.New("order not paid, cannot cook")
		return
	}
	cook(o)
	span.AddEvent(fmt.Sprintf("order_cook: %v", o))
	if err := c.orderGRPC.UpdateOrder(mqctx, &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      "ready",
		PaymentLink: o.PaymentLink,
		Items:       o.Items,
	}); err != nil {
		if err = broker.HandlerRetry(mqctx, ch, &msg); err != nil {
			logrus.Warnf("kitchen: error handling retry: err=%v", err)
		}
		return
	}
	span.AddEvent("kitchen.order.finished.updated")
	logrus.Infof("Consumer success")
}

func cook(o *Order) {
	logrus.Printf("cooking order: %s\n", o.ID)
	time.Sleep(5 * time.Second)
	logrus.Printf("order %s done!\n", o.ID)
}
