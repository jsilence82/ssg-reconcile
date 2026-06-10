package reconcile

import (
	"github.com/jsilence82/ssg-reconcile/internal/model"
)

// Validate checks each joined order for gross mismatches between PayPal and
// the sum of TT ticket values. feeTolerance is in euro-cents.
func Validate(orders []JoinedOrder, feeTolerance model.Money) []model.MismatchRecord {
	var mismatches []model.MismatchRecord

	for _, order := range orders {
		ttTotal := model.Money(0)
		for _, t := range order.Tickets {
			ttTotal += t.Value
		}

		diff := order.Transaction.Gross - ttTotal
		if model.Abs(diff) > feeTolerance {
			mismatches = append(mismatches, model.MismatchRecord{
				TransactionID: order.Transaction.TransactionID,
				PayPalGross:   order.Transaction.Gross,
				TTValue:       ttTotal,
				Diff:          diff,
			})
		}
	}

	return mismatches
}
