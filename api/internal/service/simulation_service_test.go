package service

import (
	"testing"

	"rev-saas-api/internal/config"
	"rev-saas-api/internal/model"
)

// loadTestConfig loads the elasticity config for testing.
func loadTestConfig(t *testing.T) *config.ElasticityConfig {
	cfg, err := config.LoadElasticityConfig()
	if err != nil {
		t.Fatalf("failed to load elasticity config: %v", err)
	}
	return cfg
}

// TestFindBucket tests the bucket finding logic.
func TestFindBucket(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name           string
		priceChangePct float64
		wantNil        bool
		wantDesc       string
	}{
		{"large decrease", -25, false, "Large price decrease (>20%)"},
		{"moderate decrease", -15, false, "Moderate price decrease (10-20%)"},
		{"small decrease", -5, false, "Small price decrease (0-10%)"},
		{"small increase", 5, false, "Small price increase (0-10%)"},
		{"moderate increase", 15, false, "Moderate price increase (10-20%)"},
		{"large increase", 30, false, "Large price increase (>20%)"},
		{"zero change", 0, false, "Small price increase (0-10%)"},
		{"boundary 10", 10, false, "Moderate price increase (10-20%)"},
		{"boundary -10", -10, false, "Small price decrease (0-10%)"}, // -10 is at the boundary, falls into [−10, 0) bucket
		{"boundary -10.01", -10.01, false, "Moderate price decrease (10-20%)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := cfg.FindBucket(tt.priceChangePct)
			if tt.wantNil {
				if bucket != nil {
					t.Errorf("expected nil bucket, got %v", bucket.Description)
				}
			} else {
				if bucket == nil {
					t.Errorf("expected bucket, got nil")
				} else if bucket.Description != tt.wantDesc {
					t.Errorf("expected description %q, got %q", tt.wantDesc, bucket.Description)
				}
			}
		})
	}
}

// TestGetChurnMultiplier tests the churn multiplier logic.
func TestGetChurnMultiplier(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name         string
		churnRate    float64
		wantMultiple float64
	}{
		{"high churn", 10.0, 1.3},
		{"low churn", 2.0, 0.8},
		{"medium churn", 6.0, 1.0},
		{"at high threshold", 8.0, 1.3},
		{"at low threshold", 4.0, 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetChurnMultiplier(tt.churnRate)
			if got != tt.wantMultiple {
				t.Errorf("GetChurnMultiplier(%v) = %v, want %v", tt.churnRate, got, tt.wantMultiple)
			}
		})
	}
}

// TestDeriveRiskLevel tests the risk level derivation logic.
func TestDeriveRiskLevel(t *testing.T) {
	cfg := loadTestConfig(t)

	tests := []struct {
		name          string
		absPriceChange float64
		scenarioLevel string
		wantRisk      string
	}{
		{"small change conservative", 5, "conservative", "low"},
		{"small change base", 5, "base", "low"},
		{"small change aggressive", 5, "aggressive", "medium"},
		{"medium change conservative", 15, "conservative", "medium"},
		{"medium change base", 15, "base", "medium"},
		{"medium change aggressive", 15, "aggressive", "high"},
		{"large change conservative", 30, "conservative", "medium"},
		{"large change base", 30, "base", "high"},
		{"large change aggressive", 30, "aggressive", "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.DeriveRiskLevel(tt.absPriceChange, tt.scenarioLevel)
			if got != tt.wantRisk {
				t.Errorf("DeriveRiskLevel(%v, %v) = %v, want %v", tt.absPriceChange, tt.scenarioLevel, got, tt.wantRisk)
			}
		})
	}
}

// TestNormalizePricingGoal tests the pricing goal normalization.
func TestNormalizePricingGoal(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"revenue", "revenue"},
		{"retention", "retention"},
		{"conversion", "conversion"},
		{"differentiation", "differentiation"},
		{"unknown", "revenue"},
		{"", "revenue"},
		{"REVENUE", "revenue"}, // should fallback, case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := model.NormalizePricingGoal(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePricingGoal(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestCalculateScenario_PriceIncrease tests scenario calculation for price increases.
