package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed pricing_elasticity.json
var elasticityJSON []byte

// ScenarioBand represents customer loss/gain percentages for a scenario.
type ScenarioBand struct {
	CustomerLossMinPct float64 `json:"customerLossMinPct"`
	CustomerLossMaxPct float64 `json:"customerLossMaxPct"`
	CustomerGainMinPct float64 `json:"customerGainMinPct"`
	CustomerGainMaxPct float64 `json:"customerGainMaxPct"`
}

// GoalProfile contains bands for conservative, base, and aggressive scenarios.
type GoalProfile struct {
	Conservative ScenarioBand `json:"conservative"`
	Base         ScenarioBand `json:"base"`
	Aggressive   ScenarioBand `json:"aggressive"`
}

// PriceChangeBucket defines elasticity bands for a range of price change percentages.
type PriceChangeBucket struct {
	MinChangePct float64                `json:"minChangePct"`
	MaxChangePct float64                `json:"maxChangePct"`
	Description  string                 `json:"description"`
	Profiles     map[string]GoalProfile `json:"profiles"` // keyed by pricing goal
}

// ChurnAdjustment defines multipliers based on global churn rate.
type ChurnAdjustment struct {
	HighChurnThreshold  float64 `json:"highChurnThreshold"`
	HighChurnMultiplier float64 `json:"highChurnMultiplier"`
	LowChurnThreshold   float64 `json:"lowChurnThreshold"`
	LowChurnMultiplier  float64 `json:"lowChurnMultiplier"`
}

// RiskLevelThresholds defines thresholds for determining risk level.
type RiskLevelThresholds struct {
	LowMaxPct    float64 `json:"lowMaxPct"`
	MediumMaxPct float64 `json:"mediumMaxPct"`
}

// ElasticityConfig holds all configuration for pricing elasticity calculations.
type ElasticityConfig struct {
	PriceChangeBuckets    []PriceChangeBucket `json:"priceChangeBuckets"`
	ChurnAdjustment       ChurnAdjustment     `json:"churnAdjustment"`
	RiskLevelThresholds   RiskLevelThresholds `json:"riskLevelThresholds"`
}

// LoadElasticityConfig loads the embedded elasticity configuration.
func LoadElasticityConfig() (*ElasticityConfig, error) {
	var cfg ElasticityConfig
	if err := json.Unmarshal(elasticityJSON, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse elasticity config: %w", err)
	}
	return &cfg, nil
}

// FindBucket returns the bucket that matches the given price change percentage.
func (c *ElasticityConfig) FindBucket(priceChangePct float64) *PriceChangeBucket {
	for i := range c.PriceChangeBuckets {
		bucket := &c.PriceChangeBuckets[i]
		if priceChangePct >= bucket.MinChangePct && priceChangePct < bucket.MaxChangePct {
			return bucket
		}
	}
	// Edge case: exactly at maxChangePct of last bucket
	if len(c.PriceChangeBuckets) > 0 {
		lastBucket := &c.PriceChangeBuckets[len(c.PriceChangeBuckets)-1]
		if priceChangePct >= lastBucket.MinChangePct {
			return lastBucket
		}
	}
	return nil
}

// GetChurnMultiplier returns the churn multiplier based on the global churn rate.
func (c *ElasticityConfig) GetChurnMultiplier(globalChurnRate float64) float64 {
	if globalChurnRate >= c.ChurnAdjustment.HighChurnThreshold {
		return c.ChurnAdjustment.HighChurnMultiplier
	}
	if globalChurnRate <= c.ChurnAdjustment.LowChurnThreshold {
		return c.ChurnAdjustment.LowChurnMultiplier
	}
	// Linear interpolation between low and high thresholds
	// For simplicity, return 1.0 (no adjustment) for middle range
	return 1.0
}

// DeriveRiskLevel returns the risk level based on price change percentage and scenario.
func (c *ElasticityConfig) DeriveRiskLevel(absPriceChangePct float64, scenarioLevel string) string {
	// Base risk from price change magnitude
	baseRisk := "low"
	if absPriceChangePct > c.RiskLevelThresholds.MediumMaxPct {
		baseRisk = "high"
	} else if absPriceChangePct > c.RiskLevelThresholds.LowMaxPct {
		baseRisk = "medium"
	}

	// Adjust based on scenario level
	switch scenarioLevel {
	case "conservative":
		// Conservative scenarios have lower risk
		if baseRisk == "high" {
			return "medium"
		}
		return baseRisk
	case "aggressive":
		// Aggressive scenarios have higher risk
		if baseRisk == "low" {
			return "medium"
		}
		return "high"
	default: // "base"
		return baseRisk
	}
}


