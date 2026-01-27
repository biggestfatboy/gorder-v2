package adapters

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/entity"
	"github.com/biggestfatboy/gorder-v2/stock/infrastructure/persistent"
	"github.com/biggestfatboy/gorder-v2/stock/infrastructure/persistent/builder"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MySQLStockRepository struct {
	db *persistent.MySQL
}

func NewMySQLStockRepository(db *persistent.MySQL) *MySQLStockRepository {
	return &MySQLStockRepository{db: db}
}

func (m MySQLStockRepository) GetItems(ctx context.Context, ids []string) ([]*entity.Item, error) {
	//TODO implement me
	panic("implement me")
}

func (m MySQLStockRepository) GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error) {

	query := builder.NewStock().ProductIDs(ids...)
	data, err := m.db.BatchGetStockByID(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "BatchGetStockByID error")
	}
	var result []*entity.ItemWithQuantity
	for _, d := range data {
		result = append(result, &entity.ItemWithQuantity{
			ID:       d.ProductID,
			Quantity: d.Quantity,
		})
	}
	return result, nil
}
func (m MySQLStockRepository) UpdateStock(ctx context.Context,
	data []*entity.ItemWithQuantity,
	updateFn func(
		ctx context.Context,
		existing []*entity.ItemWithQuantity,
		query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error)) error {
	return m.db.StartTransaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				logrus.Warnf("update stock transaction err=%v", err)
			}
		}()
		err = m.updatePessimistic(ctx, tx, data, updateFn)
		return err
	})
}

func (m MySQLStockRepository) unmarshalFromDatabase(dest []persistent.StockModel) []*entity.ItemWithQuantity {
	res := make([]*entity.ItemWithQuantity, len(dest))
	for i, item := range dest {
		res[i] = &entity.ItemWithQuantity{
			ID:       item.ProductID,
			Quantity: item.Quantity,
		}
	}
	return res
}

func (m MySQLStockRepository) updatePessimistic(ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error)) error {
	var dest []persistent.StockModel

	dest, err := m.db.BatchGetStockByID(ctx, builder.NewStock().ForUpdate().ProductIDs(getIDFromEntities(data)...))
	if err != nil {
		return errors.Wrap(err, "failed to find data")
	}
	existing := m.unmarshalFromDatabase(dest)
	updated, err := updateFn(ctx, existing, data)
	if err != nil {
		return err
	}
	for _, upd := range updated {
		for _, query := range data {
			if upd.ID != query.ID {
				continue
			}
			if err = m.db.Update(ctx,
				tx,
				builder.NewStock().ProductIDs(query.ID).QuantityGT(query.Quantity),
				map[string]any{
					"quantity": gorm.Expr("quantity - ?", query.Quantity),
				}); err != nil {
				return errors.Wrapf(err, "unable to update %s", upd.ID)
			}
		}

	}
	return nil
}

func (m MySQLStockRepository) updateOptimistic(ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error)) error {

	for _, queryData := range data {
		var latestRecord *persistent.StockModel
		latestRecord, err := m.db.GetStockByID(ctx, builder.NewStock().ProductIDs(queryData.ID))
		if err != nil {
			return err
		}
		err = m.db.Update(ctx,
			tx,
			builder.NewStock().ProductIDs(queryData.ID).Versions(latestRecord.Version).QuantityGT(queryData.Quantity),
			map[string]any{
				"quantity": gorm.Expr("quantity - ?", queryData.Quantity),
				"version":  latestRecord.Version + 1,
			})
		if err != nil {
			return err
		}
	}

	return nil
}

func getIDFromEntities(data []*entity.ItemWithQuantity) []string {
	var ids []string
	for _, i := range data {
		ids = append(ids, i.ID)
	}
	return ids
}
