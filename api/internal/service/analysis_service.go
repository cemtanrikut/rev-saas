package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"rev-saas-api/internal/model"
	mongorepo "rev-saas-api/internal/repository/mongo"
)

// BusinessContext holds derived business metrics for analysis.
type BusinessContext struct {
	HasMetrics   bool
	MRR          float64
	Customers    int
	ChurnRate    float64
	ARPU         float64 // Average Revenue Per User
	IsEarlyStage bool    // Low MRR and few customers
	IsHighChurn  bool    // Churn > 7%
	IsLowChurn   bool    // Churn < 3%
	IsHealthy    bool    // Decent MRR and low churn
}

// AnalysisService handles business logic for pricing analysis.
type AnalysisService struct {
	analysisRepo        *mongorepo.AnalysisRepository
	planRepo            *mongorepo.PlanRepository
	competitorRepo      *mongorepo.CompetitorRepository
	businessMetricsRepo *mongorepo.BusinessMetricsRepository
}

// NewAnalysisService creates a new AnalysisService.
func NewAnalysisService(
	analysisRepo *mongorepo.AnalysisRepository,
	planRepo *mongorepo.PlanRepository,
	competitorRepo *mongorepo.CompetitorRepository,
	businessMetricsRepo *mongorepo.BusinessMetricsRepository,
) *AnalysisService {
	return &AnalysisService{
		analysisRepo:        analysisRepo,
		planRepo:            planRepo,
		competitorRepo:      competitorRepo,
		businessMetricsRepo: businessMetricsRepo,
	}
}

// deriveBusinessContext creates a BusinessContext from raw metrics.
func deriveBusinessContext(metrics *model.BusinessMetrics) BusinessContext {
	if metrics == nil {
		return BusinessContext{HasMetrics: false}
	}

	ctx := BusinessContext{
		HasMetrics: true,
		MRR:        metrics.MRR,
		Customers:  metrics.Customers,
		ChurnRate:  metrics.MonthlyChurnRate,
	}

	// Calculate ARPU
	if ctx.Customers > 0 {
		ctx.ARPU = ctx.MRR / float64(ctx.Customers)
	}

	// Determine churn level
	ctx.IsHighChurn = ctx.ChurnRate > 7.0
	ctx.IsLowChurn = ctx.ChurnRate < 3.0

	// Determine if early stage (MRR < $5k and < 50 customers)
	ctx.IsEarlyStage = ctx.MRR < 5000 && ctx.Customers < 50

	// Determine if healthy (MRR > $10k and low churn)
	ctx.IsHealthy = ctx.MRR >= 10000 && ctx.IsLowChurn

	return ctx
}

// RunAnalysis performs a rule-based pricing analysis for the given user.
// This is a convenience wrapper around RunAnalysisWithInput.
func (s *AnalysisService) RunAnalysis(ctx context.Context, userID string) (*model.Analysis, error) {
	analysis, _, err := s.RunAnalysisWithInput(ctx, userID)
	return analysis, err
}

