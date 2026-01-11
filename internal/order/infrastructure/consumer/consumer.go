package consumer

import (
	"context"
	"encoding/json"
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
			c.handleMessage(msg, q, ch)
		}
	}()
	<-forever
}

func (c *Consumer) handleMessage(msg amqp.Delivery, q amqp.Queue, ch *amqp.Channel) {
	logrus.Infof("order receive a message from %s, msg=%v", q.Name, string(msg.Body))

	o := &order.Order{}
	if err := json.Unmarshal(msg.Body, o); err != nil {
		logrus.Infof("failed to unmarshall msg.body to domian.order, err=%v", err)
		_ = msg.Nack(false, false)
		return
	}
	if _, err := c.app.Commands.UpdateOrder.Handle(context.Background(), command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *order.Order) (*order.Order, error) {
			if err := o.IsPaid(); err != nil {
				return nil, err
			}
			return order, nil
		}}); err != nil {
		logrus.Infof("failed to updating order, orderID = %v, err=%v", o.ID, err)
		// TODO : retry
		return
	}
	_ = msg.Ack(false)
	logrus.Infof("Order consumer paid event success")
}
