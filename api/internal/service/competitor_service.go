package service

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"rev-saas-api/internal/model"
	mongorepo "rev-saas-api/internal/repository/mongo"
)

var (
	// ErrCompetitorNotFound is returned when the competitor doesn't exist or doesn't belong to the user.
	ErrCompetitorNotFound = errors.New("competitor not found")
)

// CompetitorService handles business logic for competitors.
type CompetitorService struct {
	repo *mongorepo.CompetitorRepository
}

// NewCompetitorService creates a new CompetitorService.
func NewCompetitorService(repo *mongorepo.CompetitorRepository) *CompetitorService {
	return &CompetitorService{
		repo: repo,
	}
}

// CreateCompetitor creates a new competitor for a user.
func (s *CompetitorService) CreateCompetitor(ctx context.Context, userID, name, url string, basePrice float64) (*model.Competitor, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("competitor name is required")
	}

	url = strings.TrimSpace(url)

	if basePrice < 0 {
		return nil, errors.New("base price must be non-negative")
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	competitor := &model.Competitor{
		UserID:    uid,
		Name:      name,
		URL:       url,
		BasePrice: basePrice,
	}

	if err := s.repo.Create(ctx, competitor); err != nil {
		return nil, err
	}

	return competitor, nil
}

// ListCompetitors returns all competitors for a user.
func (s *CompetitorService) ListCompetitors(ctx context.Context, userID string) ([]*model.Competitor, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	return s.repo.ListByUser(ctx, uid)
}

// DeleteCompetitor deletes a competitor by ID, ensuring it belongs to the user.
func (s *CompetitorService) DeleteCompetitor(ctx context.Context, userID, competitorID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user id")
	}

	cid, err := primitive.ObjectIDFromHex(competitorID)
	if err != nil {
		return errors.New("invalid competitor id")
	}

	err = s.repo.DeleteByIDAndUser(ctx, cid, uid)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrCompetitorNotFound
		}
		return err
	}

	return nil
}

