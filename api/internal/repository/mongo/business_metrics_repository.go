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

// BusinessMetricsRepository handles business metrics data operations in MongoDB.
type BusinessMetricsRepository struct {
	collection *mongo.Collection
}

// NewBusinessMetricsRepository creates a new BusinessMetricsRepository.
func NewBusinessMetricsRepository(db *mongo.Database) *BusinessMetricsRepository {
	return &BusinessMetricsRepository{
		collection: db.Collection("business_metrics"),
	}
}

// GetByUserID retrieves business metrics for a specific user.
func (r *BusinessMetricsRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) (*model.BusinessMetrics, error) {
	var metrics model.BusinessMetrics
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&metrics)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &metrics, nil
}

// UpsertForUser updates or inserts business metrics for a user.
func (r *BusinessMetricsRepository) UpsertForUser(ctx context.Context, metrics *model.BusinessMetrics) error {
	metrics.UpdatedAt = time.Now().UTC()

	filter := bson.M{"user_id": metrics.UserID}
	update := bson.M{
		"$set": bson.M{
			"currency":           metrics.Currency,
			"mrr":                metrics.MRR,
			"customers":          metrics.Customers,
			"monthly_churn_rate": metrics.MonthlyChurnRate,
			"updated_at":         metrics.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"user_id": metrics.UserID,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	// If it was an insert, get the new ID
	if result.UpsertedID != nil {
		if oid, ok := result.UpsertedID.(primitive.ObjectID); ok {
			metrics.ID = oid
		}
	} else {
		// If it was an update, fetch the existing document to get the ID
		existing, err := r.GetByUserID(ctx, metrics.UserID)
		if err != nil {
			return err
		}
		if existing != nil {
			metrics.ID = existing.ID
		}
	}

	return nil
}

