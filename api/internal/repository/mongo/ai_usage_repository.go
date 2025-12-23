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

// AIUsageRepository handles AI usage data persistence.
type AIUsageRepository struct {
	collection *mongo.Collection
}

// NewAIUsageRepository creates a new AIUsageRepository.
func NewAIUsageRepository(db *mongo.Database) *AIUsageRepository {
	return &AIUsageRepository{
		collection: db.Collection("ai_usage"),
	}
}

// GetUsage retrieves the AI usage for a user in a specific month.
// Returns 0 if no usage record exists.
func (r *AIUsageRepository) GetUsage(ctx context.Context, userID primitive.ObjectID, monthKey string) (int, error) {
	filter := bson.M{
		"user_id":   userID,
		"month_key": monthKey,
	}

	var usage model.AIUsage
	err := r.collection.FindOne(ctx, filter).Decode(&usage)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil // No usage record = 0 credits used
		}
		return 0, err
	}

	return usage.UsedCredits, nil
}

// IncrementUsage increments the used credits for a user in a specific month.
// Creates a new record if one doesn't exist.
func (r *AIUsageRepository) IncrementUsage(ctx context.Context, userID primitive.ObjectID, monthKey string, delta int) error {
	filter := bson.M{
		"user_id":   userID,
		"month_key": monthKey,
	}

	now := time.Now()

	update := bson.M{
		"$inc": bson.M{"used_credits": delta},
		"$set": bson.M{"updated_at": now},
		"$setOnInsert": bson.M{
			"user_id":    userID,
			"month_key":  monthKey,
			"created_at": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetUsageByUser retrieves all AI usage records for a user (for analytics/history).
func (r *AIUsageRepository) GetUsageByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.AIUsage, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "month_key", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var usages []*model.AIUsage
	if err := cursor.All(ctx, &usages); err != nil {
		return nil, err
	}

	return usages, nil
}

// ResetCredits resets the used credits for a user in a specific month to 0.
// This is called when a subscription renews (invoice.paid).
// Uses upsert to create the record if it doesn't exist.
func (r *AIUsageRepository) ResetCredits(ctx context.Context, userIDStr string, monthKey string) error {
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return err
	}

	now := time.Now()
	filter := bson.M{
		"user_id":   userID,
		"month_key": monthKey,
	}

	update := bson.M{
		"$set": bson.M{
			"used_credits": 0,
			"updated_at":   now,
		},
		"$setOnInsert": bson.M{
			"user_id":    userID,
			"month_key":  monthKey,
			"created_at": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}



