package services

import (
	"fmt"
	"mime/multipart"

	"cms/internal/models"
	"cms/internal/repositories"
	"cms/internal/utils"

	"github.com/xuri/excelize/v2"
)

type BulkImportService struct {
	userRepo     *repositories.UserRepository
	employeeRepo *repositories.EmployeeRepository
}

func NewBulkImportService(
	userRepo *repositories.UserRepository,
	employeeRepo *repositories.EmployeeRepository,
) *BulkImportService {
	return &BulkImportService{
		userRepo:     userRepo,
		employeeRepo: employeeRepo,
	}
}

type BulkImportResult struct {
	TotalRows      int                       `json:"total_rows"`
	SuccessCount   int                       `json:"success_count"`
	FailCount      int                       `json:"fail_count"`
	SuccessDetails []BulkImportSuccessDetail `json:"success_details"`
	Errors         []BulkImportError         `json:"errors"`
}

type BulkImportSuccessDetail struct {
	RowNumber int    `json:"row_number"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Password  string `json:"password"`
}

type BulkImportError struct {
	RowNumber int    `json:"row_number"`
	Error     string `json:"error"`
}

type EmployeeImportData struct {
	Name       string
	Email      string
	EmployeeID string
	Password   string
}

func (s *BulkImportService) ImportEmployeesFromExcel(file *multipart.FileHeader, companyID uint) (*BulkImportResult, error) {
	result := &BulkImportResult{
		SuccessDetails: []BulkImportSuccessDetail{},
		Errors:         []BulkImportError{},
	}

	src, err := file.Open()
	if err != nil {
		result.Errors = append(result.Errors, BulkImportError{
			RowNumber: 0,
			Error:     fmt.Sprintf("Failed to open file: %v", err),
		})
		return result, err
	}
	defer src.Close()

	xlsx, err := excelize.OpenReader(src)
	if err != nil {
		result.Errors = append(result.Errors, BulkImportError{
			RowNumber: 0,
			Error:     fmt.Sprintf("Failed to parse Excel file: %v", err),
		})
		return result, err
	}
	defer xlsx.Close()

	rows, err := xlsx.GetRows(xlsx.GetSheetName(0))
	if err != nil {
		result.Errors = append(result.Errors, BulkImportError{
			RowNumber: 0,
			Error:     fmt.Sprintf("Failed to read rows: %v", err),
		})
		return result, err
	}

	if len(rows) < 2 {
		result.Errors = append(result.Errors, BulkImportError{
			RowNumber: 0,
			Error:     "Excel file must have header row and at least one data row",
		})
		return result, nil
	}

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 3 {
			result.Errors = append(result.Errors, BulkImportError{
				RowNumber: i + 1,
				Error:     "Row must have at least 3 columns: Name, Email, Employee ID",
			})
			result.FailCount++
			continue
		}

		importData := EmployeeImportData{
			Name:       row[0],
			Email:      row[1],
			EmployeeID: row[2],
		}

		if len(row) > 3 {
			importData.Password = row[3]
		}

		if importData.Name == "" || importData.Email == "" || importData.EmployeeID == "" {
			result.Errors = append(result.Errors, BulkImportError{
				RowNumber: i + 1,
				Error:     "Name, Email, and Employee ID are required",
			})
			result.FailCount++
			continue
		}

		existingUser, _ := s.userRepo.GetUserByEmail(importData.Email)
		if existingUser != nil {
			result.Errors = append(result.Errors, BulkImportError{
				RowNumber: i + 1,
				Error:     fmt.Sprintf("User with email %s already exists", importData.Email),
			})
			result.FailCount++
			continue
		}

		finalPassword := importData.Password
		if finalPassword == "" {
			finalPassword = "User@123"
		}

		hashedPassword, err := utils.HashPassword(finalPassword)
		if err != nil {
			result.Errors = append(result.Errors, BulkImportError{
				RowNumber: i + 1,
				Error:     fmt.Sprintf("Failed to hash password: %v", err),
			})
			result.FailCount++
			continue
		}

		user := &models.User{
			Name:      importData.Name,
			Email:     importData.Email,
			Password:  hashedPassword,
			Role:      "employee",
			Status:    "active",
			CompanyID: companyID,
		}

		if err := s.userRepo.CreateUser(user); err != nil {
			result.Errors = append(result.Errors, BulkImportError{
				RowNumber: i + 1,
				Error:     fmt.Sprintf("Failed to create user: %v", err),
			})
			result.FailCount++
			continue
		}

		employee := &models.Employee{
			Name:      importData.Name,
			Position:  "Employee",
			Role:      "employee",
			Salary:    0,
			CompanyID: companyID,
		}

		if err := s.employeeRepo.Create(employee); err != nil {
			s.userRepo.DeleteUser(user.ID)
			result.Errors = append(result.Errors, BulkImportError{
				RowNumber: i + 1,
				Error:     fmt.Sprintf("Failed to create employee: %v", err),
			})
			result.FailCount++
			continue
		}

		result.SuccessCount++
		result.SuccessDetails = append(result.SuccessDetails, BulkImportSuccessDetail{
			RowNumber: i + 1,
			Email:     importData.Email,
			Name:      importData.Name,
			Password:  finalPassword,
		})
	}

	result.TotalRows = len(rows) - 1
	return result, nil
}
