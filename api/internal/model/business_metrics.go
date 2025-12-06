package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BusinessMetrics represents a user's business metrics.
type BusinessMetrics struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	Currency         string             `bson:"currency" json:"currency"`
	MRR              float64            `bson:"mrr" json:"mrr"`
	Customers        int                `bson:"customers" json:"customers"`
	MonthlyChurnRate float64            `bson:"monthly_churn_rate" json:"monthly_churn_rate"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

