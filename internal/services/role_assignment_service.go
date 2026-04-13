package services

import (
	"cms/internal/models"
	"cms/internal/repositories"
	"cms/internal/utils"
	"errors"
	"fmt"
)

type RoleAssignmentService struct {
	approvalRepo *repositories.ApprovalRequestRepository
	employeeRepo *repositories.EmployeeRepository
	userRepo     *repositories.UserRepository
	companyRepo  *repositories.CompanyRepository
}

type RoleAssignmentRequest struct {
	EmployeeID     uint   `json:"employee_id" binding:"required"`
	RequestedRole  string `json:"requested_role" binding:"required"`
	RequestedEmail string `json:"requested_email" binding:"required"`
}

type RoleAssignmentApproval struct {
	ApprovalID uint   `json:"approval_id" binding:"required"`
	Approved   bool   `json:"approved" binding:"required"`
	Message    string `json:"message"`
}

func NewRoleAssignmentService(
	approvalRepo *repositories.ApprovalRequestRepository,
	employeeRepo *repositories.EmployeeRepository,
	userRepo *repositories.UserRepository,
	companyRepo *repositories.CompanyRepository,
) *RoleAssignmentService {
	return &RoleAssignmentService{
		approvalRepo: approvalRepo,
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
		companyRepo:  companyRepo,
	}
}

// RequestRoleAssignment creates an approval request for a role assignment
func (s *RoleAssignmentService) RequestRoleAssignment(req RoleAssignmentRequest, companyID uint, requestedByID uint) (*models.ApprovalRequest, error) {
	// Validate employee exists and belongs to company
	employee, err := s.employeeRepo.GetByID(companyID, req.EmployeeID)
	if err != nil {
		return nil, errors.New("employee not found or does not belong to your company")
	}

	// Validate email
	if req.RequestedEmail == "" {
		return nil, errors.New("email is required for user account creation")
	}

	// Check if email already exists
	_, err = s.userRepo.GetUserByEmail(req.RequestedEmail)
	if err == nil {
		return nil, errors.New("email already exists in system")
	}

	// Create approval request
	approval := &models.ApprovalRequest{
		RequestType:    "role_assignment",
		EmployeeID:     req.EmployeeID,
		Employee:       employee,
		RequestedRole:  req.RequestedRole,
		RequestedEmail: req.RequestedEmail,
		RequestedBy:    requestedByID,
		CompanyID:      companyID,
		Status:         "pending",
	}

	if err := s.approvalRepo.Create(approval); err != nil {
		return nil, err
	}

	return approval, nil
}

// ApproveRoleAssignment approves a role assignment and creates the user
func (s *RoleAssignmentService) ApproveRoleAssignment(approvalID uint, approvedByID uint) (*models.User, error) {
	// Get approval request
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return nil, errors.New("approval request not found")
	}

	if approval.RequestType != "role_assignment" {
		return nil, errors.New("this is not a role assignment request")
	}

	if approval.Status != "pending" {
		return nil, fmt.Errorf("approval request is already %s", approval.Status)
	}

	// Get employee
	employee, err := s.employeeRepo.GetByID(approval.CompanyID, approval.EmployeeID)
	if err != nil {
		return nil, errors.New("employee not found")
	}

	// Create user account with the requested role
	hashedPassword, err := utils.HashPassword("User@123") // Default temporary password
	if err != nil {
		return nil, errors.New("failed to create user account")
	}

	user := &models.User{
		Name:      employee.Name,
		Email:     approval.RequestedEmail,
		Password:  hashedPassword,
		Role:      approval.RequestedRole,
		Status:    "active",
		CompanyID: approval.CompanyID,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Update employee to link with user
	employee.UserID = &user.ID
	employee.Role = approval.RequestedRole
	if err := s.employeeRepo.Update(employee); err != nil {
		return nil, errors.New("failed to update employee with user association")
	}

	// Mark approval as approved
	approval.Status = "approved"
	approval.ApprovedBy = approvedByID
	if err := s.approvalRepo.Update(approval); err != nil {
		return nil, errors.New("failed to update approval status")
	}

	return user, nil
}

// RejectRoleAssignment rejects a role assignment request
func (s *RoleAssignmentService) RejectRoleAssignment(approvalID uint, approvedByID uint, reason string) error {
	approval, err := s.approvalRepo.GetByID(approvalID)
	if err != nil {
		return errors.New("approval request not found")
	}

	if approval.RequestType != "role_assignment" {
		return errors.New("this is not a role assignment request")
	}

	if approval.Status != "pending" {
		return fmt.Errorf("approval request is already %s", approval.Status)
	}

	approval.Status = "rejected"
	approval.ApprovedBy = approvedByID
	approval.Message = reason

	if err := s.approvalRepo.Update(approval); err != nil {
		return errors.New("failed to reject approval request")
	}

	return nil
}

// GetPendingRoleAssignments gets all pending role assignment requests for a company
func (s *RoleAssignmentService) GetPendingRoleAssignments(companyID uint) ([]models.ApprovalRequest, error) {
	return s.approvalRepo.GetByCompanyAndType(companyID, "role_assignment", "pending")
}