func TestCalculateScenario_PriceIncrease(t *testing.T) {
	cfg := loadTestConfig(t)
	
	// Create a minimal simulation service for testing
	svc := &SimulationService{
		elasticityConfig: cfg,
	}

	// Test case: 15% price increase with revenue goal
	bucket := cfg.FindBucket(15)
	if bucket == nil {
		t.Fatal("expected bucket for 15% change")
	}

	profile := bucket.Profiles["revenue"]
	req := model.SimulationRequest{
		CurrentPrice:          100,
		NewPrice:              115,
		ActiveCustomersOnPlan: 1000,
		GlobalChurnRate:       5.0, // medium churn → multiplier = 1.0
	}

	scenario := svc.calculateScenario("Base", profile.Base, req, true, 1.0, 15)

	// Check that customer loss percentages match config
	expectedLossMin := profile.Base.CustomerLossMinPct // 5%
	expectedLossMax := profile.Base.CustomerLossMaxPct // 10%

	if scenario.CustomerLossMinPct != expectedLossMin {
		t.Errorf("CustomerLossMinPct = %v, want %v", scenario.CustomerLossMinPct, expectedLossMin)
	}
	if scenario.CustomerLossMaxPct != expectedLossMax {
		t.Errorf("CustomerLossMaxPct = %v, want %v", scenario.CustomerLossMaxPct, expectedLossMax)
	}

	// Check customer count calculations
	// Min loss = 5% → lose 50 customers → 950 customers remain (this is max count)
	// Max loss = 10% → lose 100 customers → 900 customers remain (this is min count)
	if scenario.NewCustomerCountMax != 950 {
		t.Errorf("NewCustomerCountMax = %v, want 950", scenario.NewCustomerCountMax)
	}
	if scenario.NewCustomerCountMin != 900 {
		t.Errorf("NewCustomerCountMin = %v, want 900", scenario.NewCustomerCountMin)
	}

	// Check MRR calculations (at new price of $115)
	expectedMRRMin := 900 * 115.0  // $103,500
	expectedMRRMax := 950 * 115.0  // $109,250

	if scenario.NewMRRMin != expectedMRRMin {
		t.Errorf("NewMRRMin = %v, want %v", scenario.NewMRRMin, expectedMRRMin)
	}
	if scenario.NewMRRMax != expectedMRRMax {
		t.Errorf("NewMRRMax = %v, want %v", scenario.NewMRRMax, expectedMRRMax)
	}

	// Check ARR = MRR * 12
	if scenario.NewARRMin != expectedMRRMin*12 {
		t.Errorf("NewARRMin = %v, want %v", scenario.NewARRMin, expectedMRRMin*12)
	}
	if scenario.NewARRMax != expectedMRRMax*12 {
		t.Errorf("NewARRMax = %v, want %v", scenario.NewARRMax, expectedMRRMax*12)
	}
}

// TestCalculateScenario_PriceDecrease tests scenario calculation for price decreases.
func TestCalculateScenario_PriceDecrease(t *testing.T) {
	cfg := loadTestConfig(t)
	
	svc := &SimulationService{
		elasticityConfig: cfg,
	}

	// Test case: 15% price decrease with conversion goal
	bucket := cfg.FindBucket(-15)
	if bucket == nil {
		t.Fatal("expected bucket for -15% change")
	}

	profile := bucket.Profiles["conversion"]
	req := model.SimulationRequest{
		CurrentPrice:          100,
		NewPrice:              85,
		ActiveCustomersOnPlan: 1000,
		GlobalChurnRate:       5.0,
	}

	scenario := svc.calculateScenario("Base", profile.Base, req, false, 1.0, 15)

	// Check that customer gain percentages are set
	expectedGainMin := profile.Base.CustomerGainMinPct // 12%
	expectedGainMax := profile.Base.CustomerGainMaxPct // 20%

	if scenario.CustomerGainMinPct != expectedGainMin {
		t.Errorf("CustomerGainMinPct = %v, want %v", scenario.CustomerGainMinPct, expectedGainMin)
	}
	if scenario.CustomerGainMaxPct != expectedGainMax {
		t.Errorf("CustomerGainMaxPct = %v, want %v", scenario.CustomerGainMaxPct, expectedGainMax)
	}

	// Check customer count calculations
	// Min gain = 12% → gain 120 customers → 1120 total (this is min count)
	// Max gain = 20% → gain 200 customers → 1200 total (this is max count)
	if scenario.NewCustomerCountMin != 1120 {
		t.Errorf("NewCustomerCountMin = %v, want 1120", scenario.NewCustomerCountMin)
	}
	if scenario.NewCustomerCountMax != 1200 {
		t.Errorf("NewCustomerCountMax = %v, want 1200", scenario.NewCustomerCountMax)
	}
}

// TestCalculateScenario_HighChurnMultiplier tests that high churn increases loss bands.
func TestCalculateScenario_HighChurnMultiplier(t *testing.T) {
	cfg := loadTestConfig(t)
	
	svc := &SimulationService{
		elasticityConfig: cfg,
	}

	bucket := cfg.FindBucket(15) // 15% price increase
	if bucket == nil {
		t.Fatal("expected bucket for 15% change")
	}

	profile := bucket.Profiles["revenue"]
	req := model.SimulationRequest{
		CurrentPrice:          100,
		NewPrice:              115,
		ActiveCustomersOnPlan: 1000,
		GlobalChurnRate:       10.0, // high churn
	}

	// High churn multiplier = 1.3
	churnMultiplier := cfg.GetChurnMultiplier(10.0)
	if churnMultiplier != 1.3 {
		t.Fatalf("expected churn multiplier 1.3, got %v", churnMultiplier)
	}

	scenario := svc.calculateScenario("Base", profile.Base, req, true, churnMultiplier, 15)

	// With 1.3x multiplier, loss percentages should be higher
	expectedLossMin := profile.Base.CustomerLossMinPct * 1.3 // 5% * 1.3 = 6.5%
	expectedLossMax := profile.Base.CustomerLossMaxPct * 1.3 // 10% * 1.3 = 13%

	if scenario.CustomerLossMinPct != expectedLossMin {
		t.Errorf("CustomerLossMinPct with high churn = %v, want %v", scenario.CustomerLossMinPct, expectedLossMin)
	}
	if scenario.CustomerLossMaxPct != expectedLossMax {
		t.Errorf("CustomerLossMaxPct with high churn = %v, want %v", scenario.CustomerLossMaxPct, expectedLossMax)
	}
}

