package command

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/genproto/orderpb"
)

type OrderService interface {
	UpdateOrder(ctx context.Context, order *orderpb.Order) error
}
