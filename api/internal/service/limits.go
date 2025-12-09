package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"rev-saas-api/internal/model"
	mongorepo "rev-saas-api/internal/repository/mongo"
)

// LimitError codes for frontend to handle
const (
	LimitCodeCompetitors   = "LIMIT_COMPETITORS"
	LimitCodePlans         = "LIMIT_PLANS"
	LimitCodeAnalyses      = "LIMIT_ANALYSES"
	LimitCodeTrialExpired  = "LIMIT_TRIAL_EXPIRED"
)

// LimitCheckResult represents the result of a limit check.
type LimitCheckResult struct {
	Allowed   bool   `json:"allowed"`
	ErrorCode string `json:"error_code,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Plan      string `json:"plan,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Current   int    `json:"current,omitempty"`
}

// LimitsService handles plan-based limit checking.
type LimitsService struct {
	userRepo       *mongorepo.UserRepository
	planRepo       *mongorepo.PlanRepository
	competitorRepo *mongorepo.CompetitorRepository
	analysisRepo   *mongorepo.AnalysisRepository
}

// NewLimitsService creates a new LimitsService.
func NewLimitsService(
	userRepo *mongorepo.UserRepository,
	planRepo *mongorepo.PlanRepository,
	competitorRepo *mongorepo.CompetitorRepository,
	analysisRepo *mongorepo.AnalysisRepository,
) *LimitsService {
	return &LimitsService{
		userRepo:       userRepo,
		planRepo:       planRepo,
		competitorRepo: competitorRepo,
		analysisRepo:   analysisRepo,
	}
}

// CanAddCompetitor checks if the user can add another competitor.
func (s *LimitsService) CanAddCompetitor(ctx context.Context, user *model.User) LimitCheckResult {
	// Admin users have unlimited access
	if user.HasUnlimitedAccess() {
		return LimitCheckResult{Allowed: true}
	}

	plan := user.GetEffectivePlan()
	limits := GetPlanLimits(plan)

	if limits.IsUnlimited {
		return LimitCheckResult{Allowed: true}
	}

	// Count current competitors
	uid, _ := primitive.ObjectIDFromHex(user.ID.Hex())
	competitors, err := s.competitorRepo.ListByUser(ctx, uid)
	if err != nil {
		// On error, allow the operation (fail open for better UX)
		return LimitCheckResult{Allowed: true}
	}

	currentCount := len(competitors)

	if currentCount >= limits.MaxCompetitors {
		return LimitCheckResult{
			Allowed:   false,
			ErrorCode: LimitCodeCompetitors,
			Reason:    "Maximum competitors limit reached for your plan",
			Plan:      plan,
			Limit:     limits.MaxCompetitors,
			Current:   currentCount,
		}
	}

	return LimitCheckResult{Allowed: true}
}

// CanAddPlan checks if the user can add another pricing plan.
func (s *LimitsService) CanAddPlan(ctx context.Context, user *model.User) LimitCheckResult {
	// Admin users have unlimited access
	if user.HasUnlimitedAccess() {
		return LimitCheckResult{Allowed: true}
	}

	plan := user.GetEffectivePlan()
	limits := GetPlanLimits(plan)

	if limits.IsUnlimited {
		return LimitCheckResult{Allowed: true}
	}

	// Count current plans
	uid, _ := primitive.ObjectIDFromHex(user.ID.Hex())
	plans, err := s.planRepo.ListByUser(ctx, uid)
	if err != nil {
		return LimitCheckResult{Allowed: true}
	}

	currentCount := len(plans)

	if currentCount >= limits.MaxPlans {
		return LimitCheckResult{
			Allowed:   false,
			ErrorCode: LimitCodePlans,
			Reason:    "Maximum pricing plans limit reached for your plan",
			Plan:      plan,
			Limit:     limits.MaxPlans,
			Current:   currentCount,
		}
	}

	return LimitCheckResult{Allowed: true}
}

