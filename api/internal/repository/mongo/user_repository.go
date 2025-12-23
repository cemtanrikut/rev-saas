package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"rev-saas-api/internal/model"
)

// UserRepository handles user data operations in MongoDB.
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	// Set the ID on the user object
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	}
	return nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by their ID (ObjectID).
func (r *UserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByIDString retrieves a user by their ID (string).
func (r *UserRepository) GetByIDString(ctx context.Context, id string) (*model.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, oid)
}

// Update updates a user in the database.
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"email":                user.Email,
			"full_name":            user.FullName,
			"role":                 user.Role,
			"plan":                 user.Plan,
			"analysis_total_used":  user.AnalysisTotalUsed,
			"analysis_month_count": user.AnalysisMonthCount,
			"analysis_month_start": user.AnalysisMonthStart,
			"trial_expires_at":     user.TrialExpiresAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdatePlan updates only the user's plan field.
func (r *UserRepository) UpdatePlan(ctx context.Context, userID primitive.ObjectID, plan string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"plan": plan,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

