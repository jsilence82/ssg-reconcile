package model

import "time"

type RefundRecord struct {
	TransactionID string
	Date          time.Time
	Amount        Money
}

type OrphanRecord struct {
	Source        string // "paypal" or "tickettailor"
	TransactionID string
	Date          time.Time
	Amount        Money
	Detail        string
}

type MismatchRecord struct {
	TransactionID string
	PayPalGross   Money
	TTValue       Money
	Diff          Money
}

type ReconciliationReport struct {
	ShowName     string
	GeneratedAt  time.Time
	Performances []PerformanceSummary
	Totals       PerformanceSummary
	Refunds      []RefundRecord
	Orphans      []OrphanRecord
	Mismatches   []MismatchRecord
	IsClean      bool
}