// CanRunAnalysis checks if the user can run another analysis.
func (s *LimitsService) CanRunAnalysis(ctx context.Context, user *model.User) LimitCheckResult {
	// Admin users have unlimited access
	if user.HasUnlimitedAccess() {
		return LimitCheckResult{Allowed: true}
	}

	plan := user.GetEffectivePlan()
	limits := GetPlanLimits(plan)

	if limits.IsUnlimited {
		return LimitCheckResult{Allowed: true}
	}

	// For free plan, check trial expiry and lifetime limit
	if plan == model.PlanFree {
		// Check trial expiry
		if user.IsTrialExpired() {
			return LimitCheckResult{
				Allowed:   false,
				ErrorCode: LimitCodeTrialExpired,
				Reason:    "Your free trial has expired. Please upgrade to continue.",
				Plan:      plan,
			}
		}

		// Check lifetime analysis limit
		if limits.MaxAnalysesTotal > 0 && user.AnalysisTotalUsed >= limits.MaxAnalysesTotal {
			return LimitCheckResult{
				Allowed:   false,
				ErrorCode: LimitCodeAnalyses,
				Reason:    "You have used all your free analyses. Please upgrade to continue.",
				Plan:      plan,
				Limit:     limits.MaxAnalysesTotal,
				Current:   user.AnalysisTotalUsed,
			}
		}

		return LimitCheckResult{Allowed: true}
	}

	// For paid plans, check monthly limit
	if limits.MaxAnalysesPerMonth > 0 {
		monthCount := s.getMonthlyAnalysisCount(user)

		if monthCount >= limits.MaxAnalysesPerMonth {
			return LimitCheckResult{
				Allowed:   false,
				ErrorCode: LimitCodeAnalyses,
				Reason:    "You have reached your monthly analysis limit. Please upgrade or wait until next month.",
				Plan:      plan,
				Limit:     limits.MaxAnalysesPerMonth,
				Current:   monthCount,
			}
		}
	}

	return LimitCheckResult{Allowed: true}
}

// getMonthlyAnalysisCount returns the current month's analysis count, resetting if needed.
func (s *LimitsService) getMonthlyAnalysisCount(user *model.User) int {
	now := time.Now()

	// Check if we need to reset the monthly counter
	if user.AnalysisMonthStart.IsZero() || !isSameMonth(user.AnalysisMonthStart, now) {
		// Month has changed or never set - count should be reset
		// This will be done when IncrementAnalysisCount is called
		return 0
	}

	return user.AnalysisMonthCount
}

// IncrementAnalysisCount increments the analysis counter for the user.
// Call this AFTER a successful analysis run.
func (s *LimitsService) IncrementAnalysisCount(ctx context.Context, user *model.User) error {
	plan := user.GetEffectivePlan()
	now := time.Now()

	// For free plan, increment lifetime counter
	if plan == model.PlanFree {
		user.AnalysisTotalUsed++
	} else {
		// For paid plans, handle monthly counter
		if user.AnalysisMonthStart.IsZero() || !isSameMonth(user.AnalysisMonthStart, now) {
			// Reset monthly counter
			user.AnalysisMonthStart = now
			user.AnalysisMonthCount = 1
		} else {
			user.AnalysisMonthCount++
		}
	}

	// Update user in database
	return s.userRepo.Update(ctx, user)
}

// GetUserUsageStats returns usage statistics for a user.
func (s *LimitsService) GetUserUsageStats(ctx context.Context, user *model.User) map[string]interface{} {
	plan := user.GetEffectivePlan()
	limits := GetPlanLimits(plan)

	uid, _ := primitive.ObjectIDFromHex(user.ID.Hex())

	// Count current resources
	competitors, _ := s.competitorRepo.ListByUser(ctx, uid)
	plans, _ := s.planRepo.ListByUser(ctx, uid)

	competitorCount := 0
	if competitors != nil {
		competitorCount = len(competitors)
	}

	planCount := 0
	if plans != nil {
		planCount = len(plans)
	}

	stats := map[string]interface{}{
		"plan":              plan,
		"competitors_used":  competitorCount,
		"competitors_limit": limits.MaxCompetitors,
		"plans_used":        planCount,
		"plans_limit":       limits.MaxPlans,
		"is_unlimited":      limits.IsUnlimited || user.HasUnlimitedAccess(),
	}

	if plan == model.PlanFree {
		stats["analyses_used"] = user.AnalysisTotalUsed
		stats["analyses_limit"] = limits.MaxAnalysesTotal
		stats["trial_expires_at"] = user.TrialExpiresAt
		stats["trial_expired"] = user.IsTrialExpired()
	} else if !limits.IsUnlimited {
		monthCount := s.getMonthlyAnalysisCount(user)
		stats["analyses_used_this_month"] = monthCount
		stats["analyses_limit_per_month"] = limits.MaxAnalysesPerMonth
	}

	return stats
}

// isSameMonth checks if two times are in the same month and year.
func isSameMonth(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month()
}



