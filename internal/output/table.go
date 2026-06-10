package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jsilence82/ssg-reconcile/internal/model"
)

// TableRenderer writes a plain-text table to an io.Writer.
type TableRenderer struct {
	W io.Writer
}

// NewTableRenderer returns a TableRenderer writing to stdout.
func NewTableRenderer() *TableRenderer {
	return &TableRenderer{W: os.Stdout}
}

func (r *TableRenderer) Render(report *model.ReconciliationReport) error {
	w := r.W

	fmt.Fprintf(w, "SSG Ticket Reconciliation — %s\n", report.ShowName)
	fmt.Fprintf(w, "Generated: %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Financial summary table
	const finFmt = "%-13s %-12s %-14s %-10s %-10s %-10s\n"
	divider := strings.Repeat("─", 65)

	fmt.Fprintf(w, finFmt, "Performance", "Date", "Transactions", "Gross", "Fees", "Net")
	fmt.Fprintln(w, divider)
	for _, p := range report.Performances {
		dateStr := ""
		if !p.Date.IsZero() {
			dateStr = p.Date.Format("2006-01-02")
		}
		fmt.Fprintf(w, finFmt,
			fmt.Sprintf("%d", p.PerformanceNumber),
			dateStr,
			fmt.Sprintf("%d", p.TransactionCount),
			fmt.Sprintf("%.2f", p.Gross.Euros()),
			fmt.Sprintf("%.2f", p.Fees.Euros()),
			fmt.Sprintf("%.2f", p.Net.Euros()),
		)
	}
	fmt.Fprintln(w, divider)
	t := report.Totals
	fmt.Fprintf(w, finFmt, "TOTAL", "", fmt.Sprintf("%d", t.TransactionCount),
		fmt.Sprintf("%.2f", t.Gross.Euros()),
		fmt.Sprintf("%.2f", t.Fees.Euros()),
		fmt.Sprintf("%.2f", t.Net.Euros()),
	)

	// Ticket counts table
	fmt.Fprintf(w, "\nTicket Counts:\n")
	const cntFmt = "%-13s %-9s %-9s %-6s %-6s\n"
	fmt.Fprintf(w, cntFmt, "Performance", "General", "Student", "Comp", "Total")
	fmt.Fprintln(w, strings.Repeat("─", 44))
	for _, p := range report.Performances {
		fmt.Fprintf(w, cntFmt,
			fmt.Sprintf("%d", p.PerformanceNumber),
			fmt.Sprintf("%d", p.TicketCounts[model.CategoryGeneral]),
			fmt.Sprintf("%d", p.TicketCounts[model.CategoryStudent]),
			fmt.Sprintf("%d", p.TicketCounts[model.CategoryComp]),
			fmt.Sprintf("%d", p.TotalTickets()),
		)
	}
	fmt.Fprintln(w, strings.Repeat("─", 44))
	fmt.Fprintf(w, cntFmt, "TOTAL",
		fmt.Sprintf("%d", t.TicketCounts[model.CategoryGeneral]),
		fmt.Sprintf("%d", t.TicketCounts[model.CategoryStudent]),
		fmt.Sprintf("%d", t.TicketCounts[model.CategoryComp]),
		fmt.Sprintf("%d", t.TotalTickets()),
	)

	// Issues
	r.renderIssues(w, report)

	// Status
	fmt.Fprintln(w)
	if report.IsClean {
		fmt.Fprintln(w, "Status: ✓ CLEAN — all PayPal transactions matched, no discrepancies")
	} else {
		fmt.Fprintln(w, "Status: ✗ ISSUES FOUND — see warnings above")
	}

	return nil
}

func (r *TableRenderer) renderIssues(w io.Writer, report *model.ReconciliationReport) {
	if len(report.Refunds) > 0 {
		fmt.Fprintf(w, "\nRefunds (%d):\n", len(report.Refunds))
		for _, rf := range report.Refunds {
			fmt.Fprintf(w, "  %-24s  %s  €%.2f\n",
				rf.TransactionID, rf.Date.Format("2006-01-02"), rf.Amount.Euros())
		}
	}

	if len(report.Orphans) > 0 {
		fmt.Fprintf(w, "\n⚠ ORPHANED PAYPAL TRANSACTIONS (no matching Ticket Tailor entry):\n")
		for _, o := range report.Orphans {
			fmt.Fprintf(w, "  %-24s  %s  €%.2f\n",
				o.TransactionID, o.Date.Format("2006-01-02"), o.Amount.Euros())
		}
	}

	if len(report.Mismatches) > 0 {
		fmt.Fprintf(w, "\n⚠ AMOUNT MISMATCHES:\n")
		for _, m := range report.Mismatches {
			fmt.Fprintf(w, "  %s: PayPal gross €%.2f ≠ TT order value €%.2f  (diff: €%.2f)\n",
				m.TransactionID, m.PayPalGross.Euros(), m.TTValue.Euros(), m.Diff.Euros())
		}
	}
}
