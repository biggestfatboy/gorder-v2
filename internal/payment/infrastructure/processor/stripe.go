package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/tracing"

	"github.com/biggestfatboy/gorder-v2/common/genproto/orderpb"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
)

type StripeProcessor struct {
	apiKey string
}

var (
	successURL = "http://192.168.77.38:8282/success"
)

func NewStripeProcessor(apiKey string) *StripeProcessor {
	if apiKey == "" {
		panic("empty api key")
	}
	stripe.Key = apiKey
	return &StripeProcessor{apiKey: apiKey}
}

func (s *StripeProcessor) CreatePaymentLink(ctx context.Context, order *orderpb.Order) (string, error) {
	_, span := tracing.Start(ctx, "stripe_processor.create_payment_link")
	defer span.End()
	var items []*stripe.CheckoutSessionLineItemParams
	for _, item := range order.Items {
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}
	marshalledItems, _ := json.Marshal(order.Items)
	metaData := map[string]string{
		"orderID":     order.ID,
		"customerID":  order.CustomerID,
		"status":      order.Status,
		"items":       string(marshalledItems),
		"paymentLink": order.PaymentLink,
	}

	params := &stripe.CheckoutSessionParams{
		Metadata:   metaData,
		LineItems:  items,
		Mode:       stripe.String(stripe.CheckoutSessionModePayment),
		SuccessURL: stripe.String(fmt.Sprintf("%s?customerID=%s&orderID=%s", successURL, order.CustomerID, order.ID)),
	}
	result, err := session.New(params)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}
