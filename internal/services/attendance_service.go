package services

import (
	"errors"
	"time"

	"cms/internal/models"
	"cms/internal/repositories"

	"gorm.io/gorm"
)

type AttendanceService struct {
	repo         *repositories.AttendanceRepository
	employeeRepo *repositories.EmployeeRepository
	userRepo     *repositories.UserRepository
}

type CreateAttendanceRequest struct {
	EmployeeID uint   `json:"employee_id" binding:"required,gt=0"`
	Date       string `json:"date" binding:"required,datetime=2006-01-02"`
	Status     string `json:"status" binding:"required,oneof=present absent leave"`
}

func NewAttendanceService(repo *repositories.AttendanceRepository, employeeRepo *repositories.EmployeeRepository, userRepo *repositories.UserRepository) *AttendanceService {
	return &AttendanceService{repo: repo, employeeRepo: employeeRepo, userRepo: userRepo}
}

func (s *AttendanceService) List(companyID uint, limit, offset int) ([]models.Attendance, int64, error) {
	return s.repo.List(companyID, limit, offset)
}

func (s *AttendanceService) Create(companyID uint, input CreateAttendanceRequest) (*models.Attendance, error) {
	if _, err := s.employeeRepo.GetByID(companyID, input.EmployeeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found")
		}
		return nil, err
	}

	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, err
	}

	record := &models.Attendance{
		EmployeeID: input.EmployeeID,
		Date:       parsedDate,
		CompanyID:  companyID,
	}
	if err := s.repo.Create(record); err != nil {
		return nil, err
	}
	return record, nil
}

// MarkIn - Employee marks check-in time for today
func (s *AttendanceService) MarkIn(companyID, userID uint) (*models.Attendance, error) {
	// Get user to find employee
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Find employee by name and company
	employee, err := s.employeeRepo.GetByNameAndCompany(companyID, user.Name)
	if err != nil {
		return nil, errors.New("employee record not found for this user")
	}

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Check if attendance record exists for today
	existing, _ := s.repo.FindByEmployeeIDAndDate(employee.ID, todayDate.Format("2006-01-02"))
	
	var record *models.Attendance
	if existing != nil && existing.ID != 0 {
		// Record exists but already has IN time
		if existing.InTime != nil {
			return nil, errors.New("you have already checked in today")
		}
		// Update existing record with IN time
		now := time.Now()
		existing.InTime = &now
		record = existing
		if err := s.repo.Update(record); err != nil {
			return nil, err
		}
	} else {
		// Create new record with IN time
		now := time.Now()
		record = &models.Attendance{
			EmployeeID: employee.ID,
			Date:       todayDate,
			InTime:     &now,
			CompanyID:  companyID,
		}
		if err := s.repo.Create(record); err != nil {
			return nil, err
		}
	}

	return record, nil
}

// MarkOut - Employee marks check-out time for today
func (s *AttendanceService) MarkOut(companyID, userID uint) (*models.Attendance, error) {
	// Get user to find employee
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Find employee by name and company
	employee, err := s.employeeRepo.GetByNameAndCompany(companyID, user.Name)
	if err != nil {
		return nil, errors.New("employee record not found for this user")
	}

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Get today's attendance record
	existing, err := s.repo.FindByEmployeeIDAndDate(employee.ID, todayDate.Format("2006-01-02"))
	if err != nil || existing == nil || existing.ID == 0 {
		return nil, errors.New("no check-in found for today. please check in first")
	}

	// Check if already checked out
	if existing.OutTime != nil {
		return nil, errors.New("you have already checked out today")
	}

	// Update with OUT time
	now := time.Now()
	existing.OutTime = &now
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// GetMyAttendance - Employee sees their own attendance records
func (s *AttendanceService) GetMyAttendance(companyID, userID uint, limit, offset int) ([]models.Attendance, int64, error) {
	// Get user to find employee
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, 0, errors.New("user not found")
	}

	// Find employee by name and company
	employee, err := s.employeeRepo.GetByNameAndCompany(companyID, user.Name)
	if err != nil {
		return nil, 0, errors.New("employee record not found")
	}

	return s.repo.FindByEmployeeID(employee.ID, limit, offset)
}

// GetByDate - Admin sees attendance for a specific date
func (s *AttendanceService) GetByDate(companyID uint, date string, limit, offset int) ([]models.Attendance, int64, error) {
	return s.repo.FindByDateAndCompany(companyID, date, limit, offset)
}
