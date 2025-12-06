package service

import "rev-saas-api/internal/model"

// PlanLimits defines the limits for a subscription plan.
type PlanLimits struct {
	MaxCompetitors       int  // Maximum number of competitors allowed
	MaxPlans             int  // Maximum number of pricing plans allowed
	MaxAnalysesPerMonth  int  // Maximum analyses per month (0 = unlimited)
	MaxAnalysesTotal     int  // Maximum total analyses lifetime (0 = unlimited, used for free plan)
	TrialDays            int  // Trial period in days (0 = no trial)
	IsUnlimited          bool // If true, skip all checks
}

// Plan limits configuration (hardcoded)
var planLimitsConfig = map[string]PlanLimits{
	model.PlanFree: {
		MaxCompetitors:      1,
		MaxPlans:            3,
		MaxAnalysesPerMonth: 0, // Not used for free plan
		MaxAnalysesTotal:    2, // Lifetime limit
		TrialDays:           14,
		IsUnlimited:         false,
	},
	model.PlanStarter: {
		MaxCompetitors:      3,
		MaxPlans:            3,
		MaxAnalysesPerMonth: 5,
		MaxAnalysesTotal:    0, // Unlimited lifetime
		TrialDays:           0, // No trial for paid plans
		IsUnlimited:         false,
	},
	model.PlanGrowth: {
		MaxCompetitors:      5,
		MaxPlans:            5,
		MaxAnalysesPerMonth: 10,
		MaxAnalysesTotal:    0,
		TrialDays:           0,
		IsUnlimited:         false,
	},
	model.PlanEnterprise: {
		MaxCompetitors:      10,
		MaxPlans:            7,
		MaxAnalysesPerMonth: 20,
		MaxAnalysesTotal:    0,
		TrialDays:           0,
		IsUnlimited:         false,
	},
	model.PlanAdmin: {
		MaxCompetitors:      0,
		MaxPlans:            0,
		MaxAnalysesPerMonth: 0,
		MaxAnalysesTotal:    0,
		TrialDays:           0,
		IsUnlimited:         true,
	},
}

// GetPlanLimits returns the limits for a given plan.
// If the plan is not found, returns free plan limits.
func GetPlanLimits(plan string) PlanLimits {
	if plan == "" {
		plan = model.PlanFree
	}

	limits, exists := planLimitsConfig[plan]
	if !exists {
		return planLimitsConfig[model.PlanFree]
	}

	return limits
}

// GetAllPlanLimits returns all plan limits (useful for API responses).
func GetAllPlanLimits() map[string]PlanLimits {
	return planLimitsConfig
}

