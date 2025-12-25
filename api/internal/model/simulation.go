package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SimulationRequest represents the input for a pricing simulation.
type SimulationRequest struct {
	PlanID                string  `json:"plan_id"`
	CurrentPrice          float64 `json:"current_price"`
	NewPrice              float64 `json:"new_price"`
	Currency              string  `json:"currency"`
	ActiveCustomersOnPlan int     `json:"active_customers_on_plan"`
	GlobalMRR             float64 `json:"global_mrr"`
	GlobalChurnRate       float64 `json:"global_churn_rate"`
	PricingGoal           string  `json:"pricing_goal"` // "revenue" | "retention" | "conversion" | "differentiation"
}

// SimulationScenario represents one of the three scenarios in a simulation result.
type SimulationScenario struct {
	Name                 string  `bson:"name" json:"name"`                                       // "Conservative" | "Base" | "Aggressive"
	CustomerLossMinPct   float64 `bson:"customer_loss_min_pct" json:"customer_loss_min_pct"`     // For price increase
	CustomerLossMaxPct   float64 `bson:"customer_loss_max_pct" json:"customer_loss_max_pct"`     // For price increase
	CustomerGainMinPct   float64 `bson:"customer_gain_min_pct" json:"customer_gain_min_pct"`     // For price decrease
	CustomerGainMaxPct   float64 `bson:"customer_gain_max_pct" json:"customer_gain_max_pct"`     // For price decrease
	NewCustomerCountMin  int     `bson:"new_customer_count_min" json:"new_customer_count_min"`
	NewCustomerCountMax  int     `bson:"new_customer_count_max" json:"new_customer_count_max"`
	NewMRRMin            float64 `bson:"new_mrr_min" json:"new_mrr_min"`
	NewMRRMax            float64 `bson:"new_mrr_max" json:"new_mrr_max"`
	NewARRMin            float64 `bson:"new_arr_min" json:"new_arr_min"`
	NewARRMax            float64 `bson:"new_arr_max" json:"new_arr_max"`
	EstimatedChurnMinPct float64 `bson:"estimated_churn_min_pct" json:"estimated_churn_min_pct"`
	EstimatedChurnMaxPct float64 `bson:"estimated_churn_max_pct" json:"estimated_churn_max_pct"`
	RiskLevel            string  `bson:"risk_level" json:"risk_level"` // "low" | "medium" | "high"
}

// SimulationResult represents the complete result of a pricing simulation.
type SimulationResult struct {
	ID                    primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID                primitive.ObjectID   `bson:"user_id" json:"user_id"`
	PlanID                primitive.ObjectID   `bson:"plan_id" json:"plan_id"`
	PlanName              string               `bson:"plan_name" json:"plan_name"`
	CurrentPrice          float64              `bson:"current_price" json:"current_price"`
	NewPrice              float64              `bson:"new_price" json:"new_price"`
	Currency              string               `bson:"currency" json:"currency"`
	PriceChangePct        float64              `bson:"price_change_pct" json:"price_change_pct"`
	ActiveCustomersOnPlan int                  `bson:"active_customers_on_plan" json:"active_customers_on_plan"`
	GlobalMRR             float64              `bson:"global_mrr" json:"global_mrr"`
	GlobalChurnRate       float64              `bson:"global_churn_rate" json:"global_churn_rate"`
	PricingGoal           string               `bson:"pricing_goal" json:"pricing_goal"`
	Scenarios             []SimulationScenario `bson:"scenarios" json:"scenarios"`
	AINarrative           string               `bson:"ai_narrative,omitempty" json:"ai_narrative,omitempty"`
	CreatedAt             time.Time            `bson:"created_at" json:"created_at"`
}

// ValidPricingGoals contains all valid pricing goal values.
var ValidPricingGoals = map[string]bool{
	"revenue":         true,
	"retention":       true,
	"conversion":      true,
	"differentiation": true,
}

// NormalizePricingGoal returns a valid pricing goal, defaulting to "revenue" if unknown.
func NormalizePricingGoal(goal string) string {
	if ValidPricingGoals[goal] {
		return goal
	}
	return "revenue"
}






