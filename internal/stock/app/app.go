package app

import "github.com/biggestfatboy/gorder-v2/stock/app/query"

type Application struct {
	Commands
	Queries
}

type Commands struct {
}

type Queries struct {
	CheckIfItemsInStock query.CheckIfItemsInStockHandler
	GetItems            query.GetItemsHandler
}
