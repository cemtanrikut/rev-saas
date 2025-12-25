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

// SimulationRepository handles simulation data operations in MongoDB.
type SimulationRepository struct {
	collection *mongo.Collection
}

// NewSimulationRepository creates a new SimulationRepository.
func NewSimulationRepository(db *mongo.Database) *SimulationRepository {
	return &SimulationRepository{
		collection: db.Collection("pricing_simulations"),
	}
}

// Create inserts a new simulation result into the database.
func (r *SimulationRepository) Create(ctx context.Context, simulation *model.SimulationResult) error {
	simulation.CreatedAt = time.Now().UTC()
	result, err := r.collection.InsertOne(ctx, simulation)
	if err != nil {
		return err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		simulation.ID = oid
	}
	return nil
}

// ListByUser retrieves simulations for a user, optionally filtered by plan ID.
func (r *SimulationRepository) ListByUser(ctx context.Context, userID primitive.ObjectID, planID *primitive.ObjectID, limit int) ([]*model.SimulationResult, error) {
	filter := bson.M{"user_id": userID}
	
	if planID != nil {
		filter["plan_id"] = *planID
	}

	// Default limit
	if limit <= 0 {
		limit = 50
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var simulations []*model.SimulationResult
	for cursor.Next(ctx) {
		var s model.SimulationResult
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		simulations = append(simulations, &s)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return simulations, nil
}

// GetByID retrieves a simulation by its ID.
func (r *SimulationRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.SimulationResult, error) {
	var simulation model.SimulationResult
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&simulation)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &simulation, nil
}

// GetByIDAndUser retrieves a simulation by ID and ensures it belongs to the specified user.
func (r *SimulationRepository) GetByIDAndUser(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*model.SimulationResult, error) {
	var simulation model.SimulationResult
	filter := bson.M{
		"_id":     id,
		"user_id": userID,
	}
	err := r.collection.FindOne(ctx, filter).Decode(&simulation)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &simulation, nil
}

// Update updates an existing simulation (e.g., to add AI narrative).
func (r *SimulationRepository) Update(ctx context.Context, simulation *model.SimulationResult) error {
	filter := bson.M{"_id": simulation.ID}
	update := bson.M{
		"$set": bson.M{
			"ai_narrative": simulation.AINarrative,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// CountByUser returns the total number of simulations for a user.
func (r *SimulationRepository) CountByUser(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"user_id": userID})
}

// DeleteByIDAndUser deletes a simulation by ID, ensuring it belongs to the specified user.
func (r *SimulationRepository) DeleteByIDAndUser(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
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






