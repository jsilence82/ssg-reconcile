package cmd

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/jsilence82/ssg-reconcile/internal/config"
	"github.com/jsilence82/ssg-reconcile/internal/model"
	"github.com/jsilence82/ssg-reconcile/internal/output"
	"github.com/jsilence82/ssg-reconcile/internal/parse"
	"github.com/jsilence82/ssg-reconcile/internal/reconcile"
)

var rootCmd = &cobra.Command{
	Use:   "ssg-reconcile",
	Short: "Reconcile SSG ticket sales between Ticket Tailor and PayPal",
	RunE:  runReconcile,
}

var stripCmd = &cobra.Command{
	Use:   "strip",
	Short: "Strip PII from CSV exports only; do not reconcile",
	RunE:  runStrip,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Run reconciliation and report issues only (no summary output)",
	RunE:  runValidate,
}

// Shared flags
var (
	flagPayPal    string
	flagTickets   string
	flagConfig    string
	flagOutput    string
	flagOutFile   string
	flagStrip     bool
	flagStrict    bool
	flagVerbose   bool
)

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&flagPayPal, "paypal", "", "Path to raw PayPal CSV export (required)")
	pf.StringVar(&flagTickets, "tickets", "", "Path to raw Ticket Tailor CSV export (required)")
	pf.StringVar(&flagConfig, "config", "", "Path to config file (default: ./ssg-reconcile.yaml)")
	pf.BoolVar(&flagVerbose, "verbose", false, "Print per-transaction detail to stderr")

	rootCmd.Flags().StringVar(&flagOutput, "output", "table", "Output format: table|csv|excel")
	rootCmd.Flags().StringVar(&flagOutFile, "out-file", "", "Output file path (required for csv and excel)")
	rootCmd.Flags().BoolVar(&flagStrip, "strip", true, "Write PII-stripped copies of input CSVs")
	rootCmd.Flags().BoolVar(&flagStrict, "strict", false, "Exit non-zero if any orphans or mismatches found")

	validateCmd.Flags().BoolVar(&flagStrict, "strict", false, "Exit non-zero if any issues found")

	rootCmd.AddCommand(stripCmd, validateCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runReconcile(cmd *cobra.Command, args []string) error {
	if err := requireFlags("paypal", "tickets"); err != nil {
		return err
	}
	if (flagOutput == "csv" || flagOutput == "excel") && flagOutFile == "" {
		return fmt.Errorf("--out-file is required when --output is %s", flagOutput)
	}

	cfg, err := config.Load(flagConfig)
	if err != nil {
		return err
	}

	report, err := buildReport(cfg, flagStrip)
	if err != nil {
		return err
	}

	renderer, err := newRenderer(flagOutput, flagOutFile)
	if err != nil {
		return err
	}

	if err := renderer.Render(report); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	if flagStrict && !report.IsClean {
		return fmt.Errorf("reconciliation has issues (--strict mode)")
	}

	return nil
}

func runStrip(cmd *cobra.Command, args []string) error {
	if err := requireFlags("paypal", "tickets"); err != nil {
		return err
	}

	cfg, err := config.Load(flagConfig)
	if err != nil {
		return err
	}

	g := new(errgroup.Group)
	g.Go(func() error { _, err := parse.PayPal(cfg, flagPayPal, true); return err })
	g.Go(func() error { _, err := parse.TicketTailor(cfg, flagTickets, true); return err })
	return g.Wait()
}

func runValidate(cmd *cobra.Command, args []string) error {
	if err := requireFlags("paypal", "tickets"); err != nil {
		return err
	}

	cfg, err := config.Load(flagConfig)
	if err != nil {
		return err
	}

	report, err := buildReport(cfg, false)
	if err != nil {
		return err
	}

	if !report.IsClean {
		renderer := output.NewTableRenderer()
		if err := renderer.Render(report); err != nil {
			return err
		}
	}

	if flagStrict && !report.IsClean {
		return fmt.Errorf("reconciliation has issues (--strict mode)")
	}

	fmt.Fprintln(os.Stdout, "Validation complete.")
	return nil
}

// buildReport runs the full pipeline and returns a ReconciliationReport.
func buildReport(cfg *config.Config, writeStripped bool) (*model.ReconciliationReport, error) {
	var paypalTxns []model.PayPalTransaction
	var ttTickets []model.TTTicket

	g := new(errgroup.Group)
	g.Go(func() error {
		var err error
		paypalTxns, err = parse.PayPal(cfg, flagPayPal, writeStripped)
		return err
	})
	g.Go(func() error {
		var err error
		ttTickets, err = parse.TicketTailor(cfg, flagTickets, writeStripped)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Separate refunds before joining
	normalTxns, refunds := reconcile.ExtractRefunds(paypalTxns)

	// Join
	joinResult := reconcile.Join(normalTxns, ttTickets)

	// Validate
	feeTolerance := model.Money(int64(cfg.FeeTolerance * 100))
	mismatches := reconcile.Validate(joinResult.Orders, feeTolerance)

	// Group by performance
	groupResult := reconcile.GroupByPerformance(cfg, joinResult.Orders, ttTickets)

	// Aggregate per performance, sorted by performance number
	perfMap := cfg.EventIDToPerformance()
	type numAndID struct {
		num int
		id  string
	}
	var ordered []numAndID
	for eid, p := range perfMap {
		ordered = append(ordered, numAndID{p.Number, eid})
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].num < ordered[j].num })

	performances := make([]model.PerformanceSummary, 0, len(ordered))
	for _, ni := range ordered {
		orders := groupResult.ByEventID[ni.id]
		comps := groupResult.CompTickets[ni.id]
		summary := reconcile.Aggregate(cfg, ni.id, orders, comps)
		performances = append(performances, summary)
	}

	totals := reconcile.SumPerformances(performances)

	// Build orphan records
	orphans := make([]model.OrphanRecord, 0, len(joinResult.OrphanPayPal))
	for _, t := range joinResult.OrphanPayPal {
		orphans = append(orphans, model.OrphanRecord{
			Source:        "paypal",
			TransactionID: t.TransactionID,
			Date:          t.Date,
			Amount:        t.Gross,
		})
	}
	for _, t := range joinResult.OrphanTT {
		orphans = append(orphans, model.OrphanRecord{
			Source:        "tickettailor",
			TransactionID: t.TransactionID,
			Detail:        fmt.Sprintf("EventID=%s TicketCode=%s", t.EventID, t.TicketCode),
		})
	}

	isClean := len(orphans) == 0 && len(mismatches) == 0 && len(joinResult.BlankTTID) == 0

	return &model.ReconciliationReport{
		ShowName:     cfg.ShowName,
		GeneratedAt:  time.Now(),
		Performances: performances,
		Totals:       totals,
		Refunds:      refunds,
		Orphans:      orphans,
		Mismatches:   mismatches,
		IsClean:      isClean,
	}, nil
}

func newRenderer(format, outFile string) (output.Renderer, error) {
	switch format {
	case "table":
		return output.NewTableRenderer(), nil
	case "csv":
		return output.NewCSVRenderer(outFile), nil
	case "excel":
		return output.NewExcelRenderer(outFile), nil
	default:
		return nil, fmt.Errorf("unknown output format %q (must be table, csv, or excel)", format)
	}
}

func requireFlags(names ...string) error {
	missing := []string{}
	for _, name := range names {
		switch name {
		case "paypal":
			if flagPayPal == "" {
				missing = append(missing, "--paypal")
			}
		case "tickets":
			if flagTickets == "" {
				missing = append(missing, "--tickets")
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required flags not set: %v", missing)
	}
	return nil
}
