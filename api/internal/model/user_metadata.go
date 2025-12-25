package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserMetadata stores additional signup/onboarding information about a user.
type UserMetadata struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	HeardFrom string             `bson:"heard_from,omitempty" json:"heard_from,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}







