package services

import (
	"errors"
	"strings"
	"time"

	"cms/internal/models"
	"cms/internal/repositories"

	"gorm.io/gorm"
)

type SalesService struct {
	repo         *repositories.SaleRepository
	userRepo     *repositories.UserRepository
	employeeRepo *repositories.EmployeeRepository
}

type CreateSaleRequest struct {
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	Product     string    `json:"product" binding:"required,min=2,max=255"`
	Customer    string    `json:"customer" binding:"required,min=2,max=255"`
	Description string    `json:"description" binding:"required,min=5"`
	SaleDate    time.Time `json:"sale_date" binding:"required"`
}

type UpsertSaleRequest struct {
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	Product     string    `json:"product" binding:"required,min=2,max=255"`
	Customer    string    `json:"customer" binding:"required,min=2,max=255"`
	Description string    `json:"description" binding:"required,min=5"`
	SaleDate    time.Time `json:"sale_date" binding:"required"`
}

func NewSalesService(repo *repositories.SaleRepository, userRepo *repositories.UserRepository, employeeRepo *repositories.EmployeeRepository) *SalesService {
	return &SalesService{repo: repo, userRepo: userRepo, employeeRepo: employeeRepo}
}

// ListSales returns all sales for a company
func (s *SalesService) ListSales(companyID uint, limit, offset int) ([]models.Sale, int64, error) {
	return s.repo.List(companyID, limit, offset)
}

// ListMySales returns sales for a specific employee
func (s *SalesService) ListMySales(companyID, employeeID uint, limit, offset int) ([]models.Sale, int64, error) {
	return s.repo.ListByEmployee(companyID, employeeID, limit, offset)
}

// ListPendingApprovals returns pending sales for admin approval
func (s *SalesService) ListPendingApprovals(companyID uint) ([]models.Sale, error) {
	return s.repo.ListPending(companyID)
}

// CreateSale creates a new sale request - userID is converted to employeeID
func (s *SalesService) CreateSale(companyID, userID uint, input CreateSaleRequest) (*models.Sale, error) {
	if input.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if strings.TrimSpace(input.Product) == "" {
		return nil, errors.New("product is required")
	}

	if strings.TrimSpace(input.Customer) == "" {
		return nil, errors.New("customer is required")
	}

	if strings.TrimSpace(input.Description) == "" {
		return nil, errors.New("description is required")
	}

	if input.SaleDate.IsZero() {
		return nil, errors.New("sale date is required")
	}

	// Find or create employee for this user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	var employee *models.Employee
	empEmployees, _, err := s.employeeRepo.List(companyID, 1000, 0)
	var emp *models.Employee
	if err == nil {
		for i, e := range empEmployees {
			if e.Name == user.Name {
				emp = &empEmployees[i]
				break
			}
		}
	}

	if emp != nil {
		employee = emp
	} else {
		// Create employee if not exists
		newEmployee := &models.Employee{
			CompanyID: companyID,
			Name:      user.Name,
			Position:  "Sales",
			Salary:    0,
		}
		if err := s.employeeRepo.Create(newEmployee); err != nil {
			return nil, errors.New("failed to create employee")
		}
		employee = newEmployee
	}

	sale := &models.Sale{
		EmployeeID:  employee.ID,
		CompanyID:   companyID,
		Amount:      input.Amount,
		Product:     strings.TrimSpace(input.Product),
		Customer:    strings.TrimSpace(input.Customer),
		Description: strings.TrimSpace(input.Description),
		SaleDate:    input.SaleDate,
		Status:      "pending",
	}

	if err := s.repo.Create(sale); err != nil {
		return nil, errors.New("failed to create sale")
	}

	// Reload with relations
	if sale, err := s.repo.GetByID(companyID, sale.ID); err == nil {
		return sale, nil
	}

	return sale, nil
}

// UpdateSale updates an existing sale
func (s *SalesService) UpdateSale(companyID, saleID uint, input UpsertSaleRequest) (*models.Sale, error) {
	if input.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if strings.TrimSpace(input.Product) == "" {
		return nil, errors.New("product is required")
	}

	if strings.TrimSpace(input.Customer) == "" {
		return nil, errors.New("customer is required")
	}

	sale, err := s.repo.GetByID(companyID, saleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("sale not found")
		}
		return nil, errors.New("failed to fetch sale")
	}

	// Only allow updating pending sales (not approved/rejected)
	if sale.Status != "pending" {
		return nil, errors.New("can only update pending sales")
	}

	updateData := &models.Sale{
		Amount:      input.Amount,
		Product:     strings.TrimSpace(input.Product),
		Customer:    strings.TrimSpace(input.Customer),
		Description: strings.TrimSpace(input.Description),
		SaleDate:    input.SaleDate,
	}

	if err := s.repo.Update(companyID, saleID, updateData); err != nil {
		return nil, errors.New("failed to update sale")
	}

	return s.repo.GetByID(companyID, saleID)
}

// ApproveSale approves a pending sale
func (s *SalesService) ApproveSale(companyID, saleID, adminID uint) (*models.Sale, error) {
	sale, err := s.repo.GetByID(companyID, saleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("sale not found")
		}
		return nil, errors.New("failed to fetch sale")
	}

	if sale.Status != "pending" {
		return nil, errors.New("sale is not pending")
	}

	now := time.Now()
	updateData := &models.Sale{
		Status:     "approved",
		ApprovedBy: &adminID,
		ApprovedAt: &now,
	}

	if err := s.repo.Update(companyID, saleID, updateData); err != nil {
		return nil, errors.New("failed to approve sale")
	}

	return s.repo.GetByID(companyID, saleID)
}

// RejectSale rejects a pending sale
func (s *SalesService) RejectSale(companyID, saleID uint) (*models.Sale, error) {
	sale, err := s.repo.GetByID(companyID, saleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("sale not found")
		}
		return nil, errors.New("failed to fetch sale")
	}

	if sale.Status != "pending" {
		return nil, errors.New("sale is not pending")
	}

	updateData := &models.Sale{
		Status: "rejected",
	}

	if err := s.repo.Update(companyID, saleID, updateData); err != nil {
		return nil, errors.New("failed to reject sale")
	}

	return s.repo.GetByID(companyID, saleID)
}

// DeleteSale deletes a sale (only if pending)
func (s *SalesService) DeleteSale(companyID, saleID uint) error {
	sale, err := s.repo.GetByID(companyID, saleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("sale not found")
		}
		return errors.New("failed to fetch sale")
	}

	if sale.Status != "pending" {
		return errors.New("can only delete pending sales")
	}

	return s.repo.Delete(companyID, saleID)
}
