package processor

import (
	"context"

	"github.com/biggestfatboy/gorder-v2/common/genproto/orderpb"
)

// stub

type InmemProcessor struct {
}

func NewInmemProcessor() *InmemProcessor {
	return &InmemProcessor{}
}

func (i InmemProcessor) CreatePaymentLink(_ context.Context, _ *orderpb.Order) (string, error) {
	return "inmem-payment-link", nil
}
