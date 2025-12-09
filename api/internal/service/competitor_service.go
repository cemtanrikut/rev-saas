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

// CompetitorPlanInput represents input for a competitor's plan.
type CompetitorPlanInput struct {
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	BillingCycle string  `json:"billing_cycle"`
}

// CompetitorInput represents input for creating a competitor.
type CompetitorInput struct {
	Name  string                `json:"name"`
	URL   string                `json:"url"`
	Plans []CompetitorPlanInput `json:"plans"`
}

// CreateCompetitor creates a new competitor for a user.
func (s *CompetitorService) CreateCompetitor(ctx context.Context, userID string, input CompetitorInput) (*model.Competitor, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("competitor name is required")
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	// Convert plan inputs to model plans
	plans := make([]model.CompetitorPlan, 0, len(input.Plans))
	for _, p := range input.Plans {
		planName := strings.TrimSpace(p.Name)
		if planName == "" {
			planName = "Default"
		}

		currency := strings.TrimSpace(p.Currency)
		if currency == "" {
			currency = "USD"
		}

		billingCycle := strings.TrimSpace(p.BillingCycle)
		if billingCycle == "" {
			billingCycle = "monthly"
		}

		if p.Price < 0 {
			return nil, errors.New("plan price must be non-negative")
		}

		plans = append(plans, model.CompetitorPlan{
			Name:         planName,
			Price:        p.Price,
			Currency:     currency,
			BillingCycle: billingCycle,
		})
	}

	competitor := &model.Competitor{
		UserID: uid,
		Name:   name,
		URL:    strings.TrimSpace(input.URL),
		Plans:  plans,
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

// UpdateCompetitor updates a competitor by ID, ensuring it belongs to the user.
func (s *CompetitorService) UpdateCompetitor(ctx context.Context, userID, competitorID string, input CompetitorInput) (*model.Competitor, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("competitor name is required")
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	cid, err := primitive.ObjectIDFromHex(competitorID)
	if err != nil {
		return nil, errors.New("invalid competitor id")
	}

	// Convert plan inputs to model plans
	plans := make([]model.CompetitorPlan, 0, len(input.Plans))
	for _, p := range input.Plans {
		planName := strings.TrimSpace(p.Name)
		if planName == "" {
			planName = "Default"
		}

		currency := strings.TrimSpace(p.Currency)
		if currency == "" {
			currency = "USD"
		}

		billingCycle := strings.TrimSpace(p.BillingCycle)
		if billingCycle == "" {
			billingCycle = "monthly"
		}

		if p.Price < 0 {
			return nil, errors.New("plan price must be non-negative")
		}

		plans = append(plans, model.CompetitorPlan{
			Name:         planName,
			Price:        p.Price,
			Currency:     currency,
			BillingCycle: billingCycle,
		})
	}

	competitor := &model.Competitor{
		ID:     cid,
		UserID: uid,
		Name:   name,
		URL:    strings.TrimSpace(input.URL),
		Plans:  plans,
	}

	err = s.repo.UpdateByIDAndUser(ctx, cid, uid, competitor)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCompetitorNotFound
		}
		return nil, err
	}

	return competitor, nil
}
