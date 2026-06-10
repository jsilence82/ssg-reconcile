package reconcile

import (
	"github.com/jsilence82/ssg-reconcile/internal/model"
)

// JoinedOrder holds one PayPal transaction and all its associated TT tickets.
type JoinedOrder struct {
	Transaction model.PayPalTransaction
	Tickets     []model.TTTicket
}

// JoinResult carries the output of the join stage.
type JoinResult struct {
	Orders      []JoinedOrder
	OrphanPayPal []model.PayPalTransaction // PayPal txns with no TT match
	OrphanTT    []model.TTTicket           // TT tickets (PAYPAL method) with no PayPal match
	BlankTTID   []model.TTTicket           // TT tickets with PAYPAL method but empty Transaction ID
}

// Join matches PayPal transactions with Ticket Tailor tickets by Transaction ID.
// Comp tickets (NO_COST) are excluded from joining and returned in the result unchanged.
func Join(txns []model.PayPalTransaction, tickets []model.TTTicket) JoinResult {
	// Index PayPal transactions by ID
	txnByID := make(map[string]model.PayPalTransaction, len(txns))
	for _, t := range txns {
		if t.TransactionID != "" {
			txnByID[t.TransactionID] = t
		}
	}

	// Index TT tickets by Transaction ID
	ttByTxnID := make(map[string][]model.TTTicket)
	var result JoinResult

	for _, ticket := range tickets {
		if ticket.IsComp() {
			continue
		}
		if ticket.TransactionID == "" {
			result.BlankTTID = append(result.BlankTTID, ticket)
			continue
		}
		ttByTxnID[ticket.TransactionID] = append(ttByTxnID[ticket.TransactionID], ticket)
	}

	// Build joined orders; detect orphaned PayPal transactions
	matchedTxnIDs := make(map[string]bool)
	for _, txn := range txns {
		ttTickets, ok := ttByTxnID[txn.TransactionID]
		if !ok || len(ttTickets) == 0 {
			result.OrphanPayPal = append(result.OrphanPayPal, txn)
			continue
		}
		result.Orders = append(result.Orders, JoinedOrder{
			Transaction: txn,
			Tickets:     ttTickets,
		})
		matchedTxnIDs[txn.TransactionID] = true
	}

	// Detect orphaned TT tickets (have a Transaction ID but no matching PayPal txn)
	for txnID, ttTickets := range ttByTxnID {
		if !matchedTxnIDs[txnID] {
			result.OrphanTT = append(result.OrphanTT, ttTickets...)
		}
	}

	return result
}
