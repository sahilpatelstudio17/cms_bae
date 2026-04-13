package repositories

import (
	"cms/internal/models"

	"gorm.io/gorm"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) List(companyID uint, limit, offset int) ([]models.Employee, int64, error) {
	var employees []models.Employee
	var total int64

	base := r.db.Model(&models.Employee{}).Where("company_id = ?", companyID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Order("id desc").Limit(limit).Offset(offset).Find(&employees).Error; err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

func (r *EmployeeRepository) GetByID(companyID, employeeID uint) (*models.Employee, error) {
	var employee models.Employee
	if err := r.db.Where("id = ? AND company_id = ?", employeeID, companyID).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetByNameAndCompany finds an employee by name and company
func (r *EmployeeRepository) GetByNameAndCompany(companyID uint, name string) (*models.Employee, error) {
	var employee models.Employee
	if err := r.db.Where("company_id = ? AND name = ?", companyID, name).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetByUserID finds an employee by user ID and company
func (r *EmployeeRepository) GetByUserID(companyID, userID uint) (*models.Employee, error) {
	var employee models.Employee
	if err := r.db.Where("company_id = ? AND user_id = ?", companyID, userID).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *EmployeeRepository) Create(employee *models.Employee) error {
	return r.db.Create(employee).Error
}

func (r *EmployeeRepository) Update(employee *models.Employee) error {
	return r.db.Save(employee).Error
}

func (r *EmployeeRepository) Delete(companyID, employeeID uint) error {
	// Get the employee first to check if there's an associated user
	var employee models.Employee
	if err := r.db.Where("id = ? AND company_id = ?", employeeID, companyID).First(&employee).Error; err != nil {
		return err
	}

	// Delete the associated user if it exists
	if employee.UserID != nil {
		if err := r.db.Delete(&models.User{}, "id = ?", *employee.UserID).Error; err != nil {
			return err
		}
	}

	// Delete the employee
	return r.db.Where("id = ? AND company_id = ?", employeeID, companyID).Delete(&models.Employee{}).Error
}
