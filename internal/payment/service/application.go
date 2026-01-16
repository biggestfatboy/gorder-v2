package service

import (
	"context"

	grpcClient "github.com/biggestfatboy/gorder-v2/common/client"
	"github.com/biggestfatboy/gorder-v2/common/metrics"
	adapterGRPC "github.com/biggestfatboy/gorder-v2/payment/adapters/grpc"
	"github.com/biggestfatboy/gorder-v2/payment/app"
	"github.com/biggestfatboy/gorder-v2/payment/app/command"
	"github.com/biggestfatboy/gorder-v2/payment/domain"
	"github.com/biggestfatboy/gorder-v2/payment/infrastructure/processor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	orderClient, closeOrderClient, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	orderGRPC := adapterGRPC.NewOrderGRPC(orderClient)
	stripeProcessor := processor.NewStripeProcessor(viper.GetString("stripe-key"))
	return newApplication(ctx, orderGRPC, stripeProcessor), func() {
		_ = closeOrderClient()
	}
}

func newApplication(_ context.Context, orderGRPC command.OrderService, processor domain.Processor) app.Application {
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreatePayment: command.NewCreatePaymentHandler(
				processor,
				orderGRPC,
				logger, metricsClient,
			),
		},
	}
}
