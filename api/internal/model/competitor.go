package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CompetitorPlan represents a pricing plan of a competitor.
type CompetitorPlan struct {
	Name         string  `bson:"name" json:"name"`
	Price        float64 `bson:"price" json:"price"`
	Currency     string  `bson:"currency,omitempty" json:"currency,omitempty"`
	BillingCycle string  `bson:"billing_cycle,omitempty" json:"billing_cycle,omitempty"`
}

// Competitor represents a competitor belonging to a user.
type Competitor struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name      string             `bson:"name" json:"name"`
	URL       string             `bson:"url,omitempty" json:"url,omitempty"`
	Plans     []CompetitorPlan   `bson:"plans" json:"plans"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
