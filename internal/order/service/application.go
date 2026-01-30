package service

import (
	"context"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/broker"
	grpcClient "github.com/biggestfatboy/gorder-v2/common/client"
	_ "github.com/biggestfatboy/gorder-v2/common/config"
	"github.com/biggestfatboy/gorder-v2/common/metrics"
	"github.com/biggestfatboy/gorder-v2/order/adapters"
	"github.com/biggestfatboy/gorder-v2/order/adapters/grpc"
	"github.com/biggestfatboy/gorder-v2/order/app"
	"github.com/biggestfatboy/gorder-v2/order/app/command"
	"github.com/biggestfatboy/gorder-v2/order/app/query"
	"github.com/biggestfatboy/gorder-v2/order/infrastructure/mq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

func NewApplication(ctx context.Context) (app.Application, func()) {

	stockClient, closeStockClient, err := grpcClient.NewStockGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	ch, closeCh := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"))

	stockGRPC := grpc.NewStockGRPC(stockClient)
	return newApplication(ctx, stockGRPC, ch), func() {
		_ = closeStockClient()
		_ = ch.Close()
		_ = closeCh()
	}

}

func newApplication(_ context.Context, stockGRPC query.StockService, ch *amqp.Channel) app.Application {
	mongoClient := newMongoClient()
	orderRepo := adapters.NewOrderRepositoryMongo(mongoClient)
	metricsClient := metrics.NewPrometheusMetricsClient(&metrics.PrometheusMetricsClientConfig{
		Addr:        viper.GetString("order.metrics_export_addr"),
		ServiceName: viper.GetString("order.service_name"),
	})
	eventPublisher := mq.NewRabbitMQEventPublisher(ch)
	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(orderRepo, stockGRPC, eventPublisher, logrus.StandardLogger(), metricsClient),
			UpdateOrder: command.NewUpdateOrderHandler(orderRepo, logrus.StandardLogger(), metricsClient),
		},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(orderRepo, logrus.StandardLogger(), metricsClient),
		},
	}
}

func newMongoClient() *mongo.Client {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s",
		viper.GetString("mongo.user"),
		viper.GetString("mongo.password"),
		viper.GetString("mongo.host"),
		viper.GetString("mongo.port"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	if err = c.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	return c
}
