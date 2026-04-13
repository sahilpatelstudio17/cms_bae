package services

import (
	"strings"

	"cms/internal/models"
	"cms/internal/repositories"
)

type EmployeeService struct {
	repo *repositories.EmployeeRepository
}

type UpsertEmployeeRequest struct {
	Name     string  `json:"name" binding:"required,min=2,max=120"`
	Position string  `json:"position" binding:"required,min=2,max=100"`
	Salary   float64 `json:"salary" binding:"required,gt=0"`
}

func NewEmployeeService(repo *repositories.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) List(companyID uint, limit, offset int) ([]models.Employee, int64, error) {
	return s.repo.List(companyID, limit, offset)
}

func (s *EmployeeService) Create(companyID uint, input UpsertEmployeeRequest) (*models.Employee, error) {
	employee := &models.Employee{
		Name:      strings.TrimSpace(input.Name),
		Position:  strings.TrimSpace(input.Position),
		Salary:    input.Salary,
		CompanyID: companyID,
	}
	if err := s.repo.Create(employee); err != nil {
		return nil, err
	}
	return employee, nil
}

func (s *EmployeeService) Update(companyID, employeeID uint, input UpsertEmployeeRequest) (*models.Employee, error) {
	employee, err := s.repo.GetByID(companyID, employeeID)
	if err != nil {
		return nil, err
	}

	employee.Name = strings.TrimSpace(input.Name)
	employee.Position = strings.TrimSpace(input.Position)
	employee.Salary = input.Salary

	if err := s.repo.Update(employee); err != nil {
		return nil, err
	}
	return employee, nil
}

func (s *EmployeeService) Delete(companyID, employeeID uint) error {
	return s.repo.Delete(companyID, employeeID)
}
