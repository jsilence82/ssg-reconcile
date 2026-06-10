package output

import (
	"fmt"

	"github.com/jsilence82/ssg-reconcile/internal/model"
	"github.com/xuri/excelize/v2"
)

// ExcelRenderer writes the reconciliation report as an xlsx workbook.
type ExcelRenderer struct {
	Path string
}

func NewExcelRenderer(path string) *ExcelRenderer {
	return &ExcelRenderer{Path: path}
}

func (r *ExcelRenderer) Render(report *model.ReconciliationReport) error {
	f := excelize.NewFile()
	defer f.Close()

	if err := r.writeFinancialSheet(f, report); err != nil {
		return err
	}
	if err := r.writeTicketCountSheet(f, report); err != nil {
		return err
	}
	if len(report.Refunds) > 0 || len(report.Orphans) > 0 || len(report.Mismatches) > 0 {
		if err := r.writeIssuesSheet(f, report); err != nil {
			return err
		}
	}

	// Remove default empty sheet
	f.DeleteSheet("Sheet1")

	if err := f.SaveAs(r.Path); err != nil {
		return fmt.Errorf("saving Excel file %q: %w", r.Path, err)
	}

	return nil
}

func (r *ExcelRenderer) writeFinancialSheet(f *excelize.File, report *model.ReconciliationReport) error {
	sheet := "Financial Summary"
	f.NewSheet(sheet)

	headers := []string{"Performance", "Date", "Transactions", "Gross (€)", "Fees (€)", "Net (€)"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	row := 2
	for _, p := range report.Performances {
		dateStr := ""
		if !p.Date.IsZero() {
			dateStr = p.Date.Format("2006-01-02")
		}
		vals := []interface{}{
			p.PerformanceNumber, dateStr, p.TransactionCount,
			p.Gross.Euros(), p.Fees.Euros(), p.Net.Euros(),
		}
		for col, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, v)
		}
		row++
	}

	// Totals
	t := report.Totals
	totals := []interface{}{
		"TOTAL", "", t.TransactionCount,
		t.Gross.Euros(), t.Fees.Euros(), t.Net.Euros(),
	}
	for col, v := range totals {
		cell, _ := excelize.CoordinatesToCellName(col+1, row)
		f.SetCellValue(sheet, cell, v)
	}

	return nil
}

func (r *ExcelRenderer) writeTicketCountSheet(f *excelize.File, report *model.ReconciliationReport) error {
	sheet := "Ticket Counts"
	f.NewSheet(sheet)

	headers := []string{"Performance", "Date", "General", "Student", "Comp", "Total"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	row := 2
	for _, p := range report.Performances {
		dateStr := ""
		if !p.Date.IsZero() {
			dateStr = p.Date.Format("2006-01-02")
		}
		vals := []interface{}{
			p.PerformanceNumber, dateStr,
			p.TicketCounts[model.CategoryGeneral],
			p.TicketCounts[model.CategoryStudent],
			p.TicketCounts[model.CategoryComp],
			p.TotalTickets(),
		}
		for col, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, v)
		}
		row++
	}

	t := report.Totals
	totals := []interface{}{
		"TOTAL", "",
		t.TicketCounts[model.CategoryGeneral],
		t.TicketCounts[model.CategoryStudent],
		t.TicketCounts[model.CategoryComp],
		t.TotalTickets(),
	}
	for col, v := range totals {
		cell, _ := excelize.CoordinatesToCellName(col+1, row)
		f.SetCellValue(sheet, cell, v)
	}

	return nil
}

func (r *ExcelRenderer) writeIssuesSheet(f *excelize.File, report *model.ReconciliationReport) error {
	sheet := "Issues"
	f.NewSheet(sheet)

	currentRow := 1

	if len(report.Refunds) > 0 {
		f.SetCellValue(sheet, cellName(1, currentRow), "REFUNDS")
		currentRow++
		for _, rf := range report.Refunds {
			f.SetCellValue(sheet, cellName(1, currentRow), rf.TransactionID)
			f.SetCellValue(sheet, cellName(2, currentRow), rf.Date.Format("2006-01-02"))
			f.SetCellValue(sheet, cellName(3, currentRow), rf.Amount.Euros())
			currentRow++
		}
		currentRow++
	}

	if len(report.Orphans) > 0 {
		f.SetCellValue(sheet, cellName(1, currentRow), "ORPHANED PAYPAL TRANSACTIONS")
		currentRow++
		for _, o := range report.Orphans {
			f.SetCellValue(sheet, cellName(1, currentRow), o.TransactionID)
			f.SetCellValue(sheet, cellName(2, currentRow), o.Date.Format("2006-01-02"))
			f.SetCellValue(sheet, cellName(3, currentRow), o.Amount.Euros())
			currentRow++
		}
		currentRow++
	}

	if len(report.Mismatches) > 0 {
		f.SetCellValue(sheet, cellName(1, currentRow), "AMOUNT MISMATCHES")
		currentRow++
		for _, m := range report.Mismatches {
			f.SetCellValue(sheet, cellName(1, currentRow), m.TransactionID)
			f.SetCellValue(sheet, cellName(2, currentRow), m.PayPalGross.Euros())
			f.SetCellValue(sheet, cellName(3, currentRow), m.TTValue.Euros())
			f.SetCellValue(sheet, cellName(4, currentRow), m.Diff.Euros())
			currentRow++
		}
	}

	return nil
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
