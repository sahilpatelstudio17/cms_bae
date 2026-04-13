package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cms/internal/models"
	"cms/internal/repositories"
	"cms/internal/utils"

	"gorm.io/gorm"
)

type ApprovalService struct {
	userRepo     *repositories.UserRepository
	approvalRepo *repositories.ApprovalRequestRepository
	companyRepo  *repositories.CompanyRepository
	authRepo     *repositories.AuthRepository
	employeeRepo *repositories.EmployeeRepository
	jwtSecret    string
	jwtExpiresHr int
}

type RegisterUserWithApprovalRequest struct {
	CompanyName string `json:"company_name" binding:"required,min=2,max=120"`
	UserName    string `json:"user_name" binding:"required,min=2,max=120"`
	UserEmail   string `json:"user_email" binding:"required,email"`
	AdminEmail  string `json:"admin_email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=72"`
}

type ApprovalRequestResponse struct {
	ID          uint   `json:"id"`
	RequestType string `json:"request_type"`
	UserID      uint   `json:"user_id"`
	UserName    string `json:"user_name"`
	UserEmail   string `json:"user_email"`
	AdminName   string `json:"admin_name"`
	AdminEmail  string `json:"admin_email"`
	CompanyID   uint   `json:"company_id"`
	CompanyName string `json:"company_name"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type ApprovalActionRequest struct {
	Message string `json:"message"`
}

func NewApprovalService(
	userRepo *repositories.UserRepository,
	approvalRepo *repositories.ApprovalRequestRepository,
	companyRepo *repositories.CompanyRepository,
	authRepo *repositories.AuthRepository,
	employeeRepo *repositories.EmployeeRepository,
	jwtSecret string,
	jwtExpiresHr int,
) *ApprovalService {
	return &ApprovalService{
		userRepo:     userRepo,
		approvalRepo: approvalRepo,
		companyRepo:  companyRepo,
		authRepo:     authRepo,
		employeeRepo: employeeRepo,
		jwtSecret:    jwtSecret,
		jwtExpiresHr: jwtExpiresHr,
	}
}

// RegisterUserWithApproval - Register a new user that needs admin approval
func (s *ApprovalService) RegisterUserWithApproval(req RegisterUserWithApprovalRequest) (*ApprovalRequestResponse, error) {
	// Find the company by name
	company, err := s.companyRepo.GetByName(strings.TrimSpace(req.CompanyName))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("company not found")
		}
		return nil, err
	}

	// Check if user email already exists
	_, emailErr := s.authRepo.FindUserByEmail(strings.TrimSpace(strings.ToLower(req.UserEmail)))
	if emailErr == nil {
		return nil, errors.New("email already in use")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user with pending status
	user := &models.User{
		Name:      strings.TrimSpace(req.UserName),
		Email:     strings.TrimSpace(strings.ToLower(req.UserEmail)),
		Password:  hashedPassword,
		Role:      "employee",
		Status:    "pending",
		CompanyID: company.ID,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	// Create approval request
	approval := &models.ApprovalRequest{
		UserID:    user.ID,
		CompanyID: company.ID,
		Status:    "pending",
		Message:   "Waiting for admin approval",
	}

	if err := s.approvalRepo.CreateApprovalRequest(approval); err != nil {
		return nil, err
	}

	return &ApprovalRequestResponse{
		ID:        approval.ID,
		UserID:    user.ID,
		UserName:  user.Name,
		UserEmail: user.Email,
		CompanyID: user.CompanyID,
		Status:    approval.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// ListPendingApprovals - List pending approvals for a company
func (s *ApprovalService) ListPendingApprovals(companyID uint) ([]ApprovalRequestResponse, error) {
	approvals, err := s.approvalRepo.ListPendingByCompany(companyID)
	if err != nil {
		return nil, err
	}

	var results []ApprovalRequestResponse
	for _, approval := range approvals {
		userName := approval.AdminName
		userEmail := approval.AdminEmail

		// For user/employee approvals, use the user data if available
		if approval.User != nil && approval.User.Name != "" {
			userName = approval.User.Name
		}
		if approval.User != nil && approval.User.Email != "" {
			userEmail = approval.User.Email
		}

		results = append(results, ApprovalRequestResponse{
			ID:          approval.ID,
			RequestType: approval.RequestType,
			UserID:      approval.UserID,
			UserName:    userName,
			UserEmail:   userEmail,
			AdminName:   approval.AdminName,
			AdminEmail:  approval.AdminEmail,
			CompanyID:   approval.CompanyID,
			CompanyName: approval.CompanyName,
			Status:      approval.Status,
			CreatedAt:   approval.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return results, nil
}

// ApproveUser - Admin approves a pending user
func (s *ApprovalService) ApproveUser(companyID, approvalID uint, adminID uint) error {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return errors.New("approval request not found")
	}

	if approval.CompanyID != companyID {
		return errors.New("approval request does not belong to this company")
	}

	if approval.Status != "pending" {
		return errors.New("approval request is not pending")
	}

	// Update user status to active
	user, err := s.userRepo.GetUserByID(approval.UserID)
	if err != nil {
		return err
	}

	user.Status = "active"
	if err := s.userRepo.UpdateUser(user); err != nil {
		return err
	}

	// Create an Employee record for the user so they appear in the employee dropdown
	employee := &models.Employee{
		Name:      user.Name,
		Position:  "Employee",
		Salary:    0,
		CompanyID: companyID,
	}

	if err := s.employeeRepo.Create(employee); err != nil {
		// If employee creation fails, we still want to approve the user
		// Just log it and continue
		return nil
	}

	// Update approval request
	approval.Status = "approved"
	approval.ApprovedBy = adminID
	if err := s.approvalRepo.UpdateApprovalRequest(approval); err != nil {
		return err
	}

	return nil
}

// RejectUser - Admin rejects a pending user
func (s *ApprovalService) RejectUser(companyID, approvalID uint, adminID uint, message string) error {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return errors.New("approval request not found")
	}

	if approval.CompanyID != companyID {
		return errors.New("approval request does not belong to this company")
	}

	if approval.Status != "pending" {
		return errors.New("approval request is not pending")
	}

	// Update user status to rejected
	user, err := s.userRepo.GetUserByID(approval.UserID)
	if err != nil {
		return err
	}

	user.Status = "rejected"
	if err := s.userRepo.UpdateUser(user); err != nil {
		return err
	}

	// Update approval request
	approval.Status = "rejected"
	approval.ApprovedBy = adminID
	approval.Message = message
	if err := s.approvalRepo.UpdateApprovalRequest(approval); err != nil {
		return err
	}

	return nil
}

// AdminApprovalRequest - Request to create an admin account (requires super admin approval)
type AdminApprovalRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// RequestAdminApproval - Create an admin approval request
func (s *ApprovalService) RequestAdminApproval(req AdminApprovalRequest) (*ApprovalRequestResponse, error) {
	// Check if email already exists
	_, emailErr := s.authRepo.FindUserByEmail(strings.TrimSpace(strings.ToLower(req.Email)))
	if emailErr == nil {
		return nil, errors.New("email already in use")
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// Create approval request (don't create user or company yet - wait for super admin approval)
	// CompanyID will be set when super admin approves
	approval := &models.ApprovalRequest{
		RequestType:    "admin",
		UserID:         0,  // User hasn't been created yet
		CompanyID:      0,  // Will be created when approved
		CompanyName:    "", // Will be set when approved
		Status:         "pending",
		AdminName:      strings.TrimSpace(req.Name),
		AdminEmail:     strings.TrimSpace(strings.ToLower(req.Email)),
		PasswordHash:   hashedPassword, // Store hashed password for when admin approves
		RequestedEmail: strings.TrimSpace(strings.ToLower(req.Email)),
		Message:        "Admin user approval request",
	}

	if err := s.approvalRepo.CreateApprovalRequest(approval); err != nil {
		return nil, errors.New("failed to create approval request")
	}

	return &ApprovalRequestResponse{
		ID:          approval.ID,
		UserID:      0,
		UserName:    approval.AdminName,
		UserEmail:   approval.AdminEmail,
		AdminName:   approval.AdminName,
		AdminEmail:  approval.AdminEmail,
		CompanyID:   0,
		CompanyName: "Pending - will be created on approval",
		Status:      "pending",
		CreatedAt:   approval.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// ApproveAdminRequest - Super admin approves admin user creation
func (s *ApprovalService) ApproveAdminRequest(approvalID uint, adminID uint) (*models.User, error) {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return nil, errors.New("approval request not found")
	}

	if approval.RequestType != "admin" {
		return nil, errors.New("this approval request is not for admin creation")
	}

	if approval.Status != "pending" {
		return nil, errors.New("approval request is not pending")
	}

	var user *models.User
	companyID := approval.CompanyID

	// If UserID is 0, create new admin user from approval data
	if approval.UserID == 0 {
		// Create a new company for this admin with isolated data
		// Use unique email to avoid database constraint violations
		uniqueEmail := fmt.Sprintf("company_%d_%d@system.local", approval.ID, time.Now().Unix())
		newCompany := &models.Company{
			Name:  approval.AdminName + "'s Company",
			Email: uniqueEmail,
		}
		if err := s.companyRepo.Create(newCompany); err != nil {
			return nil, errors.New("failed to create company for admin: " + err.Error())
		}
		companyID = newCompany.ID

		newUser := &models.User{
			Name:      approval.AdminName,
			Email:     approval.AdminEmail,
			Password:  approval.PasswordHash, // Use hashed password from approval
			Role:      "admin",
			Status:    "active",
			CompanyID: companyID,
		}
		if err := s.userRepo.CreateUser(newUser); err != nil {
			return nil, errors.New("failed to create admin user: " + err.Error())
		}
		user = newUser
		approval.UserID = user.ID
		approval.CompanyID = companyID
	} else {
		// If UserID > 0, update existing user
		existingUser, err := s.userRepo.GetUserByID(approval.UserID)
		if err != nil {
			return nil, errors.New("user not found")
		}
		existingUser.Status = "active"
		existingUser.Role = "admin"
		if err := s.userRepo.UpdateUser(existingUser); err != nil {
			return nil, err
		}
		user = existingUser
	}

	// Mark approval as approved
	approval.Status = "approved"
	approval.ApprovedBy = adminID
	if err := s.approvalRepo.UpdateApprovalRequest(approval); err != nil {
		return nil, err
	}

	return user, nil
}

// RejectAdminRequest - Super admin rejects admin user creation
func (s *ApprovalService) RejectAdminRequest(approvalID uint, adminID uint, reason string) error {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return errors.New("approval request not found")
	}

	if approval.RequestType != "admin" {
		return errors.New("this approval request is not for admin creation")
	}

	if approval.Status != "pending" {
		return errors.New("approval request is not pending")
	}

	// Delete the temporary user
	user, err := s.userRepo.GetUserByID(approval.UserID)
	if err == nil {
		s.userRepo.DeleteUser(user.ID)
	}

	// Mark approval as rejected
	approval.Status = "rejected"
	approval.ApprovedBy = adminID
	approval.Message = reason
	if err := s.approvalRepo.UpdateApprovalRequest(approval); err != nil {
		return err
	}

	return nil
}

// GetPendingAdminRequests - Get all pending admin approval requests
func (s *ApprovalService) GetPendingAdminRequests() ([]ApprovalRequestResponse, error) {
	approvals, err := s.approvalRepo.GetByRequestType("admin")
	if err != nil {
		return nil, err
	}

	var responses []ApprovalRequestResponse
	for _, approval := range approvals {
		if approval.Status == "pending" {
			responses = append(responses, ApprovalRequestResponse{
				ID:          approval.ID,
				UserID:      approval.UserID,
				UserName:    approval.AdminName,
				UserEmail:   approval.AdminEmail,
				AdminName:   approval.AdminName,
				AdminEmail:  approval.AdminEmail,
				CompanyID:   approval.CompanyID,
				CompanyName: approval.CompanyName,
				Status:      approval.Status,
				CreatedAt:   approval.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return responses, nil
}

// ApproveCompanySignup - Approve a company signup request and create company, user, subscription
func (s *ApprovalService) ApproveCompanySignup(approvalID uint, adminID uint) (*AuthResponse, error) {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return nil, errors.New("approval request not found")
	}

	if approval.RequestType != "user" {
		return nil, errors.New("invalid approval type - must be user signup")
	}

	if approval.Status != "pending" {
		return nil, errors.New("approval request is not pending")
	}

	// Create the company
	company := &models.Company{
		Name:  approval.CompanyName,
		Email: approval.RequestedEmail,
	}

	// Create subscription
	subscription := &models.Subscription{
		Plan:   "starter",
		Status: "active",
	}

	// Create admin user with stored password
	user := &models.User{
		Name:     approval.AdminName,
		Email:    approval.AdminEmail,
		Password: approval.PasswordHash, // Use stored hashed password
		Role:     "admin",
		Status:   "active",
	}

	// Use transaction to create all together
	if err := s.authRepo.CreateCompanyWithAdmin(company, user, subscription); err != nil {
		return nil, errors.New("failed to create company and user")
	}

	// Update approval request
	approval.Status = "approved"
	approval.ApprovedBy = adminID
	approval.UserID = user.ID // Store created user ID
	approval.Message = "Company signup approved and account created"
	if err := s.approvalRepo.UpdateApprovalRequest(approval); err != nil {
		return nil, err
	}

	// Generate JWT token
	jwtSecret := s.jwtSecret // Use the secret from ApprovalService
	token, err := utils.GenerateJWT(user.ID, user.CompanyID, user.Role, jwtSecret, s.jwtExpiresHr)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// GetPendingCompanySignups - Get all pending company signup requests
func (s *ApprovalService) GetPendingCompanySignups() ([]ApprovalRequestResponse, error) {
	approvals, err := s.approvalRepo.GetByRequestType("user")
	if err != nil {
		return nil, err
	}

	var responses []ApprovalRequestResponse
	for _, approval := range approvals {
		if approval.Status == "pending" {
			responses = append(responses, ApprovalRequestResponse{
				ID:          approval.ID,
				UserID:      approval.UserID,
				UserName:    approval.AdminName,
				UserEmail:   approval.AdminEmail,
				AdminName:   approval.AdminName,
				AdminEmail:  approval.AdminEmail,
				CompanyID:   approval.CompanyID,
				CompanyName: approval.CompanyName,
				Status:      approval.Status,
				CreatedAt:   approval.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return responses, nil
}

// EmployeeApprovalRequest - Request to add a new employee (requires admin approval)
type EmployeeApprovalRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=120"`
	Position string  `json:"position" binding:"required,min=2,max=120"`
	Salary   float64 `json:"salary" binding:"required,gt=0"`
}

