package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{
		auth: auth,
	}
}

type signupRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	FullName       string `json:"full_name"`
	Role           string `json:"role"`
	CompanyName    string `json:"company_name"`
	CompanyWebsite string `json:"company_website"`
	MRRRange       string `json:"mrr_range"`
	HeardFrom      string `json:"heard_from"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type companyResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Website  string `json:"website"`
	MRRRange string `json:"mrr_range,omitempty"`
}

type authUserResponse struct {
	ID        string           `json:"id"`
	Email     string           `json:"email"`
	FullName  string           `json:"full_name,omitempty"`
	Role      string           `json:"role,omitempty"`
	Plan      string           `json:"plan,omitempty"`
	CreatedAt string           `json:"created_at"`
	Company   *companyResponse `json:"company,omitempty"`
}

type signupResponse struct {
	Token   string           `json:"token"`
	User    authUserResponse `json:"user"`
	Company *companyResponse `json:"company,omitempty"`
}

type loginResponse struct {
	Token string           `json:"token"`
	User  authUserResponse `json:"user"`
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Signup handles user registration.
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSONError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	input := service.SignupInput{
		Email:          req.Email,
		Password:       req.Password,
		FullName:       req.FullName,
		Role:           req.Role,
		CompanyName:    req.CompanyName,
		CompanyWebsite: req.CompanyWebsite,
		MRRRange:       req.MRRRange,
		HeardFrom:      req.HeardFrom,
	}

	result, err := h.auth.Register(r.Context(), input)
	if err != nil {
		if err == service.ErrEmailAlreadyInUse {
			writeJSONError(w, "email already in use", http.StatusConflict)
			return
		}
		if err.Error() == "password must be at least 6 characters" {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSONError(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	// Generate token for auto-login after signup
	token, err := h.auth.GenerateTokenForUser(result.User.ID.Hex())
	if err != nil {
		writeJSONError(w, "failed to generate auth token", http.StatusInternalServerError)
		return
	}

	resp := signupResponse{
		Token: token,
		User: authUserResponse{
			ID:        result.User.ID.Hex(),
			Email:     result.User.Email,
			FullName:  result.User.FullName,
			Role:      result.User.Role,
			Plan:      result.User.Plan,
			CreatedAt: result.User.CreatedAt.Format(time.RFC3339),
		},
	}

	if result.Company != nil {
		resp.Company = &companyResponse{
			ID:       result.Company.ID.Hex(),
			Name:     result.Company.Name,
			Website:  result.Company.Website,
			MRRRange: result.Company.MRRRange,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSONError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	token, user, company, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			writeJSONError(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		writeJSONError(w, "failed to login", http.StatusInternalServerError)
		return
	}

	resp := loginResponse{
		Token: token,
		User: authUserResponse{
			ID:        user.ID.Hex(),
			Email:     user.Email,
			FullName:  user.FullName,
			Role:      user.Role,
			Plan:      user.Plan,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	}

	if company != nil {
		resp.User.Company = &companyResponse{
			ID:       company.ID.Hex(),
			Name:     company.Name,
			Website:  company.Website,
			MRRRange: company.MRRRange,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Me returns the current authenticated user's information.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, company, err := h.auth.GetUserWithCompany(r.Context(), userID)
	if err != nil {
		writeJSONError(w, "failed to fetch user", http.StatusInternalServerError)
		return
	}

	resp := authUserResponse{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		Plan:      user.Plan,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	if company != nil {
		resp.Company = &companyResponse{
			ID:       company.ID.Hex(),
			Name:     company.Name,
			Website:  company.Website,
			MRRRange: company.MRRRange,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
