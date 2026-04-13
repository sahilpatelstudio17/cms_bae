package services

import (
	"errors"
	"strings"

	"cms/internal/models"
	"cms/internal/repositories"
	"cms/internal/utils"

	"gorm.io/gorm"
)

type AuthService struct {
	repo         *repositories.AuthRepository
	approvalRepo *repositories.ApprovalRequestRepository
	jwtSecret    string
	jwtExpiresHr int
}

type RegisterRequest struct {
	CompanyName  string `json:"company_name" binding:"required,min=2,max=120"`
	CompanyEmail string `json:"company_email" binding:"required,email"`
	AdminName    string `json:"admin_name" binding:"required,min=2,max=120"`
	AdminEmail   string `json:"admin_email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8,max=72"`
}

type RegisterWithApprovalRequest struct {
	CompanyName string `json:"company_name" binding:"required,min=2,max=120"`
	UserName    string `json:"user_name" binding:"required,min=2,max=120"`
	UserEmail   string `json:"user_email" binding:"required,email"`
	AdminEmail  string `json:"admin_email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func NewAuthService(repo *repositories.AuthRepository, approvalRepo *repositories.ApprovalRequestRepository, jwtSecret string, jwtExpiresHr int) *AuthService {
	return &AuthService{repo: repo, approvalRepo: approvalRepo, jwtSecret: jwtSecret, jwtExpiresHr: jwtExpiresHr}
}

func (s *AuthService) Register(input RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	companyEmail := strings.TrimSpace(strings.ToLower(input.CompanyEmail))

	// Check if email already has pending approval (optional check - don't block if fails)
	isDuplicate, _ := s.approvalRepo.CheckDuplicatePendingEmail(companyEmail, "user")
	if isDuplicate {
		return nil, errors.New("this company email already has a pending approval request")
	}

	// Create an approval request instead of immediately creating user
	approval := &models.ApprovalRequest{
		RequestType:    "user", // Company signup requests
		CompanyName:    strings.TrimSpace(input.CompanyName),
		AdminName:      strings.TrimSpace(input.AdminName),
		AdminEmail:     strings.TrimSpace(strings.ToLower(input.CompanyEmail)),
		RequestedEmail: companyEmail,
		PasswordHash:   hashedPassword, // Store hashed password for when admin approves
		Status:         "pending",
		Message:        "New company registration - waiting for admin approval",
		CompanyID:      1, // Default company for now
	}

	if err := s.approvalRepo.CreateApprovalRequest(approval); err != nil {
		return nil, errors.New("failed to create signup request: " + err.Error())
	}

	// Return response with no token (user not created yet)
	return &AuthResponse{
		Token: "",
		User: models.User{
			ID:    0,
			Name:  approval.AdminName,
			Email: approval.AdminEmail,
			Role:  "admin",
		},
	}, nil
}

func (s *AuthService) Login(input LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindUserByEmail(strings.TrimSpace(strings.ToLower(input.Email)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := utils.ComparePassword(user.Password, input.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is pending approval
	if user.Status == "pending" {
		return nil, errors.New("your account is pending approval from admin")
	}

	// Check if user is rejected
	if user.Status == "rejected" {
		return nil, errors.New("your account has been rejected")
	}

	token, err := utils.GenerateJWT(user.ID, user.CompanyID, user.Role, s.jwtSecret, s.jwtExpiresHr)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *AuthService) UpdateUserProfile(userID uint, name, email string) (*models.User, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if email is already taken by another user
	if email != user.Email {
		existingUser, err := s.repo.FindUserByEmail(strings.TrimSpace(strings.ToLower(email)))
		if err == nil && existingUser.ID != userID {
			return nil, errors.New("email already in use")
		}
	}

	user.Name = strings.TrimSpace(name)
	user.Email = strings.TrimSpace(strings.ToLower(email))

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Verify old password
	if err := utils.ComparePassword(user.Password, oldPassword); err != nil {
		return errors.New("invalid current password")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	if err := s.repo.UpdateUser(user); err != nil {
		return err
	}

	return nil
}
