package service

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/broker"
	"github.com/biggestfatboy/gorder-v2/common/entity"
	"github.com/biggestfatboy/gorder-v2/order/domain/order"
	"github.com/pkg/errors"
)

type OrderDomainService struct {
	Repo           order.Repository
	EventPublisher order.EventPublisher
}

func NewOrderDomainService(repo order.Repository, eventPublisher order.EventPublisher) *OrderDomainService {
	return &OrderDomainService{Repo: repo, EventPublisher: eventPublisher}
}

func (s *OrderDomainService) CreateOrder(ctx context.Context, od *order.Order) (res *entity.Order, err error) {
	root := order.NewAggregateRoot(order.Identity{
		CustomerID: od.CustomerID,
		OrderID:    od.ID,
	}, od)
	o, err := s.Repo.Create(ctx, root.OrderEntity)
	if err != nil {
		return nil, err
	}
	err = s.EventPublisher.Publish(ctx, order.DomainEvent{
		Dest: broker.EventOrderCreated,
		Data: o,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "publish event error q.Name=%s", broker.EventOrderCreated)
	}
	return &entity.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       o.Items,
	}, nil
}
