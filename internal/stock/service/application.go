package service

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/stock/infrastructure/integration"
	"github.com/biggestfatboy/gorder-v2/stock/infrastructure/persistent"
	"github.com/spf13/viper"

	"github.com/biggestfatboy/gorder-v2/common/metrics"
	"github.com/biggestfatboy/gorder-v2/stock/adapters"
	"github.com/biggestfatboy/gorder-v2/stock/app"
	"github.com/biggestfatboy/gorder-v2/stock/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(_ context.Context) app.Application {
	stockRepo := adapters.NewMySQLStockRepository(persistent.NewMySQL())
	metricsClient := metrics.NewPrometheusMetricsClient(&metrics.PrometheusMetricsClientConfig{
		Addr:        viper.GetString("stock.metrics_export_addr"),
		ServiceName: viper.GetString("stock.service_name"),
	})
	stripeAPI := integration.NewStripeAPI()
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, stripeAPI, logrus.StandardLogger(), metricsClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logrus.StandardLogger(), metricsClient),
		},
	}
}
