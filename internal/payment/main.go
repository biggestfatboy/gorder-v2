package main

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/broker"
	"github.com/biggestfatboy/gorder-v2/common/config"
	"github.com/biggestfatboy/gorder-v2/common/discovery"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/biggestfatboy/gorder-v2/common/server"
	"github.com/biggestfatboy/gorder-v2/payment/infrastructure/consumer"
	"github.com/biggestfatboy/gorder-v2/payment/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatal(err)
	}
}

func main() {
	serverType := viper.GetString("payment.server-to-run")
	serviceName := viper.GetString("payment.service-Name")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	application, cleanup := service.NewApplication(ctx)
	defer cleanup()

	ch, closeCh := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"))

	defer func() {
		_ = ch.Close()
		_ = closeCh()
	}()

	go consumer.NewConsumer(application).Listen(ch)

	deregisterFunc, err := discovery.RegisterToConsul(ctx, serviceName)
	defer func() {
		_ = deregisterFunc()
	}()
	if err != nil {
		logrus.Fatal(err)
	}

	paymentHandler := NewPaymentHandler(ch)
	switch serverType {
	case "http":
		server.RunHTTPServer(serviceName, paymentHandler.RegisterRoutes)
	case "grpc":
		logrus.Panic("unsupported server type: grpc")
	default:
		logrus.Panic("unreachable code")
	}
}
