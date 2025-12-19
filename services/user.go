package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	auth "github.com/ODawah/Trading-Insights/authentication"
	"github.com/ODawah/Trading-Insights/models"
	"github.com/ODawah/Trading-Insights/repository"
	"gorm.io/gorm"
)

type AuthService interface {
	Signup(ctx context.Context, req *SignupRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

// DTOs (Data Transfer Objects)
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (s *authService) Signup(ctx context.Context, req *SignupRequest) (*AuthResponse, error) {
	log.Printf("🔐 Signup attempt for email: %s", req.Email)

	// Validate input
	if err := validateSignupRequest(req); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %v", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Create new user
	user := &models.User{
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Password: req.Password,
		Name:     strings.TrimSpace(req.Name),
	}

	// Hash password
	if err := user.HashPassword(user.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	log.Printf("✅ User created successfully: %s (ID: %d)", user.Email, user.ID)

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	log.Printf("🔐 Login attempt for email: %s", req.Email)

	// Validate input
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		log.Printf("⚠️  Failed login attempt for: %s (wrong password)", req.Email)
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	log.Printf("✅ User logged in successfully: %s (ID: %d)", user.Email, user.ID)

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Validation helpers
func validateSignupRequest(req *SignupRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	// Basic email validation
	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validateLoginRequest(req *LoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
