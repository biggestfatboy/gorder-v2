package query

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/entity"
	"github.com/biggestfatboy/gorder-v2/common/handler/redis"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/biggestfatboy/gorder-v2/stock/infrastructure/integration"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/biggestfatboy/gorder-v2/common/decorator"
	domain "github.com/biggestfatboy/gorder-v2/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

const (
	redisLockPrefix = "check_stock"
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
	logger *logrus.Logger,
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
	if err := lock(ctx, getLockKey(query)); err != nil {
		return nil, errors.Wrapf(err, "redis lock error: key=%s", getLockKey(query))
	}
	defer func() {
		if err := unlock(ctx, getLockKey(query)); err != nil {
			logging.Warnf(ctx, nil, "redis unlock failed, err=%v", err)
		}
	}()
	var err error
	var res []*entity.Item
	defer func() {
		f := logrus.Fields{
			"query": query,
			"res":   res,
		}
		if err != nil {
			logging.Errof(ctx, f, "checkIfItemsInStock err=%v", err)
		} else {
			logging.Infof(ctx, f, "%s", "checkIfItemsInStock success")
		}
	}()

	for _, i := range query.Items {
		productI, err := g.stripeAPI.GetProductByID(ctx, i.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, entity.NewItem(i.ID, productI.Name, i.Quantity, productI.DefaultPrice.ID))
	}

	if err = g.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}
	return res, nil
}

func getLockKey(query CheckIfItemsInStock) string {
	var ids []string
	for _, i := range query.Items {
		ids = append(ids, i.ID)
	}
	return redisLockPrefix + strings.Join(ids, "_")
}

func lock(ctx context.Context, key string) error {
	return redis.SetNX(ctx, redis.LocalClient(), key, "1", 5*time.Minute)
}

func unlock(ctx context.Context, key string) error {
	return redis.Del(ctx, redis.LocalClient(), key)
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
		return h.stockRepo.UpdateStock(ctx, queryItems, func(
			ctx context.Context,
			existing []*entity.ItemWithQuantity,
			query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error) {
			var newItems []*entity.ItemWithQuantity
			for _, e := range existing {
				for _, q := range query {
					if e.ID == q.ID {
						iq, err := entity.NewValidItemWithQuantity(e.ID, e.Quantity-q.Quantity)
						if err != nil {
							return nil, err
						}
						newItems = append(newItems, iq)
					}
				}
			}
			return newItems, nil
		})
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
