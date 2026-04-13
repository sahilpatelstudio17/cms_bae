package repositories

import (
	"cms/internal/models"

	"gorm.io/gorm"
)

type SaleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

// List returns all sales for a company with pagination
func (r *SaleRepository) List(companyID uint, limit, offset int) ([]models.Sale, int64, error) {
	var sales []models.Sale
	var total int64

	if err := r.db.Model(&models.Sale{}).Where("company_id = ?", companyID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Where("company_id = ?", companyID).
		Preload("Employee").
		Preload("Company").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sales).Error; err != nil {
		return nil, 0, err
	}

	return sales, total, nil
}

// ListByEmployee returns sales for a specific employee
func (r *SaleRepository) ListByEmployee(companyID, employeeID uint, limit, offset int) ([]models.Sale, int64, error) {
	var sales []models.Sale
	var total int64

	if err := r.db.Model(&models.Sale{}).Where("company_id = ? AND employee_id = ?", companyID, employeeID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Where("company_id = ? AND employee_id = ?", companyID, employeeID).
		Preload("Employee").
		Preload("Company").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sales).Error; err != nil {
		return nil, 0, err
	}

	return sales, total, nil
}

// ListPending returns pending sales awaiting approval
func (r *SaleRepository) ListPending(companyID uint) ([]models.Sale, error) {
	var sales []models.Sale

	if err := r.db.
		Where("company_id = ? AND status = ?", companyID, "pending").
		Preload("Employee").
		Preload("Company").
		Order("created_at DESC").
		Find(&sales).Error; err != nil {
		return nil, err
	}

	return sales, nil
}

// Create creates a new sale
func (r *SaleRepository) Create(sale *models.Sale) error {
	return r.db.Create(sale).Error
}

// Update updates an existing sale
func (r *SaleRepository) Update(companyID, saleID uint, sale *models.Sale) error {
	return r.db.Where("id = ? AND company_id = ?", saleID, companyID).Updates(sale).Error
}

// Delete deletes a sale
func (r *SaleRepository) Delete(companyID, saleID uint) error {
	return r.db.Where("id = ? AND company_id = ?", saleID, companyID).Delete(&models.Sale{}).Error
}

// GetByID gets a sale by ID
func (r *SaleRepository) GetByID(companyID, saleID uint) (*models.Sale, error) {
	var sale models.Sale

	if err := r.db.
		Where("id = ? AND company_id = ?", saleID, companyID).
		Preload("Employee").
		Preload("Company").
		First(&sale).Error; err != nil {
		return nil, err
	}

	return &sale, nil
}
