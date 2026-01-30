package command

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/consts"
	"github.com/biggestfatboy/gorder-v2/common/convertor"
	"github.com/biggestfatboy/gorder-v2/common/decorator"
	"github.com/biggestfatboy/gorder-v2/common/entity"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/biggestfatboy/gorder-v2/payment/domain"
	"github.com/sirupsen/logrus"
)

//TODO: ACL清理

type CreatePayment struct {
	Order *entity.Order
}

type CreatePaymentHandler decorator.CommandHandler[CreatePayment, string]

type createPaymentHandler struct {
	processor domain.Processor
	orderGRPC OrderService
}

func NewCreatePaymentHandler(
	processor domain.Processor,
	orderGRPC OrderService,
	logger *logrus.Logger,
	metricClient decorator.MetricsClient) CreatePaymentHandler {
	if processor == nil {
		panic("nil orderRepo")
	}
	if orderGRPC == nil {
		panic("nil stockGRPC")
	}

	return decorator.ApplyCommandDecorators[CreatePayment, string](
		createPaymentHandler{
			processor: processor,
			orderGRPC: orderGRPC,
		},
		logger,
		metricClient)

}

func (c createPaymentHandler) Handle(ctx context.Context, cmd CreatePayment) (string, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "CreatePaymentHandler", cmd, err)
	link, err := c.processor.CreatePaymentLink(ctx, cmd.Order)
	if err != nil {
		return "", err
	}
	newOrder, err := entity.NewValidOrder(
		cmd.Order.ID,
		cmd.Order.CustomerID,
		consts.OrderStatusWaitingForPayment,
		link,
		cmd.Order.Items,
	)
	err = c.orderGRPC.UpdateOrder(ctx, convertor.NewOrderConvertor().EntityToProto(newOrder))
	return link, err
}
