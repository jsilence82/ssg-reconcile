package reconcile

import (
	"github.com/jsilence82/ssg-reconcile/internal/config"
	"github.com/jsilence82/ssg-reconcile/internal/model"
)

const unknownEventBucket = "OTHER_PRODUCTION"

// GroupResult carries the output of the grouping stage.
type GroupResult struct {
	// ByEventID maps known event IDs (from config) to their joined orders.
	ByEventID map[string][]JoinedOrder
	// Other holds orders whose event IDs are not in the config (other productions).
	Other []JoinedOrder
	// CompTickets holds all comp (NO_COST) tickets, grouped by event ID.
	CompTickets map[string][]model.TTTicket
}

// GroupByPerformance partitions joined orders and comp tickets by event ID.
// Orders whose tickets span multiple event IDs are attributed to each event
// proportionally; the split happens in the aggregate stage.
func GroupByPerformance(cfg *config.Config, orders []JoinedOrder, allTickets []model.TTTicket) GroupResult {
	knownIDs := cfg.EventIDToPerformance()

	result := GroupResult{
		ByEventID:   make(map[string][]JoinedOrder),
		CompTickets: make(map[string][]model.TTTicket),
	}

	// Distribute comp tickets
	for _, t := range allTickets {
		if !t.IsComp() {
			continue
		}
		result.CompTickets[t.EventID] = append(result.CompTickets[t.EventID], t)
	}

	// Distribute orders: an order belongs to every event ID present in its tickets.
	// The aggregate stage handles proration for cross-performance orders.
	for _, order := range orders {
		eventIDs := uniqueEventIDs(order.Tickets)
		placed := false
		for _, eid := range eventIDs {
			if _, known := knownIDs[eid]; known {
				result.ByEventID[eid] = append(result.ByEventID[eid], order)
				placed = true
			}
		}
		if !placed {
			result.Other = append(result.Other, order)
		}
	}

	return result
}

func uniqueEventIDs(tickets []model.TTTicket) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, t := range tickets {
		if _, ok := seen[t.EventID]; !ok {
			seen[t.EventID] = struct{}{}
			ids = append(ids, t.EventID)
		}
	}
	return ids
}