// RunAnalysisWithInput performs a rule-based pricing analysis and returns the input data for AI.
func (s *AnalysisService) RunAnalysisWithInput(ctx context.Context, userID string) (*model.Analysis, *AIPricingInput, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, nil, errors.New("invalid user id")
	}

	// Fetch user's plans
	plans, err := s.planRepo.ListByUser(ctx, uid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch plans: %w", err)
	}

	// Fetch user's competitors
	competitors, err := s.competitorRepo.ListByUser(ctx, uid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch competitors: %w", err)
	}

	// Fetch business metrics (optional)
	metrics, err := s.businessMetricsRepo.GetByUserID(ctx, uid)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch business metrics: %w", err)
	}

	// Build AI input
	aiInput := &AIPricingInput{
		Plans:           plans,
		Competitors:     competitors,
		BusinessMetrics: metrics,
	}

	// Derive business context
	bizCtx := deriveBusinessContext(metrics)

	numPlans := len(plans)
	numCompetitors := len(competitors)

	// Build the analysis
	analysis := &model.Analysis{
		UserID:          uid,
		NumPlans:        numPlans,
		NumCompetitors:  numCompetitors,
		Recommendations: []model.Recommendation{},
	}

	// Handle edge cases - avoid mentioning counts or specifics
	if numPlans == 0 && numCompetitors == 0 {
		analysis.Summary = "No data available. Please add your pricing plans and competitor information to receive recommendations."
		if err := s.analysisRepo.Create(ctx, analysis); err != nil {
			return nil, nil, fmt.Errorf("failed to save analysis: %w", err)
		}
		return analysis, aiInput, nil
	}

	if numPlans == 0 {
		analysis.Summary = "No pricing plans found. Please define your pricing plans to receive recommendations."
		if err := s.analysisRepo.Create(ctx, analysis); err != nil {
			return nil, nil, fmt.Errorf("failed to save analysis: %w", err)
		}
		return analysis, aiInput, nil
	}

	if numCompetitors == 0 {
		analysis.Summary = "No competitor data available. Add competitor pricing information to receive positioning recommendations."
		// Still generate basic recommendations without competitor data
		for _, plan := range plans {
			rec := model.Recommendation{
				PlanID:            plan.ID,
				PlanName:          plan.Name,
				CurrentPrice:      plan.Price,
				Position:          "unknown",
				SuggestedAction:   "add_competitors",
				SuggestedNewPrice: plan.Price,
				Rationale:         "Add competitor data to receive pricing recommendations.",
			}
			analysis.Recommendations = append(analysis.Recommendations, rec)
		}
		if err := s.analysisRepo.Create(ctx, analysis); err != nil {
			return nil, nil, fmt.Errorf("failed to save analysis: %w", err)
		}
		return analysis, aiInput, nil
	}

	// Extract competitor prices and compute statistics
	// Now competitors have multiple plans, extract all prices
	competitorPrices := make([]float64, 0)
	for _, c := range competitors {
		for _, p := range c.Plans {
			if p.Price > 0 {
				competitorPrices = append(competitorPrices, p.Price)
			}
		}
	}

	// If no competitor prices available, use a default approach
	if len(competitorPrices) == 0 {
		analysis.Summary = "Competitor entries exist but no base prices are set. Add base prices to competitor entries for positioning analysis."
		for _, plan := range plans {
			rec := model.Recommendation{
				PlanID:            plan.ID,
				PlanName:          plan.Name,
				CurrentPrice:      plan.Price,
				Position:          "unknown",
				SuggestedAction:   "keep",
				SuggestedNewPrice: plan.Price,
				Rationale:         "No competitor price data available for comparison. Add base prices to competitors.",
			}
			analysis.Recommendations = append(analysis.Recommendations, rec)
		}
		if err := s.analysisRepo.Create(ctx, analysis); err != nil {
			return nil, nil, fmt.Errorf("failed to save analysis: %w", err)
		}
		return analysis, aiInput, nil
	}

	// Compute statistics
	sort.Float64s(competitorPrices)
	minPrice := competitorPrices[0]
	maxPrice := competitorPrices[len(competitorPrices)-1]
	medianPrice := computeMedian(competitorPrices)

	// Generate recommendations for each plan
	var belowCount, aroundCount, aboveCount int

	for _, plan := range plans {
		rec := s.generateRecommendation(plan, minPrice, maxPrice, medianPrice, bizCtx)
		analysis.Recommendations = append(analysis.Recommendations, rec)

		switch rec.Position {
		case "below_median", "lowest":
			belowCount++
		case "around_median":
			aroundCount++
		case "above_median", "highest":
			aboveCount++
		}
	}

	// Build professional summary with full plan analysis
	analysis.Summary = s.buildSummaryWithPlans(plans, bizCtx)

	// Save the analysis
	if err := s.analysisRepo.Create(ctx, analysis); err != nil {
		return nil, nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	return analysis, aiInput, nil
}

// UpdateAnalysis updates an existing analysis in the database.
func (s *AnalysisService) UpdateAnalysis(ctx context.Context, analysis *model.Analysis) error {
	return s.analysisRepo.Update(ctx, analysis)
}

