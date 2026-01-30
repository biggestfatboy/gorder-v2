package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/broker"
	"github.com/biggestfatboy/gorder-v2/common/consts"
	"github.com/biggestfatboy/gorder-v2/common/convertor"
	"github.com/biggestfatboy/gorder-v2/common/entity"
	"github.com/biggestfatboy/gorder-v2/common/genproto/orderpb"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"time"
)

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
	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	tr := otel.Tracer("rabbitMQ")
	ctx, span := tr.Start(ctx, fmt.Sprintf("rabbitMQ.%s.consume", q.Name))
	defer span.End()
	defer func() {
		if err != nil {
			logging.Warnf(ctx, nil, "consume failed || from=%s || msg=%+v || err=%v", q.Name, msg, err)
			_ = msg.Nack(false, false)
		} else {
			logging.Infof(ctx, nil, "%s", "consume success")
			_ = msg.Ack(false)
		}
	}()
	o := &entity.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "fail to unmarshall msg to order")
		return
	}
	if o.Status != "paid" {
		err = errors.New("order not paid, cannot cook")
		return
	}
	cook(ctx, o)
	span.AddEvent(fmt.Sprintf("order_cook: %v", o))
	if err = c.orderGRPC.UpdateOrder(ctx, &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      consts.OrderStatusReady,
		PaymentLink: o.PaymentLink,
		Items:       convertor.NewItemConvertor().EntitiesToProtos(o.Items),
	}); err != nil {
		logging.Errof(ctx, nil, "failed to updating order, orderID = %v, err=%v", o.ID, err)
		if err = broker.HandlerRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error, error handling retry, messageID=%s || err=%v", msg.MessageId, err)
		}
		return
	}
	span.AddEvent("kitchen.order.finished.updated")
}

func cook(ctx context.Context, o *entity.Order) {
	logrus.WithContext(ctx).Printf("cooking order: %s\n", o.ID)
	time.Sleep(5 * time.Second)
	logrus.WithContext(ctx).Printf("order %s done!\n", o.ID)
}
