package main

import (
	"fmt"
	"github.com/biggestfatboy/gorder-v2/common"
	client "github.com/biggestfatboy/gorder-v2/common/client/order"
	"github.com/biggestfatboy/gorder-v2/common/consts"
	"github.com/biggestfatboy/gorder-v2/common/handler/errors"
	"github.com/biggestfatboy/gorder-v2/common/tracing"
	"github.com/biggestfatboy/gorder-v2/order/app"
	"github.com/biggestfatboy/gorder-v2/order/app/command"
	"github.com/biggestfatboy/gorder-v2/order/app/dto"
	"github.com/biggestfatboy/gorder-v2/order/app/query"
	"github.com/biggestfatboy/gorder-v2/order/convertor"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HTTPServer struct {
	app app.Application
	common.BaseResponse
}

func (H HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, _ string) {
	//ctx, span := tracing.Start(c, "PostCustomerCustomerIDOrders")
	//defer span.End()

	var (
		req  client.CreateOrderRequest
		err  error
		resp dto.CreateOrderResponse
	)
	defer func() {
		H.Response(c, err, &resp)
	}()
	if err = c.ShouldBindJSON(&req); err != nil {
		err = errors.NewWithError(consts.ErrnoBindRequestError, err)
		return
	}
	if err = H.validate(req); err != nil {
		err = errors.NewWithError(consts.ErrnoRequestValidateError, err)
		return
	}
	logrus.Infof("start to create customer customer order with customer id %s", req.CustomerId)
	r, err := H.app.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: req.CustomerId,
		Items:      convertor.NewItemWithQuantityConvertor().ClientsToEntities(req.Items),
	})
	if err != nil {
		return
	}
	resp = dto.CreateOrderResponse{
		CustomerID:  req.CustomerId,
		OrderID:     r.OrderID,
		RedirectURL: fmt.Sprintf("http://192.168.77.38:8282/success?customerID=%s&orderID=%s", req.CustomerId, r.OrderID),
	}
}

func (H HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerID string, orderID string) {
	var (
		err  error
		resp struct {
			Order *client.Order
		}
	)
	defer func() {
		H.Response(c, err, resp)
	}()
	ctx, span := tracing.Start(c, "GetCustomerCustomerIDOrdersOrderID")
	defer span.End()
	o, err := H.app.GetCustomerOrder.Handle(ctx, query.GetCustomerOrder{
		OrderID:    orderID,
		CustomerID: customerID,
	})
	if err != nil {
		return
	}
	resp.Order = convertor.NewOrderConvertor().EntityToClient(o)
}

func (H HTTPServer) validate(req client.CreateOrderRequest) error {
	for _, v := range req.Items {
		if v.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive, got %d from %s", v.Quantity, v.Id)
		}
	}
	return nil
}