// generateRecommendation creates a recommendation for a single plan based on competitor data and business context.
// NOTE: Rationale text avoids mentioning exact median values - uses directional language only.
func (s *AnalysisService) generateRecommendation(plan *model.Plan, minPrice, maxPrice, medianPrice float64, bizCtx BusinessContext) model.Recommendation {
	rec := model.Recommendation{
		PlanID:       plan.ID,
		PlanName:     plan.Name,
		CurrentPrice: plan.Price,
	}

	price := plan.Price

	// Determine base position (internal use only, not exposed in rationale)
	var position string
	var baseAction string
	var baseSuggestedPrice float64
	var baseRationale string

	switch {
	case price < medianPrice*0.5:
		position = "lowest"
		baseAction = "raise_price"
		baseSuggestedPrice = medianPrice * 0.8
		baseRationale = "Your price appears significantly lower than competitors."

	case price < medianPrice*0.7:
		position = "below_median"
		baseAction = "raise_price"
		baseSuggestedPrice = medianPrice * 0.9
		baseRationale = "Your price appears lower than most competitors."

	case price >= medianPrice*0.7 && price <= medianPrice*1.3:
		position = "around_median"
		baseAction = "keep"
		baseSuggestedPrice = price
		baseRationale = "Your price appears competitively positioned."

	case price > medianPrice*1.5:
		position = "highest"
		baseAction = "lower_price"
		baseSuggestedPrice = medianPrice * 1.2
		baseRationale = "Your price appears significantly higher than competitors."

	case price > medianPrice*1.3:
		position = "above_median"
		baseAction = "lower_price"
		baseSuggestedPrice = medianPrice * 1.1
		baseRationale = "Your price appears higher than most competitors."

	default:
		position = "around_median"
		baseAction = "keep"
		baseSuggestedPrice = price
		baseRationale = "Your pricing appears competitive."
	}

	rec.Position = position

	// Apply business context adjustments
	if bizCtx.HasMetrics {
		rec.SuggestedAction, rec.SuggestedNewPrice, rec.Rationale = s.applyBusinessContextAdjustments(
			position, baseAction, baseSuggestedPrice, baseRationale, price, medianPrice, bizCtx,
		)
	} else {
		// No business metrics - use base recommendations
		rec.SuggestedAction = baseAction
		rec.SuggestedNewPrice = roundToTwoDecimals(baseSuggestedPrice)
		rec.Rationale = baseRationale + " Consider adding your business metrics for more tailored recommendations."
	}

	return rec
}

// applyBusinessContextAdjustments modifies recommendations based on business metrics.
// NOTE: Rationale avoids exact median values - uses directional language only.
func (s *AnalysisService) applyBusinessContextAdjustments(
	position, baseAction string,
	baseSuggestedPrice float64,
	baseRationale string,
	currentPrice, medianPrice float64,
	bizCtx BusinessContext,
) (string, float64, string) {
	action := baseAction
	suggestedPrice := baseSuggestedPrice
	rationale := baseRationale

	switch {
	// HIGH CHURN (>7%): Be conservative with price increases
	case bizCtx.IsHighChurn:
		if baseAction == "raise_price" {
			// Reduce the aggressiveness of price increases
			if position == "lowest" {
				suggestedPrice = currentPrice * 1.1
				action = "raise_price_conservative"
				rationale = fmt.Sprintf("%s However, your churn rate (%.1f%%) is elevated. Consider a conservative, small increase. Focus on retention before aggressive price changes.", baseRationale, bizCtx.ChurnRate)
			} else if position == "below_median" {
				suggestedPrice = currentPrice * 1.05
				action = "raise_price_conservative"
				rationale = fmt.Sprintf("%s Given your elevated churn rate (%.1f%%), consider a modest increase. Address retention issues before larger pricing adjustments.", baseRationale, bizCtx.ChurnRate)
			}
		} else if baseAction == "keep" {
			rationale = fmt.Sprintf("%s Note: Your churn rate (%.1f%%) is concerning. Prioritize improving retention over pricing changes.", baseRationale, bizCtx.ChurnRate)
		} else if baseAction == "lower_price" {
			rationale = fmt.Sprintf("%s With your elevated churn rate (%.1f%%), slightly reducing prices could help improve retention.", baseRationale, bizCtx.ChurnRate)
		}

	// LOW CHURN (<3%) + HEALTHY BUSINESS: Allow more aggressive increases
	case bizCtx.IsHealthy:
		if baseAction == "raise_price" {
			if position == "lowest" {
				suggestedPrice = medianPrice * 0.95
				action = "raise_price_aggressive"
				rationale = fmt.Sprintf("%s Your business is healthy (churn: %.1f%%). With strong retention, you can confidently raise prices.", baseRationale, bizCtx.ChurnRate)
			} else if position == "below_median" {
				suggestedPrice = medianPrice
				action = "raise_price_aggressive"
				rationale = fmt.Sprintf("%s With excellent retention (%.1f%% churn), consider a moderate price increase.", baseRationale, bizCtx.ChurnRate)
			}
		} else if baseAction == "keep" && position == "around_median" {
			suggestedPrice = currentPrice * 1.05
			action = "consider_increase"
			rationale = fmt.Sprintf("%s Your low churn (%.1f%%) suggests customers find strong value. You could experiment with a slight increase.", baseRationale, bizCtx.ChurnRate)
		}

	// EARLY STAGE: Prioritize adoption over revenue
	case bizCtx.IsEarlyStage:
		if baseAction == "raise_price" {
			action = "keep_for_growth"
			suggestedPrice = currentPrice
			rationale = fmt.Sprintf("%s However, at your early stage, we recommend keeping prices competitive to maximize customer acquisition.", baseRationale)
		} else if baseAction == "lower_price" && position == "highest" {
			suggestedPrice = medianPrice
			action = "lower_price"
			rationale = fmt.Sprintf("%s At your growth stage, premium pricing may slow acquisition. Consider reducing prices to accelerate growth.", baseRationale)
		} else if baseAction == "keep" {
			rationale = fmt.Sprintf("%s At your early stage, competitive pricing supports customer acquisition.", baseRationale)
		}

	// LOW CHURN but not healthy yet (growing)
	case bizCtx.IsLowChurn && !bizCtx.IsHealthy:
		if baseAction == "raise_price" {
			rationale = fmt.Sprintf("%s Your excellent retention (%.1f%% churn) indicates strong product-market fit. A price increase is well-supported.", baseRationale, bizCtx.ChurnRate)
		}

	// Default: no additional context
	default:
		// Keep base rationale as-is
	}

	return action, roundToTwoDecimals(suggestedPrice), rationale
}

