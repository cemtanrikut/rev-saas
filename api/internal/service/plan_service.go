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
	// ErrPlanNotFound is returned when the plan doesn't exist or doesn't belong to the user.
	ErrPlanNotFound = errors.New("plan not found")
)

// PlanService handles business logic for plans.
type PlanService struct {
	repo *mongorepo.PlanRepository
}

// NewPlanService creates a new PlanService.
func NewPlanService(repo *mongorepo.PlanRepository) *PlanService {
	return &PlanService{
		repo: repo,
	}
}

// CreatePlan creates a new plan for a user.
func (s *PlanService) CreatePlan(ctx context.Context, userID string, name string, price float64) (*model.Plan, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("plan name is required")
	}
	if price < 0 {
		return nil, errors.New("price must be non-negative")
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	plan := &model.Plan{
		UserID: uid,
		Name:   name,
		Price:  price,
	}

	if err := s.repo.Create(ctx, plan); err != nil {
		return nil, err
	}

	return plan, nil
}

// ListPlans returns all plans for a user.
func (s *PlanService) ListPlans(ctx context.Context, userID string) ([]*model.Plan, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	return s.repo.ListByUser(ctx, uid)
}

// DeletePlan deletes a plan by ID, ensuring it belongs to the user.
func (s *PlanService) DeletePlan(ctx context.Context, userID string, planID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user id")
	}

	pid, err := primitive.ObjectIDFromHex(planID)
	if err != nil {
		return errors.New("invalid plan id")
	}

	err = s.repo.DeleteByIDAndUser(ctx, pid, uid)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrPlanNotFound
		}
		return err
	}

	return nil
}