type EmployeeApprovalResponse struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Position  string  `json:"position"`
	Salary    float64 `json:"salary"`
	CompanyID uint    `json:"company_id"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

// RequestEmployeeApproval - Create an employee approval request
func (s *ApprovalService) RequestEmployeeApproval(companyID uint, req EmployeeApprovalRequest) (*EmployeeApprovalResponse, error) {
	// Create approval request with employee details
	approval := &models.ApprovalRequest{
		RequestType:   "employee",
		CompanyID:     companyID,
		Status:        "pending",
		AdminName:     strings.TrimSpace(req.Name),
		AdminEmail:    "",                              // Will be filled when approved
		RequestedData: strings.TrimSpace(req.Position), // Store position temporarily
		Message:       "Employee request",
		Salary:        req.Salary,
	}

	if err := s.approvalRepo.CreateApprovalRequest(approval); err != nil {
		return nil, errors.New("failed to create employee approval request")
	}

	return &EmployeeApprovalResponse{
		ID:        approval.ID,
		Name:      approval.AdminName,
		Position:  approval.RequestedData,
		Salary:    approval.Salary,
		CompanyID: companyID,
		Status:    "pending",
		CreatedAt: approval.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetPendingEmployeeRequests - Get all pending employee approval requests for a company
func (s *ApprovalService) GetPendingEmployeeRequests(companyID uint) ([]EmployeeApprovalResponse, error) {
	approvals, err := s.approvalRepo.GetByRequestType("employee")
	if err != nil {
		return nil, err
	}

	var responses []EmployeeApprovalResponse
	for _, approval := range approvals {
		if approval.Status == "pending" && approval.CompanyID == companyID {
			responses = append(responses, EmployeeApprovalResponse{
				ID:        approval.ID,
				Name:      approval.AdminName,
				Position:  approval.RequestedData,
				Salary:    approval.Salary,
				CompanyID: approval.CompanyID,
				Status:    approval.Status,
				CreatedAt: approval.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return responses, nil
}

// ApproveEmployeeRequest - Admin approves employee request and creates employee + user
func (s *ApprovalService) ApproveEmployeeRequest(approvalID uint, adminID uint) (*models.Employee, error) {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return nil, errors.New("approval request not found")
	}

	if approval.RequestType != "employee" {
		return nil, errors.New("this approval request is not for employee creation")
	}

	if approval.Status != "pending" {
		return nil, errors.New("approval request is not pending")
	}

	// Create employee record
	employee := &models.Employee{
		Name:      approval.AdminName,
		Position:  approval.RequestedData,
		Salary:    approval.Salary,
		CompanyID: approval.CompanyID,
	}

	if err := s.employeeRepo.Create(employee); err != nil {
		return nil, errors.New("failed to create employee: " + err.Error())
	}

	// Mark approval as approved
	approval.Status = "approved"
	approval.ApprovedBy = adminID
	approval.UserID = employee.ID // Store employee ID for reference
	if err := s.approvalRepo.UpdateApprovalRequest(approval); err != nil {
		return nil, err
	}

	return employee, nil
}

// RejectEmployeeRequest - Admin rejects employee request
func (s *ApprovalService) RejectEmployeeRequest(approvalID uint, adminID uint, reason string) error {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return errors.New("approval request not found")
	}

	if approval.RequestType != "employee" {
		return errors.New("this approval request is not for employee creation")
	}

	if approval.Status != "pending" {
		return errors.New("approval request is not pending")
	}

	// Mark approval as rejected
	approval.Status = "rejected"
	approval.ApprovedBy = adminID
	approval.Message = reason
	if err := s.approvalRepo.UpdateApprovalRequest(approval); err != nil {
		return err
	}

	return nil
}

// UserApprovalRequest - Request to create a new user (requires admin approval)
type UserApprovalRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Role     string `json:"role" binding:"required"`
	Status   string `json:"status" binding:"required"`
}

type UserApprovalResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CompanyID uint   `json:"company_id"`
	CreatedAt string `json:"created_at"`
}

// RequestUserApproval - Create a user approval request
func (s *ApprovalService) RequestUserApproval(companyID uint, req UserApprovalRequest) (*UserApprovalResponse, error) {
	// Check if email already exists
	_, emailErr := s.authRepo.FindUserByEmail(strings.TrimSpace(strings.ToLower(req.Email)))
	if emailErr == nil {
		return nil, errors.New("email already in use")
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// Create approval request with user details
	approval := &models.ApprovalRequest{
		RequestType:     "user",
		CompanyID:       companyID,
		Status:          "pending",
		AdminName:       strings.TrimSpace(req.Name),
		AdminEmail:      strings.TrimSpace(strings.ToLower(req.Email)),
		PasswordHash:    hashedPassword,
		RequestedRole:   req.Role,
		RequestedStatus: req.Status,
		RequestedEmail:  strings.TrimSpace(strings.ToLower(req.Email)),
		Message:         "User approval request",
	}

	if err := s.approvalRepo.CreateApprovalRequest(approval); err != nil {
		return nil, errors.New("failed to create user approval request")
	}

	return &UserApprovalResponse{
		ID:        approval.ID,
		Name:      approval.AdminName,
		Email:     approval.AdminEmail,
		Role:      approval.RequestedRole,
		Status:    approval.Status,
		CompanyID: companyID,
		CreatedAt: approval.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetPendingUserApprovals - Get all pending user approval requests for a company
func (s *ApprovalService) GetPendingUserApprovals(companyID uint) ([]UserApprovalResponse, error) {
	approvals, err := s.approvalRepo.GetByRequestType("user")
	if err != nil {
		return nil, err
	}

	var responses []UserApprovalResponse
	for _, approval := range approvals {
		if approval.Status == "pending" && approval.CompanyID == companyID {
			responses = append(responses, UserApprovalResponse{
				ID:        approval.ID,
				Name:      approval.AdminName,
				Email:     approval.AdminEmail,
				Role:      approval.RequestedRole,
				Status:    approval.Status,
				CompanyID: approval.CompanyID,
				CreatedAt: approval.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return responses, nil
}

// ApproveUserRequest - Admin approves user request and creates user
func (s *ApprovalService) ApproveUserRequest(approvalID uint, adminID uint) (*models.User, error) {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return nil, errors.New("approval request not found")
	}

	if approval.RequestType != "user" {
		return nil, errors.New("this approval request is not for user creation")
	}

	if approval.Status != "pending" {
		return nil, errors.New("approval request is not pending")
	}

	// Create user record with requested status (or 'active' if not specified)
	userStatus := approval.RequestedStatus
	if userStatus == "" {
		userStatus = "active"
	}
	user := &models.User{
		Name:      approval.AdminName,
		Email:     approval.AdminEmail,
		Password:  approval.PasswordHash,
		Role:      approval.RequestedRole,
		Status:    userStatus,
		CompanyID: approval.CompanyID,
		CreatedBy: adminID,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, errors.New("failed to create user: " + err.Error())
	}

	// Auto-create Employee record linked to user
	employee := &models.Employee{
		Name:      user.Name,
		Position:  "Employee",
		Role:      approval.RequestedRole,
		Salary:    0,
		CompanyID: approval.CompanyID,
		UserID:    &user.ID,
	}
	if err := s.employeeRepo.Create(employee); err != nil {
		// Continue anyway, user is created even if employee creation fails
	}

	// Mark approval as approved
	approval.Status = "approved"
	approval.ApprovedBy = adminID
	approval.UserID = user.ID
	if err := s.approvalRepo.UpdateApprovalRequest(approval); err != nil {
		return nil, err
	}

	return user, nil
}

// RejectUserRequest - Admin rejects user request
func (s *ApprovalService) RejectUserRequest(approvalID uint, adminID uint, reason string) error {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return errors.New("approval request not found")
	}

	if approval.RequestType != "user" {
		return errors.New("this approval request is not for user creation")
	}

	if approval.Status != "pending" {
		return errors.New("approval request is not pending")
	}

	// Mark approval as rejected
	approval.Status = "rejected"
	approval.ApprovedBy = adminID
	approval.Message = reason
	if err := s.approvalRepo.UpdateApprovalRequest(approval); err != nil {
		return err
	}

	return nil
}
