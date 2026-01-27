package order

import (
	"errors"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/entity"

	"github.com/stripe/stripe-go/v84"
)

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*entity.Item
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

func (o *Order) IsPaid() error {
	if o.Status == string(stripe.CheckoutSessionPaymentStatusPaid) {
		return nil
	}
	return fmt.Errorf("order statuts not paid, order id = %s, status=%x", o.ID, o.Status)
}
func NewPendingOrder(customerID string, items []*entity.Item) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}

	if items == nil {
		return nil, errors.New("empty items")
	}
	return &Order{CustomerID: customerID, Status: "pending", Items: items}, nil
}
