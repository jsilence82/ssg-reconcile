package model

import "time"

type TicketCategory int

const (
	CategoryGeneral TicketCategory = iota
	CategoryStudent
	CategoryComp
	CategoryUnknown
)

func (c TicketCategory) String() string {
	switch c {
	case CategoryGeneral:
		return "General"
	case CategoryStudent:
		return "Student"
	case CategoryComp:
		return "Comp"
	default:
		return "Unknown"
	}
}

var AllCategories = []TicketCategory{CategoryGeneral, CategoryStudent, CategoryComp}

type TTTicket struct {
	OrderID       string
	TransactionID string
	EventID       string
	EventName     string
	EventStart    time.Time
	TicketCode    string
	Category      TicketCategory
	Value         Money
	PaymentMethod string
	CheckedIn     bool
	Attended      bool
}

func (t TTTicket) IsComp() bool {
	return t.PaymentMethod == "NO_COST"
}
