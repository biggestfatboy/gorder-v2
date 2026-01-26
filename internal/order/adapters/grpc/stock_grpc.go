package grpc

import (
	"context"
	"errors"
	"github.com/biggestfatboy/gorder-v2/common/logging"

	"github.com/biggestfatboy/gorder-v2/common/genproto/orderpb"
	"github.com/biggestfatboy/gorder-v2/common/genproto/stockpb"
)

type StockGRPC struct {
	client stockpb.StockServiceClient
}

func NewStockGRPC(client stockpb.StockServiceClient) *StockGRPC {
	return &StockGRPC{client: client}
}

func (s StockGRPC) CheckIfItemsInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (resp *stockpb.CheckIfItemsInStockResponse, err error) {
	_, deferLog := logging.WhenRequest(ctx, "CheckIfItemsInStock", items)
	defer deferLog(resp, &err)
	if items == nil {
		return nil, errors.New("grpc items cannot be empty")
	}
	return s.client.CheckIfItemsInStock(ctx, &stockpb.CheckIfItemsInStockRequest{Items: items})
}

func (s StockGRPC) GetItems(ctx context.Context, itemIDs []string) (items []*orderpb.Item, err error) {
	_, deferLog := logging.WhenRequest(ctx, "GetItems", itemIDs)
	defer deferLog(items, &err)
	resp, err := s.client.GetItems(ctx, &stockpb.GetItemsRequest{ItemIDs: itemIDs})
	if err != nil {
		return nil, err
	}
	return resp.Items, err
}
