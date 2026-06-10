package parse

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jsilence82/ssg-reconcile/internal/config"
	"github.com/jsilence82/ssg-reconcile/internal/model"
)

// Ticket Tailor date format: "Fri 29 May 2026 20:00"
const ttDateLayout = "Mon 2 Jan 2006 15:04"

// TicketTailor parses a Ticket Tailor CSV export, strips PII, and returns the tickets.
// If writeStripped is true, a _stripped.csv copy is written alongside the input.
func TicketTailor(cfg *config.Config, path string, writeStripped bool) ([]model.TTTicket, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening Ticket Tailor CSV %q: %w", path, err)
	}
	defer f.Close()

	headers, records, err := readCSV(f, path)
	if err != nil {
		return nil, err
	}

	StripPII(headers, records, cfg.PII.TicketTailor)

	if writeStripped {
		dest := strippedPath(path)
		if err := WriteStripped(dest, headers, records); err != nil {
			return nil, err
		}
	}

	idx := buildIndex(headers)
	catMap := buildCategoryMap(cfg)

	tickets := make([]model.TTTicket, 0, len(records))
	for rowNum, row := range records {
		lineNum := rowNum + 2
		ticket, err := parseTTRow(cfg, idx, catMap, row, lineNum)
		if err != nil {
			return nil, fmt.Errorf("Ticket Tailor CSV %q: %w", path, err)
		}
		if ticket != nil {
			tickets = append(tickets, *ticket)
		}
	}

	return tickets, nil
}

func parseTTRow(cfg *config.Config, idx map[string]int, catMap map[string]model.TicketCategory, row []string, lineNum int) (*model.TTTicket, error) {
	get := func(col string) string {
		i, ok := idx[col]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	orderID := get("Order ID")
	if orderID == "" {
		return nil, nil
	}

	eventStart, err := parseEventStart(get("Event start"))
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid event start %q: %w", lineNum, get("Event start"), err)
	}

	value, err := model.MoneyFromString(get("Ticket price"))
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid ticket price: %w", lineNum, err)
	}

	description := get("Ticket type")
	category, ok := catMap[description]
	if !ok {
		category = model.CategoryUnknown
	}

	paymentMethod := strings.ToUpper(get("Order status"))

	_ = cfg

	return &model.TTTicket{
		OrderID:       orderID,
		TransactionID: get("Payment reference"),
		EventID:       get("Event ID"),
		EventName:     get("Event"),
		EventStart:    eventStart,
		TicketCode:    get("Ticket reference"),
		Category:      category,
		Value:         value,
		PaymentMethod: paymentMethod,
		CheckedIn:     strings.EqualFold(get("Checked in"), "yes"),
		Attended:      strings.EqualFold(get("Attended"), "yes"),
	}, nil
}

func parseEventStart(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(ttDateLayout, s)
	if err != nil {
		// Try alternate format without weekday
		t, err = time.Parse("2 Jan 2006 15:04", s)
	}
	return t, err
}

func buildCategoryMap(cfg *config.Config) map[string]model.TicketCategory {
	return map[string]model.TicketCategory{
		cfg.TicketCategories.General: model.CategoryGeneral,
		cfg.TicketCategories.Student: model.CategoryStudent,
		cfg.TicketCategories.Comp:    model.CategoryComp,
	}
}
