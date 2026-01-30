package order

import (
	"errors"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/consts"
	"github.com/biggestfatboy/gorder-v2/common/entity"
	"slices"
)

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*entity.Item
}

func (o *Order) UpdataStatus(to string) error {
	if !o.isValidStatusTransition(to) {
		return fmt.Errorf("Cannot transition from %s to %s", o.Status, to)
	}
	o.Status = to
	return nil
}

func (o *Order) UpdataPaymentLink(paymentLink string) error {
	o.PaymentLink = paymentLink
	return nil
}

func (o *Order) UpdataItems(items []*entity.Item) error {
	o.Items = items
	return nil
}

func (o *Order) isValidStatusTransition(to string) bool {
	switch o.Status {
	case consts.OrderStatusPending:
		return slices.Contains([]string{consts.OrderStatusWaitingForPayment}, to)
	case consts.OrderStatusWaitingForPayment:
		return slices.Contains([]string{consts.OrderStatusPaid}, to)
	case consts.OrderStatusPaid:
		return slices.Contains([]string{consts.OrderStatusReady}, to)
	default:
		return false
	}
}

func NewOrder(ID string, customerID string, status string, paymentLink string, items []*entity.Item) (*Order, error) {
	if ID == "" {
		return nil, errors.New("empty id")
	}
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}
	if status == "" {
		return nil, errors.New("empty status")
	}
	if items == nil {
		return nil, errors.New("empty items")
	}
	return &Order{ID: ID, CustomerID: customerID, Status: status, PaymentLink: paymentLink, Items: items}, nil
}

func NewPendingOrder(customerID string, items []*entity.Item) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}

	if items == nil {
		return nil, errors.New("empty items")
	}
	return &Order{CustomerID: customerID, Status: consts.OrderStatusPending, Items: items}, nil
}
