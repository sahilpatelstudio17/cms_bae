package repositories

import (
	"cms/internal/models"

	"gorm.io/gorm"
)

type AttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) List(companyID uint, limit, offset int) ([]models.Attendance, int64, error) {
	var records []models.Attendance
	var total int64

	base := r.db.Model(&models.Attendance{}).Where("company_id = ?", companyID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Order("date desc").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *AttendanceRepository) Create(record *models.Attendance) error {
	return r.db.Create(record).Error
}

func (r *AttendanceRepository) Update(record *models.Attendance) error {
	return r.db.Save(record).Error
}

// FindByEmployeeIDAndDate checks if attendance is already marked
func (r *AttendanceRepository) FindByEmployeeIDAndDate(employeeID uint, date string) (*models.Attendance, error) {
	var record models.Attendance
	if err := r.db.Where("employee_id = ? AND DATE(date) = ?", employeeID, date).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByEmployeeID gets all attendance records for an employee
func (r *AttendanceRepository) FindByEmployeeID(employeeID uint, limit, offset int) ([]models.Attendance, int64, error) {
	var records []models.Attendance
	var total int64

	base := r.db.Model(&models.Attendance{}).Where("employee_id = ?", employeeID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Order("date desc").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// FindByDateAndCompany gets all attendance for a specific date in a company
func (r *AttendanceRepository) FindByDateAndCompany(companyID uint, date string, limit, offset int) ([]models.Attendance, int64, error) {
	var records []models.Attendance
	var total int64

	base := r.db.Model(&models.Attendance{}).
		Where("company_id = ? AND DATE(date) = ?", companyID, date).
		Preload("Employee")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Order("employee_id asc").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
