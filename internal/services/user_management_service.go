package services

import (
	"errors"
	"fmt"
	"strings"

	"cms/internal/models"
	"cms/internal/repositories"
	"cms/internal/utils"

	"gorm.io/gorm"
)

type UserManagementService struct {
	userRepo     *repositories.UserRepository
	employeeRepo *repositories.EmployeeRepository
	jwtSecret    string
	jwtExpiresHr int
}

type CreateAdminRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type CreateEmployeeWithUserRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=120"`
	Position string  `json:"position" binding:"required,min=2,max=100"`
	Salary   float64 `json:"salary" binding:"required,gt=0"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8,max=72"`
}

type AdminResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CompanyID uint   `json:"company_id"`
}

type EmployeeUserResponse struct {
	Employee EmployeeDetailResponse `json:"employee"`
	User     AdminResponse          `json:"user"`
}

type EmployeeDetailResponse struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Position  string  `json:"position"`
	Salary    float64 `json:"salary"`
	CompanyID uint    `json:"company_id"`
}

func NewUserManagementService(userRepo *repositories.UserRepository, employeeRepo *repositories.EmployeeRepository, jwtSecret string, jwtExpiresHr int) *UserManagementService {
	return &UserManagementService{
		userRepo:     userRepo,
		employeeRepo: employeeRepo,
		jwtSecret:    jwtSecret,
		jwtExpiresHr: jwtExpiresHr,
	}
}

// CreateAdmin - Super Admin creates an Admin user
func (s *UserManagementService) CreateAdmin(companyID uint, req CreateAdminRequest) (*AdminResponse, error) {
	// Validate email doesn't exist
	_, err := s.userRepo.GetUserByEmail(strings.ToLower(req.Email))
	if err == nil {
		return nil, errors.New("email already in use")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	admin := &models.User{
		Name:      strings.TrimSpace(req.Name),
		Email:     strings.TrimSpace(strings.ToLower(req.Email)),
		Password:  hashedPassword,
		Role:      "admin",
		CompanyID: companyID,
	}

	if err := s.userRepo.CreateUser(admin); err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	return &AdminResponse{
		ID:        admin.ID,
		Name:      admin.Name,
		Email:     admin.Email,
		Role:      admin.Role,
		CompanyID: admin.CompanyID,
	}, nil
}

// CreateEmployeeWithUser - Admin creates an Employee and generates a User account
func (s *UserManagementService) CreateEmployeeWithUser(companyID uint, req CreateEmployeeWithUserRequest) (*EmployeeUserResponse, error) {
	// Validate email doesn't exist
	_, err := s.userRepo.GetUserByEmail(strings.ToLower(req.Email))
	if err == nil {
		return nil, errors.New("email already in use")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var employee *models.Employee
	var user *models.User

	err = s.userRepo.WithTx(func(tx *gorm.DB) error {
		// Create User account for the employee first
		usr := &models.User{
			Name:      strings.TrimSpace(req.Name),
			Email:     strings.TrimSpace(strings.ToLower(req.Email)),
			Password:  hashedPassword,
			Role:      "employee",
			CompanyID: companyID,
		}
		if err := tx.Create(usr).Error; err != nil {
			return fmt.Errorf("failed to create user account: %w", err)
		}
		user = usr

		// Create Employee record with reference to the user
		emp := &models.Employee{
			Name:      strings.TrimSpace(req.Name),
			Position:  strings.TrimSpace(req.Position),
			Salary:    req.Salary,
			CompanyID: companyID,
			UserID:    &user.ID, // Store the user ID in employee
		}
		if err := tx.Create(emp).Error; err != nil {
			return fmt.Errorf("failed to create employee: %w", err)
		}
		employee = emp

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &EmployeeUserResponse{
		Employee: EmployeeDetailResponse{
			ID:        employee.ID,
			Name:      employee.Name,
			Position:  employee.Position,
			Salary:    employee.Salary,
			CompanyID: employee.CompanyID,
		},
		User: AdminResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CompanyID: user.CompanyID,
		},
	}, nil
}

// ListAdminsByCompany - Super Admin lists all admins in the company
func (s *UserManagementService) ListAdminsByCompany(companyID uint) ([]AdminResponse, error) {
	users, err := s.userRepo.ListUsersByRole(companyID, "admin")
	if err != nil {
		return nil, err
	}

	var results []AdminResponse
	for _, u := range users {
		results = append(results, AdminResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      u.Role,
			CompanyID: u.CompanyID,
		})
	}
	return results, nil
}

// ListAllAdmins - Super Admin lists all admins across all companies
func (s *UserManagementService) ListAllAdmins() ([]AdminResponse, error) {
	users, err := s.userRepo.ListUsersByRoleGlobal("admin")
	if err != nil {
		return nil, err
	}

	var results []AdminResponse
	for _, u := range users {
		results = append(results, AdminResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      u.Role,
			CompanyID: u.CompanyID,
		})
	}
	return results, nil
}

// ListEmployeesByCompany - Admin lists all employees in the company
func (s *UserManagementService) ListEmployeesByCompany(companyID uint) ([]EmployeeDetailResponse, error) {
	employees, _, err := s.employeeRepo.List(companyID, 1000, 0)
	if err != nil {
		return nil, err
	}

	var results []EmployeeDetailResponse
	for _, emp := range employees {
		results = append(results, EmployeeDetailResponse{
			ID:        emp.ID,
			Name:      emp.Name,
			Position:  emp.Position,
			Salary:    emp.Salary,
			CompanyID: emp.CompanyID,
		})
	}
	return results, nil
}

// DeleteAdminUser - Super Admin deletes an admin user
func (s *UserManagementService) DeleteAdminUser(companyID, adminID uint) error {
	user, err := s.userRepo.GetUserByID(adminID)
	if err != nil {
		return fmt.Errorf("admin user not found")
	}

	if user.CompanyID != companyID {
		return errors.New("user does not belong to this company")
	}

	if user.Role != "admin" {
		return errors.New("user is not an admin")
	}

	if err := s.userRepo.DeleteUser(adminID); err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	return nil
}

// DeleteEmployeeWithUser - Admin deletes both employee and user account
func (s *UserManagementService) DeleteEmployeeWithUser(companyID, employeeID uint) error {
	employee, err := s.employeeRepo.GetByID(companyID, employeeID)
	if err != nil {
		return fmt.Errorf("employee not found")
	}

	if employee.CompanyID != companyID {
		return errors.New("employee does not belong to this company")
	}

	// Delete employee record
	if err := s.employeeRepo.Delete(companyID, employeeID); err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}

	return nil
}
