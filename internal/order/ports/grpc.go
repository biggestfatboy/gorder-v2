package ports

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/order/convertor"

	"github.com/biggestfatboy/gorder-v2/common/genproto/orderpb"
	"github.com/biggestfatboy/gorder-v2/order/app"
	"github.com/biggestfatboy/gorder-v2/order/app/command"
	"github.com/biggestfatboy/gorder-v2/order/app/query"
	domain "github.com/biggestfatboy/gorder-v2/order/domain/order"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) CreateOrder(ctx context.Context, request *orderpb.CreateOrderRequest) (*emptypb.Empty, error) {
	_, err := G.app.CreateOrder.Handle(ctx, command.CreateOrder{
		CustomerID: request.CustomerID,
		Items:      convertor.NewItemWithQuantityConvertor().ProtosToEntities(request.Items),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (G GRPCServer) GetOrder(ctx context.Context, request *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	logrus.Infof("order_grpc || request_in || request=%+v", request)
	o, err := G.app.GetCustomerOrder.Handle(ctx, query.GetCustomerOrder{
		CustomerID: request.CustomerId,
		OrderID:    request.OrderId,
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return convertor.NewOrderConvertor().EntityToProto(o), nil
}

func (G GRPCServer) UpdateOrder(ctx context.Context, request *orderpb.Order) (_ *emptypb.Empty, err error) {
	logrus.WithContext(ctx).Infof("order_grpc || request_in || request=%+v", request)
	order, err := domain.NewOrder(request.ID,
		request.CustomerID,
		request.Status,
		request.PaymentLink,
		convertor.NewItemConvertor().ProtosToEntities(request.Items))
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
		return nil, err
	}
	_, err = G.app.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: order,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			return order, nil
		},
	})
	return nil, err
}
