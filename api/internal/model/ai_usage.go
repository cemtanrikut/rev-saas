package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AIUsage tracks AI credit usage per user per month.
type AIUsage struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	MonthKey    string             `bson:"month_key" json:"month_key"` // Format: "2025-12"
	UsedCredits int                `bson:"used_credits" json:"used_credits"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// AICreditsResponse is the response for GET /api/ai-credits endpoint.
type AICreditsResponse struct {
	PlanType           string `json:"plan_type"`
	MonthlyCredits     int    `json:"monthly_credits"`
	UsedCredits        int    `json:"used_credits"`
	RemainingCredits   int    `json:"remaining_credits"`
	SimulationsEnabled bool   `json:"simulations_enabled"`
	MonthKey           string `json:"month_key"`
}





