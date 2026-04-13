package services

import (
	"errors"
	"strings"
	"time"

	"cms/internal/models"
	"cms/internal/repositories"

	"gorm.io/gorm"
)

type ExpenseService struct {
	repo         *repositories.ExpenseRepository
	userRepo     *repositories.UserRepository
	employeeRepo *repositories.EmployeeRepository
}

type CreateExpenseRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Category    string  `json:"category" binding:"required,min=2,max=50"`
	Description string  `json:"description" binding:"required,min=5"`
}

type UpsertExpenseRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Category    string  `json:"category" binding:"required,min=2,max=50"`
	Description string  `json:"description" binding:"required,min=5"`
}

func NewExpenseService(repo *repositories.ExpenseRepository, userRepo *repositories.UserRepository, employeeRepo *repositories.EmployeeRepository) *ExpenseService {
	return &ExpenseService{repo: repo, userRepo: userRepo, employeeRepo: employeeRepo}
}

// ListExpenses returns all expenses for a company
func (s *ExpenseService) ListExpenses(companyID uint, limit, offset int) ([]models.Expense, int64, error) {
	return s.repo.List(companyID, limit, offset)
}

// ListMyExpenses returns expenses for a specific employee
func (s *ExpenseService) ListMyExpenses(companyID, employeeID uint, limit, offset int) ([]models.Expense, int64, error) {
	return s.repo.ListByEmployee(companyID, employeeID, limit, offset)
}

// ListPendingApprovals returns pending expenses for admin approval
func (s *ExpenseService) ListPendingApprovals(companyID uint) ([]models.Expense, error) {
	return s.repo.ListPending(companyID)
}

// CreateExpense creates a new expense request - userID is converted to employeeID
func (s *ExpenseService) CreateExpense(companyID, userID uint, input CreateExpenseRequest) (*models.Expense, error) {
	if input.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if strings.TrimSpace(input.Category) == "" {
		return nil, errors.New("category is required")
	}

	if strings.TrimSpace(input.Description) == "" {
		return nil, errors.New("description is required")
	}

	// Get user to find their name for employee lookup
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Try to find existing employee by name and company
	var employeeID uint
	existingEmployee, findErr := s.employeeRepo.GetByNameAndCompany(companyID, user.Name)

	if findErr == nil && existingEmployee != nil {
		// Employee exists, use their ID
		employeeID = existingEmployee.ID
	} else {
		// Employee doesn't exist, create one
		newEmployee := &models.Employee{
			Name:      user.Name,
			Position:  "Employee",
			Salary:    0,
			CompanyID: companyID,
		}

		if err := s.employeeRepo.Create(newEmployee); err != nil {
			return nil, errors.New("failed to create employee record")
		}
		employeeID = newEmployee.ID
	}

	expense := &models.Expense{
		EmployeeID:  employeeID,
		CompanyID:   companyID,
		Amount:      input.Amount,
		Category:    strings.TrimSpace(input.Category),
		Description: strings.TrimSpace(input.Description),
		Status:      "pending",
	}

	if err := s.repo.Create(expense); err != nil {
		return nil, err
	}

	return expense, nil
}

// UpdateExpense updates an existing expense (only if pending)
func (s *ExpenseService) UpdateExpense(companyID, expenseID uint, input UpsertExpenseRequest) (*models.Expense, error) {
	expense, err := s.repo.GetByID(companyID, expenseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense not found")
		}
		return nil, err
	}

	if expense.Status != "pending" {
		return nil, errors.New("can only edit pending expenses")
	}

	expense.Amount = input.Amount
	expense.Category = strings.TrimSpace(input.Category)
	expense.Description = strings.TrimSpace(input.Description)

	if err := s.repo.Update(expense); err != nil {
		return nil, err
	}

	return expense, nil
}

// ApproveExpense approves a pending expense
func (s *ExpenseService) ApproveExpense(companyID, expenseID, adminID uint) (*models.Expense, error) {
	expense, err := s.repo.GetByID(companyID, expenseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense not found")
		}
		return nil, err
	}

	if expense.Status != "pending" {
		return nil, errors.New("can only approve pending expenses")
	}

	now := time.Now()
	expense.Status = "approved"
	expense.ApprovedBy = &adminID
	expense.ApprovedAt = &now

	if err := s.repo.Update(expense); err != nil {
		return nil, err
	}

	return expense, nil
}

// RejectExpense rejects a pending expense
func (s *ExpenseService) RejectExpense(companyID, expenseID, adminID uint) (*models.Expense, error) {
	expense, err := s.repo.GetByID(companyID, expenseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense not found")
		}
		return nil, err
	}

	if expense.Status != "pending" {
		return nil, errors.New("can only reject pending expenses")
	}

	now := time.Now()
	expense.Status = "rejected"
	expense.ApprovedBy = &adminID
	expense.ApprovedAt = &now

	if err := s.repo.Update(expense); err != nil {
		return nil, err
	}

	return expense, nil
}