// ═══════════════════════════════════════════════════════════════════════════════
// PROFESSIONAL RULE-BASED ANALYSIS ENGINE
// Provides smart, deterministic pricing insights without AI
// ═══════════════════════════════════════════════════════════════════════════════

// PricingLadderAnalysis holds insights about the pricing tier structure
type PricingLadderAnalysis struct {
	HasMultipleTiers    bool
	TierRatios          []TierRatio
	HasWeakUpgradePath  bool // ratio < 1.2
	HasSteepCliff       bool // ratio > 3.0
	SpreadAnalysis      string
	IsCompressed        bool // max <= 1.5x min
	IsPolarized         bool // max >= 5x min
}

// TierRatio represents the price ratio between adjacent tiers
type TierRatio struct {
	LowerTier  string
	HigherTier string
	Ratio      float64
	Assessment string
}

// PsychologicalAnalysis holds insights about psychological pricing
type PsychologicalAnalysis struct {
	Insights []PsychologicalInsight
}

// PsychologicalInsight represents a single psychological pricing observation
type PsychologicalInsight struct {
	PlanName   string
	Price      float64
	Observation string
}

// Common psychological price thresholds
var psychologicalThresholds = []float64{9, 19, 29, 49, 79, 99, 149, 199, 299, 499}

// getRatioDescription converts a numeric ratio to directional language
func getRatioDescription(ratio float64) string {
	switch {
	case ratio < 1.2:
		return "a very narrow gap"
	case ratio < 1.5:
		return "a modest step"
	case ratio >= 1.5 && ratio <= 2.0:
		return "a solid value step"
	case ratio > 2.0 && ratio <= 2.5:
		return "a strong value step"
	case ratio > 2.5 && ratio <= 3.0:
		return "a significant jump"
	case ratio > 3.0:
		return "a steep jump"
	default:
		return "a reasonable step"
	}
}

// getTierOrder returns the expected order rank for common tier names
// Lower number = entry tier, higher number = premium tier
func getTierOrder(name string) int {
	nameLower := strings.ToLower(name)
	
	// Entry-level tiers
	if strings.Contains(nameLower, "free") || strings.Contains(nameLower, "basic") {
		return 1
	}
	if strings.Contains(nameLower, "starter") || strings.Contains(nameLower, "lite") || strings.Contains(nameLower, "personal") {
		return 2
	}
	
	// Mid-tier
	if strings.Contains(nameLower, "pro") || strings.Contains(nameLower, "professional") || strings.Contains(nameLower, "standard") {
		return 3
	}
	if strings.Contains(nameLower, "plus") || strings.Contains(nameLower, "growth") || strings.Contains(nameLower, "team") {
		return 4
	}
	
	// Premium tiers
	if strings.Contains(nameLower, "business") || strings.Contains(nameLower, "premium") {
		return 5
	}
	if strings.Contains(nameLower, "enterprise") || strings.Contains(nameLower, "unlimited") || strings.Contains(nameLower, "scale") {
		return 6
	}
	
	return 3 // Default to mid-tier if unknown
}

