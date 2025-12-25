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

// BillingSubscriptionRepository handles billing subscription persistence.
type BillingSubscriptionRepository struct {
	collection *mongo.Collection
}

// NewBillingSubscriptionRepository creates a new billing subscription repository.
func NewBillingSubscriptionRepository(db *mongo.Database) *BillingSubscriptionRepository {
	coll := db.Collection("billing_subscriptions")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "stripe_customer_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "stripe_subscription_id", Value: 1}},
		},
	}

	_, _ = coll.Indexes().CreateMany(ctx, indexes)

	return &BillingSubscriptionRepository{collection: coll}
}

// GetByUserID retrieves a subscription by user ID.
func (r *BillingSubscriptionRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) (*model.BillingSubscription, error) {
	var sub model.BillingSubscription
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&sub)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetByStripeCustomerID retrieves a subscription by Stripe customer ID.
func (r *BillingSubscriptionRepository) GetByStripeCustomerID(ctx context.Context, customerID string) (*model.BillingSubscription, error) {
	var sub model.BillingSubscription
	err := r.collection.FindOne(ctx, bson.M{"stripe_customer_id": customerID}).Decode(&sub)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetByStripeSubscriptionID retrieves a subscription by Stripe subscription ID.
func (r *BillingSubscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, subscriptionID string) (*model.BillingSubscription, error) {
	var sub model.BillingSubscription
	err := r.collection.FindOne(ctx, bson.M{"stripe_subscription_id": subscriptionID}).Decode(&sub)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// Upsert creates or updates a subscription for a user.
func (r *BillingSubscriptionRepository) Upsert(ctx context.Context, sub *model.BillingSubscription) error {
	now := time.Now()
	sub.UpdatedAt = now

	filter := bson.M{"user_id": sub.UserID}
	update := bson.M{
		"$set": bson.M{
			"stripe_customer_id":     sub.StripeCustomerID,
			"stripe_subscription_id": sub.StripeSubscriptionID,
			"stripe_price_id":        sub.StripePriceID,
			"plan_key":               sub.PlanKey,
			"status":                 sub.Status,
			"cancel_at_period_end":   sub.CancelAtPeriodEnd,
			"current_period_end":     sub.CurrentPeriodEnd,
			"updated_at":             now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// UpsertByStripeCustomerID creates or updates a subscription by Stripe customer ID.
func (r *BillingSubscriptionRepository) UpsertByStripeCustomerID(ctx context.Context, sub *model.BillingSubscription) error {
	now := time.Now()
	sub.UpdatedAt = now

	filter := bson.M{"stripe_customer_id": sub.StripeCustomerID}
	update := bson.M{
		"$set": bson.M{
			"user_id":                sub.UserID,
			"stripe_subscription_id": sub.StripeSubscriptionID,
			"stripe_price_id":        sub.StripePriceID,
			"plan_key":               sub.PlanKey,
			"status":                 sub.Status,
			"cancel_at_period_end":   sub.CancelAtPeriodEnd,
			"current_period_end":     sub.CurrentPeriodEnd,
			"updated_at":             now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// UpdateStatus updates just the status of a subscription.
func (r *BillingSubscriptionRepository) UpdateStatus(ctx context.Context, userID primitive.ObjectID, status model.SubscriptionStatus) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes a subscription.
func (r *BillingSubscriptionRepository) Delete(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}

// WebhookEventRepository handles webhook event idempotency.
type WebhookEventRepository struct {
	collection *mongo.Collection
}

// NewWebhookEventRepository creates a new webhook event repository.
func NewWebhookEventRepository(db *mongo.Database) *WebhookEventRepository {
	coll := db.Collection("webhook_events")

	// Create unique index on event_id
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "event_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, _ = coll.Indexes().CreateOne(ctx, index)

	// TTL index to auto-delete old events after 30 days
	ttlIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "processed_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(30 * 24 * 60 * 60), // 30 days
	}
	_, _ = coll.Indexes().CreateOne(ctx, ttlIndex)

	return &WebhookEventRepository{collection: coll}
}

// IsProcessed checks if an event has already been processed.
func (r *WebhookEventRepository) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"event_id": eventID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// MarkProcessed marks an event as processed.
func (r *WebhookEventRepository) MarkProcessed(ctx context.Context, eventID, eventType string) error {
	event := model.WebhookEvent{
		EventID:     eventID,
		EventType:   eventType,
		ProcessedAt: time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, event)
	// Ignore duplicate key errors (event already processed)
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}



