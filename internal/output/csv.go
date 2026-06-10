package output

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/jsilence82/ssg-reconcile/internal/model"
)

// CSVRenderer writes the reconciliation summary as a CSV file.
type CSVRenderer struct {
	Path string
}

func NewCSVRenderer(path string) *CSVRenderer {
	return &CSVRenderer{Path: path}
}

func (r *CSVRenderer) Render(report *model.ReconciliationReport) error {
	f, err := os.Create(r.Path)
	if err != nil {
		return fmt.Errorf("creating CSV output %q: %w", r.Path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	if err := w.Write([]string{
		"Performance", "Date", "Transactions",
		"Gross", "Fees", "Net",
		"General", "Student", "Comp", "Total Tickets",
	}); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	for _, p := range report.Performances {
		dateStr := ""
		if !p.Date.IsZero() {
			dateStr = p.Date.Format("2006-01-02")
		}
		row := []string{
			fmt.Sprintf("%d", p.PerformanceNumber),
			dateStr,
			fmt.Sprintf("%d", p.TransactionCount),
			fmt.Sprintf("%.2f", p.Gross.Euros()),
			fmt.Sprintf("%.2f", p.Fees.Euros()),
			fmt.Sprintf("%.2f", p.Net.Euros()),
			fmt.Sprintf("%d", p.TicketCounts[model.CategoryGeneral]),
			fmt.Sprintf("%d", p.TicketCounts[model.CategoryStudent]),
			fmt.Sprintf("%d", p.TicketCounts[model.CategoryComp]),
			fmt.Sprintf("%d", p.TotalTickets()),
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("writing CSV row: %w", err)
		}
	}

	// Totals row
	t := report.Totals
	totalsRow := []string{
		"TOTAL", "",
		fmt.Sprintf("%d", t.TransactionCount),
		fmt.Sprintf("%.2f", t.Gross.Euros()),
		fmt.Sprintf("%.2f", t.Fees.Euros()),
		fmt.Sprintf("%.2f", t.Net.Euros()),
		fmt.Sprintf("%d", t.TicketCounts[model.CategoryGeneral]),
		fmt.Sprintf("%d", t.TicketCounts[model.CategoryStudent]),
		fmt.Sprintf("%d", t.TicketCounts[model.CategoryComp]),
		fmt.Sprintf("%d", t.TotalTickets()),
	}
	if err := w.Write(totalsRow); err != nil {
		return fmt.Errorf("writing CSV totals row: %w", err)
	}

	return w.Error()
}
