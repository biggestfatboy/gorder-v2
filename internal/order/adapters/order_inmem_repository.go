package adapters

import (
	"context"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"strconv"
	"sync"
	"time"

	domain "github.com/biggestfatboy/gorder-v2/order/domain/order"
)

type MemoryOrderRepository struct {
	lock  *sync.RWMutex
	store []*domain.Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	s := []*domain.Order{
		{
			ID:          "fake-ID",
			CustomerID:  "fake-customer-id",
			Status:      "fake-status",
			PaymentLink: "fake-payment-link",
			Items:       nil,
		},
	}
	return &MemoryOrderRepository{
		lock:  &sync.RWMutex{},
		store: s,
	}
}

func (m *MemoryOrderRepository) Create(ctx context.Context, order *domain.Order) (created *domain.Order, err error) {
	_, deferLog := logging.WhenRequest(ctx, "MemoryOrderRepository.Create", map[string]any{"order": order})
	defer deferLog(created, &err)
	m.lock.Lock()
	defer m.lock.Unlock()
	newOrder := &domain.Order{
		ID:          strconv.FormatInt(time.Now().Unix(), 10),
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
	m.store = append(m.store, newOrder)
	return newOrder, nil
}

func (m MemoryOrderRepository) Get(ctx context.Context, id, customerID string) (got *domain.Order, err error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, o := range m.store {
		if o.ID == id && o.CustomerID == customerID {
			return o, nil
		}
	}
	return nil, domain.NotFoundError{OrderID: id}
}

func (m *MemoryOrderRepository) Update(ctx context.Context, order *domain.Order, updateFn func(context.Context, *domain.Order) (*domain.Order, error)) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	found := false

	for i, o := range m.store {
		if o.ID == order.ID && o.CustomerID == order.CustomerID {
			found = true
			updatedOrder, err := updateFn(ctx, order)
			if err != nil {
				return err
			}
			m.store[i] = updatedOrder
		}
	}

	if !found {
		return domain.NotFoundError{OrderID: order.ID}
	}

	return nil
}