// detectNamingInconsistency checks if tier names don't match price expectations
func detectNamingInconsistency(lowerPricePlan, higherPricePlan *model.Plan) string {
	lowerTierExpectedOrder := getTierOrder(lowerPricePlan.Name)
	higherTierExpectedOrder := getTierOrder(higherPricePlan.Name)
	
	// If the cheaper plan has a "higher" tier name than the expensive plan, that's inconsistent
	if lowerTierExpectedOrder > higherTierExpectedOrder {
		return fmt.Sprintf("Observation: %s is priced lower than %s, but the naming convention suggests the opposite hierarchy - aligning tier names with their price positioning can reduce customer confusion", lowerPricePlan.Name, higherPricePlan.Name)
	}
	
	return ""
}

// analyzePricingLadder analyzes the value ladder and tier ratios
func analyzePricingLadder(plans []*model.Plan) PricingLadderAnalysis {
	analysis := PricingLadderAnalysis{}
	
	if len(plans) < 2 {
		analysis.HasMultipleTiers = false
		return analysis
	}
	
	analysis.HasMultipleTiers = true
	
	// Sort plans by price (ascending: entry → mid → premium)
	sortedPlans := make([]*model.Plan, len(plans))
	copy(sortedPlans, plans)
	sort.Slice(sortedPlans, func(i, j int) bool {
		return sortedPlans[i].Price < sortedPlans[j].Price
	})
	
	// Filter out free plans for ratio analysis
	paidPlans := make([]*model.Plan, 0)
	for _, p := range sortedPlans {
		if p.Price > 0 {
			paidPlans = append(paidPlans, p)
		}
	}
	
	if len(paidPlans) < 2 {
		return analysis
	}
	
	// Calculate ratios between adjacent paid tiers (lower → higher)
	for i := 0; i < len(paidPlans)-1; i++ {
		entryTier := paidPlans[i]
		nextTier := paidPlans[i+1]
		
		if entryTier.Price > 0 {
			ratio := nextTier.Price / entryTier.Price
			
			tierRatio := TierRatio{
				LowerTier:  entryTier.Name,  // Entry/lower tier
				HigherTier: nextTier.Name,   // Next/higher tier
				Ratio:      ratio,
			}
			
			// Assess the ratio
			if ratio < 1.2 {
				tierRatio.Assessment = "too_close"
				analysis.HasWeakUpgradePath = true
			} else if ratio > 3.0 {
				tierRatio.Assessment = "steep_cliff"
				analysis.HasSteepCliff = true
			} else if ratio >= 1.5 && ratio <= 2.5 {
				tierRatio.Assessment = "healthy"
			} else {
				tierRatio.Assessment = "acceptable"
			}
			
			analysis.TierRatios = append(analysis.TierRatios, tierRatio)
		}
	}
	
	// Analyze overall spread
	minPaid := paidPlans[0].Price
	maxPaid := paidPlans[len(paidPlans)-1].Price
	
	if minPaid > 0 {
		spreadRatio := maxPaid / minPaid
		
		if spreadRatio <= 1.5 {
			analysis.IsCompressed = true
			analysis.SpreadAnalysis = "compressed"
		} else if spreadRatio >= 5.0 {
			analysis.IsPolarized = true
			analysis.SpreadAnalysis = "polarized"
		} else if spreadRatio >= 2.0 && spreadRatio <= 4.0 {
			analysis.SpreadAnalysis = "healthy"
		} else {
			analysis.SpreadAnalysis = "acceptable"
		}
	}
	
	return analysis
}

