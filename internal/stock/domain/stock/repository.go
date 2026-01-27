package stock

import (
	"context"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/entity"
	"strings"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*entity.Item, error)
	GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error)
	UpdateStock(ctx context.Context, data []*entity.ItemWithQuantity, updateFn func(ctx context.Context, existing []*entity.ItemWithQuantity, query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error)) error
}

type NotFoundError struct {
	Missing []string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("these items not found in stock: %s", strings.Join(e.Missing, ","))
}

type ExceedStockError struct {
	FailedOnItems []entity.FailedItem
}

func (e ExceedStockError) Error() string {
	var info []string
	for _, item := range e.FailedOnItems {
		info = append(info, fmt.Sprintf("product_id=%s, want %d, have %d\n", item.ID, item.Want, item.Have))
	}
	return fmt.Sprintf("not enough stock for [%s]", strings.Join(info, ","))
}
