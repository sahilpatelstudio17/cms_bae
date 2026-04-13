package repositories

import (
	"cms/internal/models"

	"gorm.io/gorm"
)

type ExpenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// List returns all expenses for a company
func (r *ExpenseRepository) List(companyID uint, limit, offset int) ([]models.Expense, int64, error) {
	var expenses []models.Expense
	var total int64

	base := r.db.Model(&models.Expense{}).Where("company_id = ?", companyID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Preload("Employee").Order("id desc").Limit(limit).Offset(offset).Find(&expenses).Error; err != nil {
		return nil, 0, err
	}

	return expenses, total, nil
}

// ListByEmployee returns expenses for a specific employee
func (r *ExpenseRepository) ListByEmployee(companyID, employeeID uint, limit, offset int) ([]models.Expense, int64, error) {
	var expenses []models.Expense
	var total int64

	base := r.db.Model(&models.Expense{}).Where("company_id = ? AND employee_id = ?", companyID, employeeID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Order("id desc").Limit(limit).Offset(offset).Find(&expenses).Error; err != nil {
		return nil, 0, err
	}

	return expenses, total, nil
}

// ListPending returns pending expenses for a company (admin approval list)
func (r *ExpenseRepository) ListPending(companyID uint) ([]models.Expense, error) {
	var expenses []models.Expense

	if err := r.db.Where("company_id = ? AND status = ?", companyID, "pending").
		Preload("Employee").
		Order("created_at asc").
		Find(&expenses).Error; err != nil {
		return nil, err
	}

	return expenses, nil
}

// GetByID returns a single expense with company filtering
func (r *ExpenseRepository) GetByID(companyID, expenseID uint) (*models.Expense, error) {
	var expense models.Expense
	if err := r.db.Where("id = ? AND company_id = ?", expenseID, companyID).
		Preload("Employee").
		First(&expense).Error; err != nil {
		return nil, err
	}
	return &expense, nil
}

// Create creates a new expense
func (r *ExpenseRepository) Create(expense *models.Expense) error {
	return r.db.Create(expense).Error
}

// Update updates an expense
func (r *ExpenseRepository) Update(expense *models.Expense) error {
	return r.db.Save(expense).Error
}

// Delete deletes an expense
func (r *ExpenseRepository) Delete(companyID, expenseID uint) error {
	return r.db.Where("id = ? AND company_id = ?", expenseID, companyID).Delete(&models.Expense{}).Error
}