// analyzePsychologicalPricing analyzes psychological pricing thresholds
// Focuses on perceived value and messaging, not tactical price changes
func analyzePsychologicalPricing(plans []*model.Plan) PsychologicalAnalysis {
	analysis := PsychologicalAnalysis{
		Insights: make([]PsychologicalInsight, 0),
	}
	
	for _, plan := range plans {
		if plan.Price <= 0 {
			continue
		}
		
		insight := PsychologicalInsight{
			PlanName: plan.Name,
			Price:    plan.Price,
		}
		
		// Check against psychological thresholds
		for _, threshold := range psychologicalThresholds {
			// Just below threshold (good positioning) - e.g., $29, $49, $99
			if plan.Price >= threshold-2 && plan.Price <= threshold {
				insight.Observation = fmt.Sprintf("is positioned effectively at $%.0f, leveraging a key psychological price point that typically resonates well with buyers", plan.Price)
				break
			}
			
			// Sitting just above threshold - focus on value clarity, not price reduction
			if plan.Price > threshold && plan.Price <= threshold+5 {
				insight.Observation = fmt.Sprintf("at $%.0f sits slightly above the $%.0f threshold - ensure the value proposition clearly justifies the premium to maintain strong conversion", plan.Price, threshold)
				break
			}
		}
		
		// Check for pricing patterns if no threshold insight
		if insight.Observation == "" {
			priceInt := int(plan.Price)
			lastDigit := priceInt % 10
			
			if lastDigit == 9 {
				insight.Observation = fmt.Sprintf("at $%.0f employs effective charm pricing", plan.Price)
			} else if lastDigit == 0 && plan.Price >= 50 {
				insight.Observation = fmt.Sprintf("at $%.0f uses round-number pricing, which can signal premium positioning but may benefit from value-focused messaging", plan.Price)
			} else if lastDigit == 5 {
				insight.Observation = fmt.Sprintf("at $%.0f sits at a mid-point, which can work well with clear tier differentiation", plan.Price)
			}
		}
		
		if insight.Observation != "" {
			analysis.Insights = append(analysis.Insights, insight)
		}
	}
	
	return analysis
}

