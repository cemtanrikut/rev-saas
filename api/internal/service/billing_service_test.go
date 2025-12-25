package service

import (
	"testing"

	"rev-saas-api/internal/config"
	"rev-saas-api/internal/model"
)

func TestGetPlanKeyForPriceID(t *testing.T) {
	cfg := &config.Config{
		StripePriceStarterID:    "price_starter_123",
		StripePriceGrowthID:     "price_growth_456",
		StripePriceEnterpriseID: "price_enterprise_789",
	}

	tests := []struct {
		name     string
		priceID  string
		expected string
	}{
		{
			name:     "starter plan",
			priceID:  "price_starter_123",
			expected: "starter",
		},
		{
			name:     "growth plan",
			priceID:  "price_growth_456",
			expected: "growth",
		},
		{
			name:     "enterprise plan",
			priceID:  "price_enterprise_789",
			expected: "enterprise",
		},
		{
			name:     "unknown price ID",
			priceID:  "price_unknown",
			expected: "free",
		},
		{
			name:     "empty price ID",
			priceID:  "",
			expected: "free",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.GetPlanKeyForPriceID(tt.priceID)
			if result != tt.expected {
				t.Errorf("GetPlanKeyForPriceID(%q) = %q, want %q", tt.priceID, result, tt.expected)
			}
		})
	}
}

func TestGetPriceIDForPlan(t *testing.T) {
	cfg := &config.Config{
		StripePriceStarterID:    "price_starter_123",
		StripePriceGrowthID:     "price_growth_456",
		StripePriceEnterpriseID: "price_enterprise_789",
	}

	tests := []struct {
		name     string
		planKey  string
		expected string
	}{
		{
			name:     "starter plan",
			planKey:  "starter",
			expected: "price_starter_123",
		},
		{
			name:     "growth plan",
			planKey:  "growth",
			expected: "price_growth_456",
		},
		{
			name:     "enterprise plan",
			planKey:  "enterprise",
			expected: "price_enterprise_789",
		},
		{
			name:     "free plan",
			planKey:  "free",
			expected: "",
		},
		{
			name:     "unknown plan",
			planKey:  "unknown",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.GetPriceIDForPlan(tt.planKey)
			if result != tt.expected {
				t.Errorf("GetPriceIDForPlan(%q) = %q, want %q", tt.planKey, result, tt.expected)
			}
		})
	}
}

func TestBillingSubscriptionIsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   model.SubscriptionStatus
		expected bool
	}{
		{
			name:     "active subscription",
			status:   model.SubscriptionStatusActive,
			expected: true,
		},
		{
			name:     "trialing subscription",
			status:   model.SubscriptionStatusTrialing,
			expected: true,
		},
		{
			name:     "past_due subscription",
			status:   model.SubscriptionStatusPastDue,
			expected: false,
		},
		{
			name:     "canceled subscription",
			status:   model.SubscriptionStatusCanceled,
			expected: false,
		},
		{
			name:     "unpaid subscription",
			status:   model.SubscriptionStatusUnpaid,
			expected: false,
		},
		{
			name:     "incomplete subscription",
			status:   model.SubscriptionStatusIncomplete,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &model.BillingSubscription{
				Status: tt.status,
			}
			result := sub.IsActive()
			if result != tt.expected {
				t.Errorf("IsActive() with status %q = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestBillingSubscriptionIsPremium(t *testing.T) {
	tests := []struct {
		name     string
		status   model.SubscriptionStatus
		planKey  model.PlanKey
		expected bool
	}{
		{
			name:     "active starter",
			status:   model.SubscriptionStatusActive,
			planKey:  model.PlanKeyStarter,
			expected: true,
		},
		{
			name:     "active growth",
			status:   model.SubscriptionStatusActive,
			planKey:  model.PlanKeyGrowth,
			expected: true,
		},
		{
			name:     "active enterprise",
			status:   model.SubscriptionStatusActive,
			planKey:  model.PlanKeyEnterprise,
			expected: true,
		},
		{
			name:     "active free",
			status:   model.SubscriptionStatusActive,
			planKey:  model.PlanKeyFree,
			expected: false,
		},
		{
			name:     "canceled starter",
			status:   model.SubscriptionStatusCanceled,
			planKey:  model.PlanKeyStarter,
			expected: false,
		},
		{
			name:     "trialing growth",
			status:   model.SubscriptionStatusTrialing,
			planKey:  model.PlanKeyGrowth,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &model.BillingSubscription{
				Status:  tt.status,
				PlanKey: tt.planKey,
			}
			result := sub.IsPremium()
			if result != tt.expected {
				t.Errorf("IsPremium() with status=%q planKey=%q = %v, want %v", tt.status, tt.planKey, result, tt.expected)
			}
		})
	}
}

func TestIsBillingEnabled(t *testing.T) {
	tests := []struct {
		name          string
		secretKey     string
		webhookSecret string
		expected      bool
	}{
		{
			name:          "fully configured",
			secretKey:     "sk_test_123",
			webhookSecret: "whsec_123",
			expected:      true,
		},
		{
			name:          "missing secret key",
			secretKey:     "",
			webhookSecret: "whsec_123",
			expected:      false,
		},
		{
			name:          "missing webhook secret",
			secretKey:     "sk_test_123",
			webhookSecret: "",
			expected:      false,
		},
		{
			name:          "both missing",
			secretKey:     "",
			webhookSecret: "",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				StripeBillingSecretKey: tt.secretKey,
				StripeWebhookSecret:    tt.webhookSecret,
			}
			result := cfg.IsBillingEnabled()
			if result != tt.expected {
				t.Errorf("IsBillingEnabled() = %v, want %v", result, tt.expected)
			}
		})
	}
}