// TestCalculateScenario_LowChurnMultiplier tests that low churn decreases loss bands.
func TestCalculateScenario_LowChurnMultiplier(t *testing.T) {
	cfg := loadTestConfig(t)
	
	svc := &SimulationService{
		elasticityConfig: cfg,
	}

	bucket := cfg.FindBucket(15)
	if bucket == nil {
		t.Fatal("expected bucket for 15% change")
	}

	profile := bucket.Profiles["revenue"]
	req := model.SimulationRequest{
		CurrentPrice:          100,
		NewPrice:              115,
		ActiveCustomersOnPlan: 1000,
		GlobalChurnRate:       2.0, // low churn
	}

	// Low churn multiplier = 0.8
	churnMultiplier := cfg.GetChurnMultiplier(2.0)
	if churnMultiplier != 0.8 {
		t.Fatalf("expected churn multiplier 0.8, got %v", churnMultiplier)
	}

	scenario := svc.calculateScenario("Base", profile.Base, req, true, churnMultiplier, 15)

	// With 0.8x multiplier, loss percentages should be lower
	expectedLossMin := profile.Base.CustomerLossMinPct * 0.8 // 5% * 0.8 = 4%
	expectedLossMax := profile.Base.CustomerLossMaxPct * 0.8 // 10% * 0.8 = 8%

	if scenario.CustomerLossMinPct != expectedLossMin {
		t.Errorf("CustomerLossMinPct with low churn = %v, want %v", scenario.CustomerLossMinPct, expectedLossMin)
	}
	if scenario.CustomerLossMaxPct != expectedLossMax {
		t.Errorf("CustomerLossMaxPct with low churn = %v, want %v", scenario.CustomerLossMaxPct, expectedLossMax)
	}
}

// TestDeterminism tests that the same input always produces the same output.
func TestDeterminism(t *testing.T) {
	cfg := loadTestConfig(t)
	
	svc := &SimulationService{
		elasticityConfig: cfg,
	}

	bucket := cfg.FindBucket(15)
	profile := bucket.Profiles["revenue"]
	
	req := model.SimulationRequest{
		CurrentPrice:          100,
		NewPrice:              115,
		ActiveCustomersOnPlan: 1000,
		GlobalChurnRate:       5.0,
	}

	// Run the same calculation multiple times
	var results []model.SimulationScenario
	for i := 0; i < 10; i++ {
		scenario := svc.calculateScenario("Base", profile.Base, req, true, 1.0, 15)
		results = append(results, scenario)
	}

	// All results should be identical
	first := results[0]
	for i, r := range results[1:] {
		if r.CustomerLossMinPct != first.CustomerLossMinPct ||
			r.CustomerLossMaxPct != first.CustomerLossMaxPct ||
			r.NewCustomerCountMin != first.NewCustomerCountMin ||
			r.NewCustomerCountMax != first.NewCustomerCountMax ||
			r.NewMRRMin != first.NewMRRMin ||
			r.NewMRRMax != first.NewMRRMax ||
			r.RiskLevel != first.RiskLevel {
			t.Errorf("Result %d differs from first result", i+1)
		}
	}
}

// TestAllScenarioLevels tests that all three scenario levels are properly differentiated.
func TestAllScenarioLevels(t *testing.T) {
	cfg := loadTestConfig(t)
	
	svc := &SimulationService{
		elasticityConfig: cfg,
	}

	bucket := cfg.FindBucket(15) // 15% price increase
	profile := bucket.Profiles["revenue"]
	
	req := model.SimulationRequest{
		CurrentPrice:          100,
		NewPrice:              115,
		ActiveCustomersOnPlan: 1000,
		GlobalChurnRate:       5.0,
	}

	conservative := svc.calculateScenario("Conservative", profile.Conservative, req, true, 1.0, 15)
	base := svc.calculateScenario("Base", profile.Base, req, true, 1.0, 15)
	aggressive := svc.calculateScenario("Aggressive", profile.Aggressive, req, true, 1.0, 15)

	// For price increase, conservative should have smallest loss, aggressive largest
	if conservative.CustomerLossMaxPct >= base.CustomerLossMaxPct {
		t.Error("Conservative should have smaller max loss than Base")
	}
	if base.CustomerLossMaxPct >= aggressive.CustomerLossMaxPct {
		t.Error("Base should have smaller max loss than Aggressive")
	}

	// Conservative should have more remaining customers than aggressive
	if conservative.NewCustomerCountMin <= aggressive.NewCustomerCountMin {
		t.Error("Conservative should have more remaining customers than Aggressive")
	}
}

