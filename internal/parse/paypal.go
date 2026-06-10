package parse

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jsilence82/ssg-reconcile/internal/config"
	"github.com/jsilence82/ssg-reconcile/internal/model"
)

// PayPal date format used in exports: MM/DD/YYYY
const paypalDateLayout = "01/02/2006"

// PayPal parses a PayPal CSV export, strips PII, and returns the transactions.
// If writeStripped is true, a _stripped.csv copy is written alongside the input.
func PayPal(cfg *config.Config, path string, writeStripped bool) ([]model.PayPalTransaction, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening PayPal CSV %q: %w", path, err)
	}
	defer f.Close()

	headers, records, err := readCSV(f, path)
	if err != nil {
		return nil, err
	}

	StripPII(headers, records, cfg.PII.PayPal)

	if writeStripped {
		dest := strippedPath(path)
		if err := WriteStripped(dest, headers, records); err != nil {
			return nil, err
		}
	}

	idx := buildIndex(headers)
	txns := make([]model.PayPalTransaction, 0, len(records))
	for rowNum, row := range records {
		lineNum := rowNum + 2 // 1-based + header row
		txn, err := parsePayPalRow(idx, row, lineNum)
		if err != nil {
			return nil, fmt.Errorf("PayPal CSV %q: %w", path, err)
		}
		if txn != nil {
			txns = append(txns, *txn)
		}
	}

	return txns, nil
}

func parsePayPalRow(idx map[string]int, row []string, lineNum int) (*model.PayPalTransaction, error) {
	get := func(col string) string {
		i, ok := idx[col]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	status := get("Status")
	if status == "" {
		return nil, nil
	}

	dateStr := get("Date")
	date, err := time.Parse(paypalDateLayout, dateStr)
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid date %q: %w", lineNum, dateStr, err)
	}

	gross, err := model.MoneyFromString(get("Gross"))
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid gross: %w", lineNum, err)
	}

	fee, err := model.MoneyFromString(get("Fee"))
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid fee: %w", lineNum, err)
	}

	net, err := model.MoneyFromString(get("Net"))
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid net: %w", lineNum, err)
	}

	return &model.PayPalTransaction{
		Date:          date,
		TransactionID: get("Transaction ID"),
		ItemTitle:     get("Item Title"),
		Gross:         gross,
		Fee:           fee,
		Net:           net,
		Status:        status,
		Currency:      get("Currency"),
		Quantity:      1,
	}, nil
}

// readCSV reads all rows from an io.Reader, returning headers and data rows.
func readCSV(r io.Reader, sourceName string) (headers []string, records [][]string, err error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	headers, err = reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("reading headers from %q: %w", sourceName, err)
	}

	records, err = reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("reading rows from %q: %w", sourceName, err)
	}

	return headers, records, nil
}

// buildIndex maps header names to column indices.
func buildIndex(headers []string) map[string]int {
	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		idx[strings.TrimSpace(h)] = i
	}
	return idx
}
