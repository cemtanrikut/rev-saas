package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"rev-saas-api/internal/model"
)

// CompanyRepository handles company data operations in MongoDB.
type CompanyRepository struct {
	collection *mongo.Collection
}

// NewCompanyRepository creates a new CompanyRepository.
func NewCompanyRepository(db *mongo.Database) *CompanyRepository {
	return &CompanyRepository{
		collection: db.Collection("companies"),
	}
}

// Create inserts a new company into the database.
func (r *CompanyRepository) Create(ctx context.Context, company *model.Company) error {
	company.CreatedAt = time.Now().UTC()
	result, err := r.collection.InsertOne(ctx, company)
	if err != nil {
		return err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		company.ID = oid
	}
	return nil
}

// GetByUserID retrieves the company for a specific user.
func (r *CompanyRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) (*model.Company, error) {
	var company model.Company
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&company)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// GetByID retrieves a company by its ID.
func (r *CompanyRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Company, error) {
	var company model.Company
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&company)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// Update updates a company's information.
func (r *CompanyRepository) Update(ctx context.Context, company *model.Company) error {
	filter := bson.M{"_id": company.ID}
	update := bson.M{
		"$set": bson.M{
			"name":      company.Name,
			"website":   company.Website,
			"mrr_range": company.MRRRange,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteByUserID deletes a company by user ID.
func (r *CompanyRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}



