package main

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/broker"
	grpcClient "github.com/biggestfatboy/gorder-v2/common/client"
	"github.com/biggestfatboy/gorder-v2/common/tracing"
	"github.com/biggestfatboy/gorder-v2/kitchen/adapters"
	"github.com/biggestfatboy/gorder-v2/kitchen/infrastructure/consumer"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/biggestfatboy/gorder-v2/common/config"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
}

func main() {
	serviceName := viper.GetString("kitchen.service-name")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, closeCh := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"))

	defer func() {
		_ = ch.Close()
		_ = closeCh()
	}()
	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		_ = shutdown(ctx)
	}()
	client, closeFunc, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	defer closeFunc()
	orderGRPC := adapters.NewOrderGRPC(client)
	go consumer.NewConsumer(orderGRPC).Listen(ch)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logrus.Infof("receive signal, exiting...")
		os.Exit(0)
	}()

	logrus.Println("to exit, press ctrl+c")
	select {}
}
