package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type PerformanceConfig struct {
	Number  int    `mapstructure:"number"`
	EventID string `mapstructure:"event_id"`
	Date    string `mapstructure:"date"`
}

func (p PerformanceConfig) ParsedDate() (time.Time, error) {
	return time.Parse("2006-01-02", p.Date)
}

type TicketCategoryConfig struct {
	General string `mapstructure:"general"`
	Student string `mapstructure:"student"`
	Comp    string `mapstructure:"comp"`
}

type PIIConfig struct {
	PayPal       []string `mapstructure:"paypal"`
	TicketTailor []string `mapstructure:"ticket_tailor"`
}

type Config struct {
	ShowName         string               `mapstructure:"show_name"`
	Performances     []PerformanceConfig  `mapstructure:"performances"`
	TicketCategories TicketCategoryConfig `mapstructure:"ticket_categories"`
	PII              PIIConfig            `mapstructure:"pii"`
	FeeTolerance     float64              `mapstructure:"fee_tolerance"`
	Currency         string               `mapstructure:"currency"`
}

// EventIDToPerformance returns a map from event_id to PerformanceConfig.
func (c *Config) EventIDToPerformance() map[string]PerformanceConfig {
	m := make(map[string]PerformanceConfig, len(c.Performances))
	for _, p := range c.Performances {
		m[p.EventID] = p
	}
	return m
}

// Load reads config from path; falls back to default search locations if path is empty.
func Load(path string) (*Config, error) {
	v := viper.New()
	applyDefaults(v)

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("ssg-reconcile")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/ssg-reconcile")
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if path != "" || !errors.As(err, &notFound) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

func applyDefaults(v *viper.Viper) {
	v.SetDefault("fee_tolerance", 0.02)
	v.SetDefault("currency", "EUR")
	v.SetDefault("ticket_categories.general", DefaultCategoryGeneral)
	v.SetDefault("ticket_categories.student", DefaultCategoryStudent)
	v.SetDefault("ticket_categories.comp", DefaultCategoryComp)
	v.SetDefault("pii.paypal", DefaultPayPalPIIColumns)
	v.SetDefault("pii.ticket_tailor", DefaultTTPIIColumns)
}
