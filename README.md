# ssg-reconcile

A command-line tool that automates the ticket sales reconciliation workflow for the
**Schauspielgruppe des Anglistischen Seminars Heidelberg (SSG)**, a registered
non-profit theatre group (*e.V.*) in Heidelberg, Germany.

SSG sells tickets through **Ticket Tailor** and collects payments via **PayPal**.
After each production, the treasurer must cross-reference both platforms, strip
personally identifiable information before filing, and produce per-performance
financial summaries. This tool replaces that manual process with a single command.

---

## Features

- Parses raw CSV exports from both Ticket Tailor and PayPal
- Strips PII (GDPR-compliant) before any data is written to disk
- Joins transactions across both platforms on the shared Transaction ID
- Groups and aggregates results by performance (Event ID)
- Prorates PayPal gross/fees for cross-performance orders
- Detects orphaned transactions and gross mismatches
- Handles comp tickets, refunds, and variable PayPal fees
- Outputs a plain-text table, CSV, or Excel workbook

---

## Installation

```bash
git clone https://github.com/jsilence82/ssg-reconcile.git
cd ssg-reconcile
go build -o ssg-reconcile ./...
```

Requires Go 1.21+.

---

## Usage

```
ssg-reconcile [flags] --paypal <file> --tickets <file>

Flags:
  --paypal    string   Path to raw PayPal CSV export (required)
  --tickets   string   Path to raw Ticket Tailor CSV export (required)
  --config    string   Path to config file (default: ./ssg-reconcile.yaml)
  --output    string   Output format: table|csv|excel (default: table)
  --out-file  string   Output file path (required for csv and excel)
  --strip     bool     Write PII-stripped copies of input CSVs (default: true)
  --strict    bool     Exit non-zero if any orphans or mismatches found
  --verbose   bool     Print per-transaction detail to stderr
```

### Subcommands

```bash
# Strip PII only — no reconciliation
ssg-reconcile strip --paypal exports/paypal_raw.csv --tickets exports/tt_raw.csv

# Validate only — report issues without producing a summary
ssg-reconcile validate --paypal exports/paypal_raw.csv --tickets exports/tt_raw.csv --strict
```

### Example

```bash
./ssg-reconcile \
  --paypal  exports/paypal_raw.csv \
  --tickets exports/tickettailor_raw.csv \
  --config  config/asking_strangers_2026.yaml \
  --output  excel \
  --out-file reports/reconciliation.xlsx
```

---

## Configuration

Copy `config/example.yaml` and fill in your show's details:

```yaml
show_name: "Asking Strangers the Meaning of Life"

performances:
  - number: 1
    event_id: "7871526"
    date: "2026-05-28"
  - number: 2
    event_id: "7871528"
    date: "2026-05-29"
  # ...

ticket_categories:
  general: "Admission - General Admission"
  student: "Admission - Student"
  comp:    "Admission - Comp"

pii:
  paypal:
    - "Name"
    - "Address 1"
    # ...
  ticket_tailor:
    - "Attendee name"
    - "Postcode / Zip"

fee_tolerance: 0.02
currency: "EUR"
```

The `event_id` values come from Ticket Tailor's event management page. The tool
uses them — not dates — to identify which performance a ticket belongs to.

---

## Output

```
SSG Ticket Reconciliation — Asking Strangers the Meaning of Life
Generated: 2026-06-10 14:30:00

Performance  Date        Transactions  Gross     Fees      Net
─────────────────────────────────────────────────────────────────
1            2026-05-28  28            491.00   -26.42   464.58
2            2026-05-29  34            620.00   -33.01   586.99
...
─────────────────────────────────────────────────────────────────
TOTAL                    171          3216.00  -165.53  3050.47

Ticket Counts:
Performance  General  Student  Comp  Total
────────────────────────────────────────────
1            27       14       1     42
...

Status: ✓ CLEAN — all PayPal transactions matched, no discrepancies
```

---

## Project Structure

```
ssg-reconcile/
├── main.go
├── cmd/root.go                # CLI (cobra): root, strip, validate
├── internal/
│   ├── config/                # YAML config loader + defaults
│   ├── model/                 # Domain types: Money, TTTicket, PayPalTransaction, …
│   ├── parse/                 # CSV parsers + PII stripping
│   ├── reconcile/             # Join, group, validate, aggregate
│   └── output/                # Renderer interface: table, CSV, Excel
├── config/example.yaml
└── testdata/                  # Anonymised sample CSVs
```

---

## Domain Notes

- One PayPal transaction maps to one Ticket Tailor order, which may contain
  tickets across multiple performances. Gross and fees are prorated by ticket
  value for cross-performance orders.
- Comp tickets (`NO_COST`) are counted in statistics but contribute €0 to
  financial totals and are never joined against PayPal.
- PayPal fees for the same ticket price can vary by €0.01 due to PayPal's own
  rounding. The `fee_tolerance` config key controls the mismatch threshold.
- All amounts are in EUR. No currency conversion is performed.

---

## License

MIT
