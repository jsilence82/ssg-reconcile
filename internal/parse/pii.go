package parse

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// StripPII zeroes the specified columns in-place across all records.
func StripPII(headers []string, records [][]string, columns []string) {
	stripSet := make(map[string]struct{}, len(columns))
	for _, col := range columns {
		stripSet[strings.TrimSpace(col)] = struct{}{}
	}

	var stripIdx []int
	for i, h := range headers {
		if _, ok := stripSet[strings.TrimSpace(h)]; ok {
			stripIdx = append(stripIdx, i)
		}
	}

	for _, row := range records {
		for _, idx := range stripIdx {
			if idx < len(row) {
				row[idx] = ""
			}
		}
	}
}

// WriteStripped writes headers and (already-stripped) records to destPath as CSV.
func WriteStripped(destPath string, headers []string, records [][]string) error {
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating stripped file %q: %w", destPath, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(headers); err != nil {
		return fmt.Errorf("writing stripped headers to %q: %w", destPath, err)
	}
	if err := w.WriteAll(records); err != nil {
		return fmt.Errorf("writing stripped records to %q: %w", destPath, err)
	}
	w.Flush()
	return w.Error()
}

// strippedPath returns "<base>_stripped.csv" for a given input path.
func strippedPath(path string) string {
	if strings.HasSuffix(path, ".csv") {
		return path[:len(path)-4] + "_stripped.csv"
	}
	return path + "_stripped.csv"
}