// buildProfessionalSummary creates a rich, consultant-quality summary
// Written in executive memo style with directional language (no raw numbers)
// Focuses on value perception, messaging, and upgrade motivation
func (s *AnalysisService) buildProfessionalSummary(plans []*model.Plan, bizCtx BusinessContext) string {
	var paragraphs []string
	
	// Sort plans by price for consistent upgrade path description
	sortedPlans := make([]*model.Plan, len(plans))
	copy(sortedPlans, plans)
	sort.Slice(sortedPlans, func(i, j int) bool {
		return sortedPlans[i].Price < sortedPlans[j].Price
	})
	
	// ═══════════════════════════════════════════════════════════════
	// PARAGRAPH 1: Overall Pricing Architecture Assessment
	// ═══════════════════════════════════════════════════════════════
	ladderAnalysis := analyzePricingLadder(plans)
	
	var para1 string
	if len(plans) == 1 {
		para1 = "Your current pricing model features a single tier. "
		para1 += "While this simplicity can be effective for focused products, introducing additional tiers could help capture different customer segments and unlock additional revenue potential."
	} else {
		para1 = fmt.Sprintf("Your pricing architecture includes %d tiers, ", len(plans))
		
		// Spread analysis with consultant language
		if ladderAnalysis.IsCompressed {
			para1 += "however the overall price range is relatively narrow. "
			para1 += "When tiers are positioned too closely together, customers may not perceive meaningful differentiation, which can reduce upgrade motivation and limit revenue capture."
		} else if ladderAnalysis.IsPolarized {
			para1 += "spanning a wide range from entry-level to premium. "
			para1 += "While this approach can effectively serve distinct customer segments, consider whether mid-market buyers have a clear fit - the gap between options may result in lost conversions."
		} else if ladderAnalysis.SpreadAnalysis == "healthy" {
			para1 += "with a well-balanced spread across price points. "
			para1 += "This structure provides customers with clear upgrade paths and positions you to capture value across different willingness-to-pay segments."
		} else {
			para1 += "offering reasonable differentiation across price points."
		}
	}
	paragraphs = append(paragraphs, para1)
	
	// ═══════════════════════════════════════════════════════════════
	// PARAGRAPH 2: Upgrade Path, Value Perception & Tier Clarity
	// ═══════════════════════════════════════════════════════════════
	var para2Parts []string
	
	// Get paid plans for upgrade path analysis
	paidPlans := make([]*model.Plan, 0)
	for _, p := range sortedPlans {
		if p.Price > 0 {
			paidPlans = append(paidPlans, p)
		}
	}
	
	// Tier ratio insights - describe upgrades using canonical tier order (Starter → Pro → Enterprise)
	if len(ladderAnalysis.TierRatios) > 0 {
		for i, tr := range ladderAnalysis.TierRatios {
			ratioDesc := getRatioDescription(tr.Ratio)
			
			// Get the actual plan objects for naming check
			var lowerPricePlan, higherPricePlan *model.Plan
			if i < len(paidPlans)-1 {
				lowerPricePlan = paidPlans[i]
				higherPricePlan = paidPlans[i+1]
			}
			
			// Determine correct upgrade direction based on tier naming convention
			// Use canonical order: Starter → Pro → Enterprise (entry → premium)
			fromTier := tr.LowerTier
			toTier := tr.HigherTier
			
			// If tier naming suggests opposite order, swap for correct phrasing
			if lowerPricePlan != nil && higherPricePlan != nil {
				lowerTierOrder := getTierOrder(lowerPricePlan.Name)
				higherTierOrder := getTierOrder(higherPricePlan.Name)
				
				// If the cheaper plan has a "higher" tier name, the naming is inverted
				// Describe upgrade in canonical direction anyway
				if lowerTierOrder > higherTierOrder {
					fromTier = tr.HigherTier  // Use the "entry-level" named tier
					toTier = tr.LowerTier     // Use the "premium" named tier
				}
			}
			
			switch tr.Assessment {
			case "too_close":
				para2Parts = append(para2Parts, fmt.Sprintf("The upgrade from %s to %s represents %s, which may not provide sufficient perceived value to motivate customers to move to the higher tier - consider strengthening the feature differentiation between these plans", fromTier, toTier, ratioDesc))
			case "steep_cliff":
				para2Parts = append(para2Parts, fmt.Sprintf("The transition from %s to %s involves %s that some customers may find challenging to justify - evaluate whether clearer value communication or an intermediate tier could help capture buyers in this gap", fromTier, toTier, ratioDesc))
			case "healthy":
				para2Parts = append(para2Parts, fmt.Sprintf("The upgrade path from %s to %s offers %s, providing a compelling value progression that customers can clearly understand and justify", fromTier, toTier, ratioDesc))
			case "acceptable":
				para2Parts = append(para2Parts, fmt.Sprintf("The transition from %s to %s provides %s for customers evaluating an upgrade", fromTier, toTier, ratioDesc))
			}
			
			// Check for naming inconsistency (price vs naming mismatch)
			if lowerPricePlan != nil && higherPricePlan != nil {
				inconsistency := detectNamingInconsistency(lowerPricePlan, higherPricePlan)
				if inconsistency != "" {
					para2Parts = append(para2Parts, inconsistency)
				}
			}
		}
	}
	
	// Psychological pricing insights - focus on value perception, not price changes
	psychAnalysis := analyzePsychologicalPricing(plans)
	if len(psychAnalysis.Insights) > 0 {
		for _, insight := range psychAnalysis.Insights {
			if insight.Observation != "" {
				para2Parts = append(para2Parts, insight.PlanName+" "+insight.Observation)
			}
		}
	}
	
	if len(para2Parts) > 0 {
		para2 := strings.Join(para2Parts, ". ") + "."
		paragraphs = append(paragraphs, para2)
	}
	
	// ═══════════════════════════════════════════════════════════════
	// PARAGRAPH 3: Retention & Business Context
	// ═══════════════════════════════════════════════════════════════
	var para3 string
	
	if bizCtx.HasMetrics {
		if bizCtx.ChurnRate > 8 {
			para3 = "Your current retention metrics indicate an opportunity to strengthen customer relationships. "
			para3 += "Before exploring pricing adjustments, we recommend focusing on understanding what drives customer decisions to leave. "
			para3 += "In environments with elevated churn, changes to pricing can have amplified effects on retention. Prioritize delivering additional value and reinforcing your core value proposition."
		} else if bizCtx.ChurnRate > 5 {
			para3 = "Your retention metrics are at a moderate level, suggesting a stable but improvable customer base. "
			para3 += "There may be opportunity for thoughtful pricing evolution, but changes should be approached deliberately. "
			para3 += "Focus on ensuring each tier clearly communicates its value, and monitor customer response to any adjustments."
		} else if bizCtx.ChurnRate <= 5 && bizCtx.ChurnRate > 0 {
			para3 = "Your retention metrics are healthy, indicating customers perceive strong value in your offering. "
			para3 += "This foundation positions you well to explore pricing optimization. "
			para3 += "You have room to refine tier positioning and value communication without significant risk to customer loyalty."
		}
		
		// Early stage context
		if bizCtx.IsEarlyStage {
			para3 += " Given your growth stage, we recommend prioritizing customer acquisition and gathering product feedback over aggressive revenue optimization. Maintaining pricing flexibility now supports faster iteration toward product-market fit."
		}
	} else {
		para3 = "To provide retention-aware insights, add your business metrics (MRR, customer count, monthly churn) in Settings. "
		para3 += "Understanding your retention dynamics enables more informed pricing decisions that support sustainable growth."
	}
	
	if para3 != "" {
		paragraphs = append(paragraphs, para3)
	}
	
	return strings.Join(paragraphs, "\n\n")
}

