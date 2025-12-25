package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"rev-saas-api/internal/config"
	"rev-saas-api/internal/model"
	mongorepo "rev-saas-api/internal/repository/mongo"
)

var (
	// ErrInvalidPriceChange is returned when price values are invalid.
	ErrInvalidPriceChange = errors.New("current price must be greater than zero")
	// ErrNoBucketFound is returned when no matching elasticity bucket is found.
	ErrNoBucketFound = errors.New("no matching price change bucket found")
	// ErrNoProfileFound is returned when the pricing goal profile is not found.
	ErrNoProfileFound = errors.New("pricing goal profile not found")
)

// SimulationService handles pricing simulation logic.
type SimulationService struct {
	elasticityConfig *config.ElasticityConfig
	simulationRepo   *mongorepo.SimulationRepository
	planRepo         *mongorepo.PlanRepository
	aiService        *AIPricingService
}

// NewSimulationService creates a new SimulationService.
func NewSimulationService(
	elasticityConfig *config.ElasticityConfig,
	simulationRepo *mongorepo.SimulationRepository,
	planRepo *mongorepo.PlanRepository,
	aiService *AIPricingService,
) *SimulationService {
	return &SimulationService{
		elasticityConfig: elasticityConfig,
		simulationRepo:   simulationRepo,
		planRepo:         planRepo,
		aiService:        aiService,
	}
}

// RunPricingSimulation runs a deterministic pricing simulation based on the request.
func (s *SimulationService) RunPricingSimulation(ctx context.Context, userID string, req model.SimulationRequest) (*model.SimulationResult, error) {
	// Validate input
	if req.CurrentPrice <= 0 {
		return nil, ErrInvalidPriceChange
	}
	if req.NewPrice < 0 {
		return nil, errors.New("new price cannot be negative")
	}
	if req.ActiveCustomersOnPlan < 0 {
		return nil, errors.New("active customers cannot be negative")
	}

	// Parse user ID
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	// Parse plan ID
	planID, err := primitive.ObjectIDFromHex(req.PlanID)
	if err != nil {
		return nil, errors.New("invalid plan id")
	}

	// Look up plan to get the name and verify ownership
	plan, err := s.planRepo.GetByIDAndUser(ctx, planID, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plan: %w", err)
	}
	if plan == nil {
		return nil, errors.New("plan not found or access denied")
	}

	// Compute price change percentage
	priceChangePct := ((req.NewPrice - req.CurrentPrice) / req.CurrentPrice) * 100

	// Normalize pricing goal
	pricingGoal := model.NormalizePricingGoal(req.PricingGoal)

	// Find the matching bucket
	bucket := s.elasticityConfig.FindBucket(priceChangePct)
	if bucket == nil {
		return nil, ErrNoBucketFound
	}

	// Get the profile for the pricing goal
	profile, ok := bucket.Profiles[pricingGoal]
	if !ok {
		return nil, ErrNoProfileFound
	}

	// Get churn multiplier
	churnMultiplier := s.elasticityConfig.GetChurnMultiplier(req.GlobalChurnRate)

	// Absolute price change for risk calculation
	absPriceChangePct := math.Abs(priceChangePct)

	// Determine if this is a price increase or decrease
	isPriceIncrease := priceChangePct >= 0

	// Calculate scenarios
	scenarios := []model.SimulationScenario{
		s.calculateScenario("Conservative", profile.Conservative, req, isPriceIncrease, churnMultiplier, absPriceChangePct),
		s.calculateScenario("Base", profile.Base, req, isPriceIncrease, churnMultiplier, absPriceChangePct),
		s.calculateScenario("Aggressive", profile.Aggressive, req, isPriceIncrease, churnMultiplier, absPriceChangePct),
	}

	// Build result
	result := &model.SimulationResult{
		UserID:                uid,
		PlanID:                planID,
		PlanName:              plan.Name,
		CurrentPrice:          req.CurrentPrice,
		NewPrice:              req.NewPrice,
		Currency:              req.Currency,
		PriceChangePct:        math.Round(priceChangePct*100) / 100, // Round to 2 decimal places
		ActiveCustomersOnPlan: req.ActiveCustomersOnPlan,
		GlobalMRR:             req.GlobalMRR,
		GlobalChurnRate:       req.GlobalChurnRate,
		PricingGoal:           pricingGoal,
		Scenarios:             scenarios,
		CreatedAt:             time.Now().UTC(),
	}

	return result, nil
}

