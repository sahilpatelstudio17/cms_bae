package services

import (
	"cms/internal/models"
	"cms/internal/repositories"
	"cms/internal/utils"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/xuri/excelize/v2"
)

type UserImportService struct {
	userRepo     *repositories.UserRepository
	employeeRepo *repositories.EmployeeRepository
}

type UserImportData struct {
	Name   string
	Email  string
	Role   string
	Status string
}

type UserImportResult struct {
	RowNumber int
	Success   bool
	UserData  UserImportData
	ErrorMsg  string
}

func NewUserImportService(userRepo *repositories.UserRepository, employeeRepo *repositories.EmployeeRepository) *UserImportService {
	return &UserImportService{
		userRepo:     userRepo,
		employeeRepo: employeeRepo,
	}
}

func (s *UserImportService) ImportUsersFromFile(file *multipart.FileHeader, companyID uint) ([]UserImportResult, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Determine file type from filename
	filename := file.Filename
	var records [][]string

	if strings.HasSuffix(strings.ToLower(filename), ".xlsx") || strings.HasSuffix(strings.ToLower(filename), ".xls") {
		records, err = s.readExcelFile(src)
	} else if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		records, err = s.readCSVFile(src)
	} else {
		return nil, fmt.Errorf("unsupported file format. Use .xlsx, .xls, or .csv")
	}

	if err != nil {
		return nil, err
	}

	return s.processRecords(records, companyID)
}

func (s *UserImportService) readExcelFile(src io.Reader) ([][]string, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *UserImportService) readCSVFile(src io.Reader) ([][]string, error) {
	reader := csv.NewReader(src)
	var records [][]string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func (s *UserImportService) processRecords(records [][]string, companyID uint) ([]UserImportResult, error) {
	var results []UserImportResult
	validRoles := map[string]bool{
		"admin":       true,
		"employee":    true,
		"super_admin": true,
		"manager":     true,
		"salesman":    true,
		"developer":   true,
		"staff":       true,
	}

	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
		"pending":  true,
	}

	// Skip header row
	startRow := 1
	if len(records) > 0 {
		// Check if first row looks like a header
		firstRow := records[0]
		if len(firstRow) >= 4 {
			if strings.ToLower(strings.TrimSpace(firstRow[0])) == "name" ||
				strings.ToLower(strings.TrimSpace(firstRow[0])) == "employee name" {
				startRow = 1
			}
		}
	}

	for i := startRow; i < len(records); i++ {
		record := records[i]
		result := UserImportResult{RowNumber: i + 1}

		// Validate minimum columns
		if len(record) < 4 {
			result.Success = false
			result.ErrorMsg = "missing required columns (need: name, email, role, status)"
			results = append(results, result)
			continue
		}

		name := strings.TrimSpace(record[0])
		email := strings.TrimSpace(record[1])
		role := strings.ToLower(strings.TrimSpace(record[2]))
		status := strings.ToLower(strings.TrimSpace(record[3]))

		// Validate data
		if name == "" {
			result.Success = false
			result.ErrorMsg = "name is required"
			results = append(results, result)
			continue
		}

		if email == "" {
			result.Success = false
			result.ErrorMsg = "email is required"
			results = append(results, result)
			continue
		}

		if !isValidEmail(email) {
			result.Success = false
			result.ErrorMsg = "invalid email format"
			results = append(results, result)
			continue
		}

		if !validRoles[role] {
			result.Success = false
			result.ErrorMsg = fmt.Sprintf("invalid role '%s'. Valid roles: admin, employee, super_admin, manager, salesman, developer, staff", role)
			results = append(results, result)
			continue
		}

		if !validStatuses[status] {
			result.Success = false
			result.ErrorMsg = fmt.Sprintf("invalid status '%s'. Valid statuses: active, inactive, pending", status)
			results = append(results, result)
			continue
		}

		// Check if email already exists
		_, err := s.userRepo.GetUserByEmail(email)
		if err == nil {
			result.Success = false
			result.ErrorMsg = "email already exists in system"
			results = append(results, result)
			continue
		}

		// Create user
		hashedPassword, err := utils.HashPassword("User@123")
		if err != nil {
			result.Success = false
			result.ErrorMsg = "failed to create user"
			results = append(results, result)
			continue
		}

		user := &models.User{
			Name:      name,
			Email:     email,
			Password:  hashedPassword,
			Role:      role,
			Status:    status,
			CompanyID: companyID,
		}

		if err := s.userRepo.CreateUser(user); err != nil {
			result.Success = false
			result.ErrorMsg = "database error: " + err.Error()
			results = append(results, result)
			continue
		}

		// Create corresponding Employee record with position based on role
		roleToPositionMap := map[string]string{
			"salesman":    "Salesman",
			"developer":   "Developer",
			"staff":       "Staff",
			"manager":     "Manager",
			"employee":    "Employee",
			"admin":       "Admin",
			"super_admin": "Super Admin",
		}

		position := roleToPositionMap[role]
		if position == "" {
			position = role // fallback to role name
		}

		employee := &models.Employee{
			Name:      name,
			Position:  position,
			Role:      role,
			Salary:    0, // Default salary, can be updated later
			CompanyID: companyID,
			UserID:    &user.ID,
		}

		if err := s.employeeRepo.Create(employee); err != nil {
			// Log error but don't fail, user is already created
			fmt.Printf("warning: failed to create employee for user %s: %v\n", email, err)
		}

		result.Success = true
		result.UserData = UserImportData{
			Name:   name,
			Email:  email,
			Role:   role,
			Status: status,
		}
		results = append(results, result)
	}

	return results, nil
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