// buildSummary creates a human-readable summary of the analysis with business context.
// This is the main entry point that calls the professional summary builder.
func (s *AnalysisService) buildSummary(numPlans, numCompetitors, belowCount, aroundCount, aboveCount int, medianPrice float64, bizCtx BusinessContext) string {
	// Note: This function signature is kept for compatibility, but we now use
	// buildProfessionalSummary internally which takes the full plans slice.
	// The actual implementation is in buildProfessionalSummaryFromCounts for backward compat.
	return s.buildProfessionalSummaryFromCounts(numPlans, belowCount, aroundCount, aboveCount, bizCtx)
}

// buildProfessionalSummaryFromCounts creates summary from count data (backward compatible)
func (s *AnalysisService) buildProfessionalSummaryFromCounts(numPlans, belowCount, aroundCount, aboveCount int, bizCtx BusinessContext) string {
	var paragraphs []string
	
	// PARAGRAPH 1: Overall structure
	var para1 string
	if numPlans == 1 {
		para1 = "You have a single pricing tier. Consider whether a multi-tier structure could better capture different customer segments and maximize revenue potential."
	} else {
		para1 = fmt.Sprintf("Your pricing structure consists of %d tiers. ", numPlans)
		
		// Position analysis
		if belowCount > 0 && aboveCount == 0 && aroundCount == 0 {
			para1 += "All your plans appear positioned on the lower end. While this can drive volume, you may be leaving revenue on the table."
		} else if aboveCount > 0 && belowCount == 0 && aroundCount == 0 {
			para1 += "Your pricing sits on the higher end. This premium positioning requires strong value differentiation to justify."
		} else if aroundCount == numPlans {
			para1 += "Your pricing appears well-balanced and competitively positioned."
		} else if belowCount > 0 && aboveCount > 0 {
			para1 += "Your tiers span a wide range, with some positioned lower and others higher. This spread can work well if each tier serves a distinct customer segment."
		}
	}
	paragraphs = append(paragraphs, para1)
	
	// PARAGRAPH 2: Churn & retention
	var para2 string
	if bizCtx.HasMetrics {
		if bizCtx.ChurnRate > 8 {
			para2 = fmt.Sprintf("Your monthly churn of %.1f%% signals retention challenges. Prioritize understanding customer needs and improving value before adjusting prices upward.", bizCtx.ChurnRate)
		} else if bizCtx.ChurnRate > 5 {
			para2 = fmt.Sprintf("With %.1f%% monthly churn, proceed carefully with any pricing changes. Monitor retention metrics closely.", bizCtx.ChurnRate)
		} else if bizCtx.ChurnRate > 0 {
			para2 = fmt.Sprintf("Strong retention (%.1f%% churn) suggests customers see good value. You likely have flexibility for pricing optimization.", bizCtx.ChurnRate)
		}
		
		if bizCtx.IsEarlyStage {
			para2 += " At your early stage, focus on acquisition and learning over revenue maximization."
		}
	} else {
		para2 = "Add business metrics in Settings for retention-aware recommendations."
	}
	
	if para2 != "" {
		paragraphs = append(paragraphs, para2)
	}
	
	return strings.Join(paragraphs, "\n\n")
}

// buildSummaryWithPlans creates the full professional summary when plans are available
func (s *AnalysisService) buildSummaryWithPlans(plans []*model.Plan, bizCtx BusinessContext) string {
	return s.buildProfessionalSummary(plans, bizCtx)
}

// ListAnalyses retrieves all analyses for a specific user.
func (s *AnalysisService) ListAnalyses(ctx context.Context, userID string) ([]*model.Analysis, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	return s.analysisRepo.ListByUser(ctx, uid)
}

// GetAnalysis retrieves a specific analysis by ID for a user.
func (s *AnalysisService) GetAnalysis(ctx context.Context, userID string, analysisID string) (*model.Analysis, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	aid, err := primitive.ObjectIDFromHex(analysisID)
	if err != nil {
		return nil, errors.New("invalid analysis id")
	}

	return s.analysisRepo.GetByIDAndUser(ctx, aid, uid)
}

// computeMedian calculates the median of a sorted slice of float64.
func computeMedian(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// roundToTwoDecimals rounds a float64 to two decimal places.
func roundToTwoDecimals(val float64) float64 {
	return float64(int(val*100+0.5)) / 100
}
