package reconcile

import (
	"time"

	"github.com/jsilence82/ssg-reconcile/internal/config"
	"github.com/jsilence82/ssg-reconcile/internal/model"
)

// Aggregate builds a PerformanceSummary for a given event ID from its orders
// and comp tickets. For cross-performance orders, PayPal gross/fee are prorated
// by the fraction of ticket value that belongs to this event.
func Aggregate(
	cfg *config.Config,
	eventID string,
	orders []JoinedOrder,
	compTickets []model.TTTicket,
) model.PerformanceSummary {
	perf := cfg.EventIDToPerformance()

	summary := model.NewPerformanceSummary()
	if p, ok := perf[eventID]; ok {
		summary.PerformanceNumber = p.Number
		summary.EventID = eventID
		if d, err := p.ParsedDate(); err == nil {
			summary.Date = d
		}
	}

	seenTxnIDs := make(map[string]bool)

	for _, order := range orders {
		// Sum ticket values for this event only
		var eventValue model.Money
		var orderTotal model.Money
		for _, t := range order.Tickets {
			orderTotal += t.Value
			if t.EventID == eventID {
				eventValue += t.Value
				summary.TicketCounts[t.Category]++
			}
		}

		if orderTotal == 0 {
			continue
		}

		// Prorate gross and fee by fraction of value belonging to this event
		fraction := float64(eventValue) / float64(orderTotal)
		gross := model.Money(float64(order.Transaction.Gross) * fraction)
		fee := model.Money(float64(order.Transaction.Fee) * fraction)

		summary.Gross += gross
		summary.Fees += fee
		summary.Net += gross + fee

		if !seenTxnIDs[order.Transaction.TransactionID] {
			seenTxnIDs[order.Transaction.TransactionID] = true
			summary.TransactionCount++
		}
	}

	// Count comp tickets
	for _, t := range compTickets {
		summary.TicketCounts[t.Category]++
	}

	return summary
}

// SumPerformances produces a grand-total PerformanceSummary across all performances.
func SumPerformances(performances []model.PerformanceSummary) model.PerformanceSummary {
	total := model.NewPerformanceSummary()
	total.Date = time.Time{} // no single date for totals row

	for _, p := range performances {
		total.TransactionCount += p.TransactionCount
		total.Gross += p.Gross
		total.Fees += p.Fees
		total.Net += p.Net
		for cat, n := range p.TicketCounts {
			total.TicketCounts[cat] += n
		}
	}

	return total
}

// ExtractRefunds separates refund transactions from the main transaction list.
func ExtractRefunds(txns []model.PayPalTransaction) (normal []model.PayPalTransaction, refunds []model.RefundRecord) {
	for _, t := range txns {
		if t.IsRefund() {
			refunds = append(refunds, model.RefundRecord{
				TransactionID: t.TransactionID,
				Date:          t.Date,
				Amount:        t.Gross,
			})
		} else {
			normal = append(normal, t)
		}
	}
	return
}
