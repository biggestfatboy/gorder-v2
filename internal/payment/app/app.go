package app

import "github.com/biggestfatboy/gorder-v2/payment/app/command"

type Application struct {
	Commands
}

type Commands struct {
	CreatePayment command.CreatePaymentHandler
}
