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

// CompetitorRepository handles competitor data operations in MongoDB.
type CompetitorRepository struct {
	collection *mongo.Collection
}

// NewCompetitorRepository creates a new CompetitorRepository.
func NewCompetitorRepository(db *mongo.Database) *CompetitorRepository {
	return &CompetitorRepository{
		collection: db.Collection("competitors"),
	}
}

// Create inserts a new competitor into the database.
func (r *CompetitorRepository) Create(ctx context.Context, competitor *model.Competitor) error {
	competitor.CreatedAt = time.Now().UTC()
	result, err := r.collection.InsertOne(ctx, competitor)
	if err != nil {
		return err
	}
	// Set the ID on the competitor object
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		competitor.ID = oid
	}
	return nil
}

// ListByUser retrieves all competitors for a specific user.
func (r *CompetitorRepository) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.Competitor, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var competitors []*model.Competitor
	for cursor.Next(ctx) {
		var c model.Competitor
		if err := cursor.Decode(&c); err != nil {
			return nil, err
		}
		competitors = append(competitors, &c)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return competitors, nil
}

// GetByID retrieves a competitor by its ID.
func (r *CompetitorRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Competitor, error) {
	var competitor model.Competitor
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&competitor)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &competitor, nil
}

// DeleteByIDAndUser deletes a competitor by ID, ensuring it belongs to the specified user.
func (r *CompetitorRepository) DeleteByIDAndUser(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	filter := bson.M{
		"_id":     id,
		"user_id": userID,
	}

	res, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// UpdateByIDAndUser updates a competitor by ID, ensuring it belongs to the specified user.
func (r *CompetitorRepository) UpdateByIDAndUser(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, update *model.Competitor) error {
	filter := bson.M{
		"_id":     id,
		"user_id": userID,
	}

	updateDoc := bson.M{
		"$set": bson.M{
			"name":  update.Name,
			"url":   update.URL,
			"plans": update.Plans,
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}


