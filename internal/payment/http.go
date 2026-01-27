package main

import (
	"encoding/json"
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common/entity"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"io/ioutil"
	"net/http"

	"github.com/biggestfatboy/gorder-v2/common/broker"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

type PaymentHandler struct {
	channel *amqp.Channel
}

func NewPaymentHandler(ch *amqp.Channel) *PaymentHandler {
	return &PaymentHandler{channel: ch}
}

func (h *PaymentHandler) RegisterRoutes(c *gin.Engine) {
	c.POST("/api/webhook", h.handleWebhook)
}

func (h *PaymentHandler) handleWebhook(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			logging.Warnf(c.Request.Context(), nil, "handleWebhook err: %v", err)
		} else {
			logging.Infof(c.Request.Context(), nil, "%s", "handleWebhook success")
		}
	}()
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		err = errors.Wrap(err, "Error reading request body")
		c.JSON(http.StatusServiceUnavailable, err)
		return
	}

	// Replace this endpoint secret with your endpoint's unique secret
	// If you are testing with the CLI, find the secret by running 'stripe listen'
	// If you are using an endpoint defined with the API or dashboard, look in your webhook settings
	// at https://dashboard.stripe.com/webhooks
	endpointSecret := viper.GetString("ENDPOINT_STRIPE_SECRET")
	signatureHeader := c.Request.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		err = errors.Wrap(err, "Webhook verifying webhook signature failed")
		c.JSON(http.StatusBadRequest, err.Error())
		// Return a 400 error on a bad signature
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err = json.Unmarshal(event.Data.Raw, &session); err != nil {
			err = errors.Wrap(err, "error unmarshall event.data.raw into session")
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			var items []*entity.Item
			_ = json.Unmarshal([]byte(session.Metadata["items"]), &items)
			body := &entity.Order{
				ID:          session.Metadata["orderID"],
				CustomerID:  session.Metadata["customerID"],
				Status:      string(stripe.CheckoutSessionPaymentStatusPaid),
				PaymentLink: session.Metadata["paymentLink"],
				Items:       items,
			}

			tr := otel.Tracer("rabbitMQ")
			ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("rabbitMQ.%s.publish", broker.EventOrderPaid))
			defer span.End()
			err = broker.PublishEvent(ctx, broker.PublishEventRequest{
				Channel:  h.channel,
				Routing:  broker.FanOut,
				Queue:    "",
				Exchange: broker.EventOrderPaid,
				Body:     body,
			})
		}
	}

	c.JSON(http.StatusOK, nil)
}
