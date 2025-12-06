package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Recommendation represents a pricing recommendation for a specific plan.
type Recommendation struct {
	PlanID            primitive.ObjectID `bson:"plan_id" json:"plan_id"`
	PlanName          string             `bson:"plan_name" json:"plan_name"`
	CurrentPrice      float64            `bson:"current_price" json:"current_price"`
	Position          string             `bson:"position" json:"position"`                       // "lowest", "below_median", "around_median", "above_median", "highest"
	SuggestedAction   string             `bson:"suggested_action" json:"suggested_action"`       // "raise_price", "lower_price", "keep", "restructure"
	SuggestedNewPrice float64            `bson:"suggested_new_price" json:"suggested_new_price"` // 0 if no change suggested
	Rationale         string             `bson:"rationale" json:"rationale"`
}

// Analysis represents a pricing analysis result for a user.
type Analysis struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	NumPlans        int                `bson:"num_plans" json:"num_plans"`
	NumCompetitors  int                `bson:"num_competitors" json:"num_competitors"`
	Summary         string             `bson:"summary" json:"summary"`
	Recommendations []Recommendation   `bson:"recommendations" json:"recommendations"`

	// AI-generated insights (only for paid plans)
	AISummary   string   `bson:"ai_summary,omitempty" json:"ai_summary,omitempty"`
	AIScenarios []string `bson:"ai_scenarios,omitempty" json:"ai_scenarios,omitempty"`
}

