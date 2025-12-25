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

const stripeConnectionsCollection = "stripe_connections"

// StripeConnectionRepository handles Stripe connection persistence.
type StripeConnectionRepository struct {
	col *mongo.Collection
}

// NewStripeConnectionRepository creates a new StripeConnectionRepository.
func NewStripeConnectionRepository(db *mongo.Database) *StripeConnectionRepository {
	return &StripeConnectionRepository{
		col: db.Collection(stripeConnectionsCollection),
	}
}

// UpsertConnection inserts or updates a Stripe connection for a user.
func (r *StripeConnectionRepository) UpsertConnection(ctx context.Context, conn *model.StripeConnection) error {
	conn.UpdatedAt = time.Now()

	filter := bson.M{"user_id": conn.UserID}
	update := bson.M{
		"$set": bson.M{
			"stripe_account_id":        conn.StripeAccountID,
			"access_token_encrypted":   conn.AccessTokenEncrypted,
			"refresh_token_encrypted":  conn.RefreshTokenEncrypted,
			"token_type":               conn.TokenType,
			"scope":                    conn.Scope,
			"livemode":                 conn.Livemode,
			"status":                   conn.Status,
			"updated_at":               conn.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"user_id":      conn.UserID,
			"connected_at": conn.ConnectedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.col.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetByUserID retrieves the Stripe connection for a user.
func (r *StripeConnectionRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) (*model.StripeConnection, error) {
	var conn model.StripeConnection
	filter := bson.M{"user_id": userID}

	err := r.col.FindOne(ctx, filter).Decode(&conn)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No connection found
		}
		return nil, err
	}

	return &conn, nil
}

// MarkDisconnected marks a connection as disconnected without deleting it.
func (r *StripeConnectionRepository) MarkDisconnected(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"status":                  model.StripeStatusDisconnected,
			"access_token_encrypted":  "",
			"refresh_token_encrypted": "",
			"updated_at":              time.Now(),
		},
	}

	_, err := r.col.UpdateOne(ctx, filter, update)
	return err
}

// UpdateLastSyncAt updates the last_sync_at timestamp.
func (r *StripeConnectionRepository) UpdateLastSyncAt(ctx context.Context, userID primitive.ObjectID) error {
	now := time.Now()
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"last_sync_at": now,
			"updated_at":   now,
		},
	}

	_, err := r.col.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes a Stripe connection completely.
func (r *StripeConnectionRepository) Delete(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID}
	_, err := r.col.DeleteOne(ctx, filter)
	return err
}




