package repositories

import (
	"cms/internal/models"

	"gorm.io/gorm"
)

type CompanyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) GetByID(companyID uint) (*models.Company, error) {
	var company models.Company
	if err := r.db.Where("id = ?", companyID).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepository) Create(company *models.Company) error {
	return r.db.Create(company).Error
}

func (r *CompanyRepository) Update(company *models.Company) error {
	return r.db.Save(company).Error
}

func (r *CompanyRepository) GetByName(name string) (*models.Company, error) {
	var company models.Company
	if err := r.db.Where("name = ?", name).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepository) GetFirst() (*models.Company, error) {
	var company models.Company
	if err := r.db.Order("id ASC").First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}