// calculateScenario calculates a single scenario based on the band configuration.
func (s *SimulationService) calculateScenario(
	name string,
	band config.ScenarioBand,
	req model.SimulationRequest,
	isPriceIncrease bool,
	churnMultiplier float64,
	absPriceChangePct float64,
) model.SimulationScenario {
	scenario := model.SimulationScenario{
		Name: name,
	}

	customers := req.ActiveCustomersOnPlan

	if isPriceIncrease {
		// Price increase: customers may leave
		lossMinPct := band.CustomerLossMinPct * churnMultiplier
		lossMaxPct := band.CustomerLossMaxPct * churnMultiplier

		// Cap at 100%
		lossMinPct = math.Min(lossMinPct, 100)
		lossMaxPct = math.Min(lossMaxPct, 100)

		scenario.CustomerLossMinPct = math.Round(lossMinPct*100) / 100
		scenario.CustomerLossMaxPct = math.Round(lossMaxPct*100) / 100

		// Calculate new customer counts (lose customers)
		lossMin := int(math.Round(float64(customers) * lossMinPct / 100))
		lossMax := int(math.Round(float64(customers) * lossMaxPct / 100))

		scenario.NewCustomerCountMin = customers - lossMax // Min customers = original - max loss
		scenario.NewCustomerCountMax = customers - lossMin // Max customers = original - min loss

		// Ensure non-negative
		if scenario.NewCustomerCountMin < 0 {
			scenario.NewCustomerCountMin = 0
		}
		if scenario.NewCustomerCountMax < 0 {
			scenario.NewCustomerCountMax = 0
		}
	} else {
		// Price decrease: customers may be gained
		gainMinPct := band.CustomerGainMinPct * churnMultiplier
		gainMaxPct := band.CustomerGainMaxPct * churnMultiplier

		scenario.CustomerGainMinPct = math.Round(gainMinPct*100) / 100
		scenario.CustomerGainMaxPct = math.Round(gainMaxPct*100) / 100

		// Calculate new customer counts (gain customers)
		gainMin := int(math.Round(float64(customers) * gainMinPct / 100))
		gainMax := int(math.Round(float64(customers) * gainMaxPct / 100))

		scenario.NewCustomerCountMin = customers + gainMin
		scenario.NewCustomerCountMax = customers + gainMax
	}

	// Calculate new MRR and ARR
	scenario.NewMRRMin = math.Round(float64(scenario.NewCustomerCountMin)*req.NewPrice*100) / 100
	scenario.NewMRRMax = math.Round(float64(scenario.NewCustomerCountMax)*req.NewPrice*100) / 100

	scenario.NewARRMin = math.Round(scenario.NewMRRMin*12*100) / 100
	scenario.NewARRMax = math.Round(scenario.NewMRRMax*12*100) / 100

	// Estimate churn adjustment
	// For price increases, churn may increase slightly
	// For price decreases, churn may decrease slightly
	baseChurn := req.GlobalChurnRate
	if isPriceIncrease {
		// Price increase: churn may increase
		churnIncrease := absPriceChangePct * 0.05 // 5% of price change adds to churn
		scenario.EstimatedChurnMinPct = math.Round((baseChurn+churnIncrease*0.5)*100) / 100
		scenario.EstimatedChurnMaxPct = math.Round((baseChurn+churnIncrease)*100) / 100
	} else {
		// Price decrease: churn may decrease
		churnDecrease := absPriceChangePct * 0.03 // 3% of price change reduces churn
		scenario.EstimatedChurnMinPct = math.Max(0, math.Round((baseChurn-churnDecrease)*100)/100)
		scenario.EstimatedChurnMaxPct = math.Max(0, math.Round((baseChurn-churnDecrease*0.5)*100)/100)
	}

	// Derive risk level
	scenarioLevel := "base"
	switch name {
	case "Conservative":
		scenarioLevel = "conservative"
	case "Aggressive":
		scenarioLevel = "aggressive"
	}
	scenario.RiskLevel = s.elasticityConfig.DeriveRiskLevel(absPriceChangePct, scenarioLevel)

	return scenario
}

// SaveSimulation persists a simulation result to the database.
func (s *SimulationService) SaveSimulation(ctx context.Context, result *model.SimulationResult) error {
	return s.simulationRepo.Create(ctx, result)
}

// ListSimulations returns all simulations for a user, optionally filtered by plan.
func (s *SimulationService) ListSimulations(ctx context.Context, userID string, planID string, limit int) ([]*model.SimulationResult, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	var pid *primitive.ObjectID
	if planID != "" {
		parsedPID, err := primitive.ObjectIDFromHex(planID)
		if err != nil {
			return nil, errors.New("invalid plan id")
		}
		pid = &parsedPID
	}

	return s.simulationRepo.ListByUser(ctx, uid, pid, limit)
}

// GetSimulationByID retrieves a simulation by ID, ensuring it belongs to the user.
func (s *SimulationService) GetSimulationByID(ctx context.Context, userID, simulationID string) (*model.SimulationResult, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	sid, err := primitive.ObjectIDFromHex(simulationID)
	if err != nil {
		return nil, errors.New("invalid simulation id")
	}

	return s.simulationRepo.GetByIDAndUser(ctx, sid, uid)
}






