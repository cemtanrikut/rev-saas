package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Plan represents a pricing plan belonging to a user.
type Plan struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name      string             `bson:"name" json:"name"`
	Price     float64            `bson:"price" json:"price"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

