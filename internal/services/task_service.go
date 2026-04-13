package services

import (
	"errors"
	"strings"

	"cms/internal/models"
	"cms/internal/repositories"

	"gorm.io/gorm"
)

type TaskService struct {
	repo         *repositories.TaskRepository
	employeeRepo *repositories.EmployeeRepository
}

type UpsertTaskRequest struct {
	Title       string `json:"title" binding:"required,min=2,max=150"`
	Description string `json:"description" binding:"max=3000"`
	Status      string `json:"status" binding:"required,oneof=pending in_progress completed"`
	AssignedTo  uint   `json:"assigned_to" binding:"required,gt=0"`
}

func NewTaskService(repo *repositories.TaskRepository, employeeRepo *repositories.EmployeeRepository) *TaskService {
	return &TaskService{repo: repo, employeeRepo: employeeRepo}
}

func (s *TaskService) List(companyID uint, limit, offset int) ([]models.Task, int64, error) {
	return s.repo.List(companyID, limit, offset)
}

func (s *TaskService) Create(companyID uint, input UpsertTaskRequest) (*models.Task, error) {
	if _, err := s.employeeRepo.GetByID(companyID, input.AssignedTo); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("assigned employee not found")
		}
		return nil, err
	}

	task := &models.Task{
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Status:      input.Status,
		AssignedTo:  input.AssignedTo,
		CompanyID:   companyID,
	}

	if err := s.repo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Update(companyID, taskID uint, input UpsertTaskRequest) (*models.Task, error) {
	if _, err := s.employeeRepo.GetByID(companyID, input.AssignedTo); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("assigned employee not found")
		}
		return nil, err
	}

	task, err := s.repo.GetByID(companyID, taskID)
	if err != nil {
		return nil, err
	}
	task.Title = strings.TrimSpace(input.Title)
	task.Description = strings.TrimSpace(input.Description)
	task.Status = input.Status
	task.AssignedTo = input.AssignedTo

	if err := s.repo.Update(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) GetAssignedEmployeeUserID(companyID, employeeID uint) (*uint, error) {
	employee, err := s.employeeRepo.GetByID(companyID, employeeID)
	if err != nil {
		return nil, err
	}
	return employee.UserID, nil
}

func (s *TaskService) Delete(companyID, taskID uint) error {
	return s.repo.Delete(companyID, taskID)
}
