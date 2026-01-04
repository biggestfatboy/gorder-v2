package app

import "github.com/biggestfatboy/gorder-v2/order/app/query"

type Application struct {
	Commands
	Queries
}

type Commands struct {
}

type Queries struct {
	GetCustomerOrder query.GetCustomerOrderHandler
}
