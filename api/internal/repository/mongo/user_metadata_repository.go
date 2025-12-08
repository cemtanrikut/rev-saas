package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"rev-saas-api/internal/model"
)

// UserMetadataRepository handles user metadata operations in MongoDB.
type UserMetadataRepository struct {
	collection *mongo.Collection
}

// NewUserMetadataRepository creates a new UserMetadataRepository.
func NewUserMetadataRepository(db *mongo.Database) *UserMetadataRepository {
	return &UserMetadataRepository{
		collection: db.Collection("user_metadata"),
	}
}

// Create inserts new user metadata into the database.
func (r *UserMetadataRepository) Create(ctx context.Context, metadata *model.UserMetadata) error {
	metadata.CreatedAt = time.Now().UTC()
	result, err := r.collection.InsertOne(ctx, metadata)
	if err != nil {
		return err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		metadata.ID = oid
	}
	return nil
}

// GetByUserID retrieves metadata for a specific user.
func (r *UserMetadataRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) (*model.UserMetadata, error) {
	var metadata model.UserMetadata
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&metadata)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

// Update updates user metadata.
func (r *UserMetadataRepository) Update(ctx context.Context, metadata *model.UserMetadata) error {
	filter := bson.M{"_id": metadata.ID}
	update := bson.M{
		"$set": bson.M{
			"heard_from": metadata.HeardFrom,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}


