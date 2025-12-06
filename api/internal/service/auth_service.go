package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"rev-saas-api/internal/model"
	mongorepo "rev-saas-api/internal/repository/mongo"
)

var (
	// ErrEmailAlreadyInUse is returned when the email is already registered.
	ErrEmailAlreadyInUse = errors.New("email is already in use")
	// ErrInvalidCredentials is returned when email or password is wrong.
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// SignupInput contains all the data needed to register a new user.
type SignupInput struct {
	Email          string
	Password       string
	FullName       string
	Role           string
	CompanyName    string
	CompanyWebsite string
	MRRRange       string
	HeardFrom      string
}

// SignupResult contains the results of a successful signup.
type SignupResult struct {
	User    *model.User
	Company *model.Company
}

// AuthService handles authentication logic.
type AuthService struct {
	users        *mongorepo.UserRepository
	companies    *mongorepo.CompanyRepository
	userMetadata *mongorepo.UserMetadataRepository
	jwt          *JWTService
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	users *mongorepo.UserRepository,
	companies *mongorepo.CompanyRepository,
	userMetadata *mongorepo.UserMetadataRepository,
	jwt *JWTService,
) *AuthService {
	return &AuthService{
		users:        users,
		companies:    companies,
		userMetadata: userMetadata,
		jwt:          jwt,
	}
}

// normalizeEmail lowercases and trims the email.
func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

// Register creates a new user account with company and metadata.
func (s *AuthService) Register(ctx context.Context, input SignupInput) (*SignupResult, error) {
	input.Email = normalizeEmail(input.Email)

	// Check if email is already in use
	existing, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyInUse
	}

	// Validate password length
	if len(input.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Hash the password
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	now := time.Now().UTC()
	user := &model.User{
		Email:     input.Email,
		Password:  string(hashed),
		FullName:  strings.TrimSpace(input.FullName),
		Role:      strings.TrimSpace(input.Role),
		Plan:      model.PlanFree, // Default to free plan
		CreatedAt: now,
	}

	// Set trial expiry for free plan
	trialLimits := GetPlanLimits(model.PlanFree)
	if trialLimits.TrialDays > 0 {
		user.TrialExpiresAt = now.AddDate(0, 0, trialLimits.TrialDays)
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create company if company name is provided
	var company *model.Company
	if strings.TrimSpace(input.CompanyName) != "" {
		company = &model.Company{
			UserID:   user.ID,
			Name:     strings.TrimSpace(input.CompanyName),
			Website:  strings.TrimSpace(input.CompanyWebsite),
			MRRRange: strings.TrimSpace(input.MRRRange),
		}

		if err := s.companies.Create(ctx, company); err != nil {
			// Log error but don't fail signup
			// In production, you might want to handle this differently
		}
	}

	// Create user metadata if heard_from is provided
	if strings.TrimSpace(input.HeardFrom) != "" {
		metadata := &model.UserMetadata{
			UserID:    user.ID,
			HeardFrom: strings.TrimSpace(input.HeardFrom),
		}

		if err := s.userMetadata.Create(ctx, metadata); err != nil {
			// Log error but don't fail signup
		}
	}

	// Don't return the password hash
	user.Password = ""

	return &SignupResult{
		User:    user,
		Company: company,
	}, nil
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, *model.User, *model.Company, error) {
	email = normalizeEmail(email)

	// Find user by email
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, nil, err
	}
	if user == nil {
		return "", nil, nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwt.GenerateToken(user.ID.Hex())
	if err != nil {
		return "", nil, nil, err
	}

	// Get user's company
	company, _ := s.companies.GetByUserID(ctx, user.ID)

	// Mask password before returning
	user.Password = ""

	return token, user, company, nil
}

// GetUserByID retrieves a user by their ID string.
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.users.GetByIDString(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Mask password before returning
	user.Password = ""
	return user, nil
}

// GetUserWithCompany retrieves a user and their company by user ID.
func (s *AuthService) GetUserWithCompany(ctx context.Context, userID string) (*model.User, *model.Company, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	company, _ := s.companies.GetByUserID(ctx, user.ID)

	return user, company, nil
}
