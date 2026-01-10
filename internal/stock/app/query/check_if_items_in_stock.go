package query

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/decorator"
	"github.com/biggestfatboy/gorder-v2/common/genproto/orderpb"
	domain "github.com/biggestfatboy/gorder-v2/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

type CheckIfItemsInStock struct {
	Items []*orderpb.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*orderpb.Item]

type checkIfItemsInStockHandler struct {
	stockRepo domain.Repository
}

func NewCheckIfItemsInStockHandler(
	stockRepo domain.Repository,
	logger *logrus.Entry,
	metricClient decorator.MetricsClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*orderpb.Item](
		checkIfItemsInStockHandler{stockRepo: stockRepo},
		logger,
		metricClient,
	)
}

// TODO:删掉

var stub = map[string]string{
	"1": "price_1SmqvRDQBps38awRPFPngavU",
	"2": "price_1SnYWUDQBps38awR2EKwnRV2",
}

func (g checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*orderpb.Item, error) {

	var res []*orderpb.Item
	for _, i := range query.Items {
		// TODO: 改成从数据库 or stripe 获取
		priceID, ok := stub[i.ID]
		if !ok {
			priceID = stub["1"]
		}
		res = append(res, &orderpb.Item{
			ID:       i.ID,
			Quantity: i.Quantity,
			PriceID:  priceID,
		})
	}

	return res, nil
}
