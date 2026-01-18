package query

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/stock/entity"
	"github.com/biggestfatboy/gorder-v2/stock/infrastructure/integration"

	"github.com/biggestfatboy/gorder-v2/common/decorator"
	domain "github.com/biggestfatboy/gorder-v2/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

type CheckIfItemsInStock struct {
	Items []*entity.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*entity.Item]

type checkIfItemsInStockHandler struct {
	stockRepo domain.Repository
	stripeAPI *integration.StripeAPI
}

func NewCheckIfItemsInStockHandler(
	stockRepo domain.Repository,
	stripeAPI *integration.StripeAPI,
	logger *logrus.Entry,
	metricClient decorator.MetricsClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	if stripeAPI == nil {
		panic("nil stripeAPI")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*entity.Item](
		checkIfItemsInStockHandler{stockRepo: stockRepo, stripeAPI: stripeAPI},
		logger,
		metricClient,
	)
}

// TODO:删掉

var stub = map[string]string{
	"1": "price_1SmqvRDQBps38awRPFPngavU",
	"2": "price_1SnYWUDQBps38awR2EKwnRV2",
}

func (g checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*entity.Item, error) {
	if err := g.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}
	var res []*entity.Item
	for _, i := range query.Items {
		priceID, err := g.stripeAPI.GetPriceByProductID(ctx, i.ID)
		if err != nil || priceID == "" {
			return nil, err
		}
		res = append(res, &entity.Item{
			ID:       i.ID,
			Quantity: i.Quantity,
			PriceID:  priceID,
		})
	}
	//TODO 扣库存
	return res, nil
}

func (h checkIfItemsInStockHandler) checkStock(ctx context.Context, queryItems []*entity.ItemWithQuantity) error {
	var ids []string
	for _, item := range queryItems {
		ids = append(ids, item.ID)
	}
	records, err := h.stockRepo.GetStock(ctx, ids)
	if err != nil {
		return err
	}
	idQuantity := make(map[string]int32)
	for _, r := range records {
		idQuantity[r.ID] += r.Quantity
	}
	var (
		ok       = true
		failedOn []entity.FailedItem
	)

	for _, item := range queryItems {
		if item.Quantity > idQuantity[item.ID] {
			ok = false
			failedOn = append(failedOn, entity.FailedItem{
				ID:   item.ID,
				Want: item.Quantity,
				Have: idQuantity[item.ID],
			})
		}
	}
	if ok {
		return nil
	}
	return domain.ExceedStockError{FailedOnItems: failedOn}
}

func getStubPriceID(id string) string {
	priceID, ok := stub[id]
	if !ok {
		priceID = stub["1"]
	}
	return priceID
}
