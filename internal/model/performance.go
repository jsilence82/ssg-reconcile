package model

import "time"

type PerformanceSummary struct {
	PerformanceNumber int
	EventID           string
	Date              time.Time
	TransactionCount  int
	Gross             Money
	Fees              Money
	Net               Money
	TicketCounts      map[TicketCategory]int
}

func NewPerformanceSummary() PerformanceSummary {
	return PerformanceSummary{
		TicketCounts: map[TicketCategory]int{
			CategoryGeneral: 0,
			CategoryStudent: 0,
			CategoryComp:    0,
		},
	}
}

func (p PerformanceSummary) TotalTickets() int {
	total := 0
	for _, n := range p.TicketCounts {
		total += n
	}
	return total
}
