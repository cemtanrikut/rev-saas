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

// AnalysisRepository handles analysis data operations in MongoDB.
type AnalysisRepository struct {
	collection *mongo.Collection
}

// NewAnalysisRepository creates a new AnalysisRepository.
func NewAnalysisRepository(db *mongo.Database) *AnalysisRepository {
	return &AnalysisRepository{
		collection: db.Collection("analyses"),
	}
}

// Create inserts a new analysis into the database.
func (r *AnalysisRepository) Create(ctx context.Context, analysis *model.Analysis) error {
	analysis.CreatedAt = time.Now().UTC()
	result, err := r.collection.InsertOne(ctx, analysis)
	if err != nil {
		return err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		analysis.ID = oid
	}
	return nil
}

// ListByUser retrieves all analyses for a specific user, sorted by CreatedAt descending.
func (r *AnalysisRepository) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.Analysis, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by creation date descending (newest first)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var analyses []*model.Analysis
	for cursor.Next(ctx) {
		var a model.Analysis
		if err := cursor.Decode(&a); err != nil {
			return nil, err
		}
		analyses = append(analyses, &a)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return analyses, nil
}

// GetByID retrieves an analysis by its ID.
func (r *AnalysisRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Analysis, error) {
	var analysis model.Analysis
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&analysis)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

// GetByIDAndUser retrieves an analysis by its ID and ensures it belongs to the specified user.
func (r *AnalysisRepository) GetByIDAndUser(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*model.Analysis, error) {
	var analysis model.Analysis
	filter := bson.M{
		"_id":     id,
		"user_id": userID,
	}
	err := r.collection.FindOne(ctx, filter).Decode(&analysis)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

// Update updates an existing analysis in the database.
func (r *AnalysisRepository) Update(ctx context.Context, analysis *model.Analysis) error {
	filter := bson.M{"_id": analysis.ID}
	update := bson.M{
		"$set": bson.M{
			"ai_summary":   analysis.AISummary,
			"ai_scenarios": analysis.AIScenarios,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

