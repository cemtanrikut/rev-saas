package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rev-saas-api/internal/model"
)

// AnalysisV2Repository handles persistence for V2 analysis results.
type AnalysisV2Repository struct {
	collection *mongo.Collection
}

// NewAnalysisV2Repository creates a new V2 analysis repository.
func NewAnalysisV2Repository(db *mongo.Database) *AnalysisV2Repository {
	return &AnalysisV2Repository{
		collection: db.Collection("analyses_v2"),
	}
}

// Create inserts a new V2 analysis result.
func (r *AnalysisV2Repository) Create(ctx context.Context, analysis *model.AnalysisResultV2) error {
	if analysis.CreatedAt.IsZero() {
		analysis.CreatedAt = time.Now()
	}

	result, err := r.collection.InsertOne(ctx, analysis)
	if err != nil {
		return err
	}

	analysis.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByIDAndUser retrieves a V2 analysis by ID for a specific user.
func (r *AnalysisV2Repository) GetByIDAndUser(ctx context.Context, id, userID primitive.ObjectID) (*model.AnalysisResultV2, error) {
	filter := bson.M{
		"_id":     id,
		"user_id": userID,
	}

	var analysis model.AnalysisResultV2
	err := r.collection.FindOne(ctx, filter).Decode(&analysis)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &analysis, nil
}

// ListByUser retrieves V2 analyses for a user, sorted by creation date (newest first).
func (r *AnalysisV2Repository) ListByUser(ctx context.Context, userID primitive.ObjectID, limit int) ([]*model.AnalysisResultV2, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var analyses []*model.AnalysisResultV2
	if err := cursor.All(ctx, &analyses); err != nil {
		return nil, err
	}

	return analyses, nil
}

// GetLatestByUser retrieves the most recent V2 analysis for a user.
func (r *AnalysisV2Repository) GetLatestByUser(ctx context.Context, userID primitive.ObjectID) (*model.AnalysisResultV2, error) {
	filter := bson.M{"user_id": userID}
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var analysis model.AnalysisResultV2
	err := r.collection.FindOne(ctx, filter, opts).Decode(&analysis)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &analysis, nil
}




