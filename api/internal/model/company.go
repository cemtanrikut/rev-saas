package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company represents a company/workspace belonging to a user.
type Company struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name      string             `bson:"name" json:"name"`
	Website   string             `bson:"website" json:"website"`
	MRRRange  string             `bson:"mrr_range,omitempty" json:"mrr_range,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}







