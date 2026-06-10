package model

import "time"

type PayPalTransaction struct {
	Date          time.Time
	TransactionID string
	ItemTitle     string
	Gross         Money
	Fee           Money
	Net           Money
	Status        string
	Currency      string
	Quantity      int
}

func (t PayPalTransaction) IsRefund() bool {
	return t.Gross < 0
}
