package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Competitor represents a competitor belonging to a user.
type Competitor struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name      string             `bson:"name" json:"name"`
	URL       string             `bson:"url" json:"url"`
	BasePrice float64            `bson:"base_price" json:"base_price"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

