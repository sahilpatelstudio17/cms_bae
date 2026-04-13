package repositories

import (
	"cms/internal/models"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) List(companyID uint, limit, offset int) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	base := r.db.Model(&models.Task{}).Where("company_id = ?", companyID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Order("id desc").Limit(limit).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *TaskRepository) GetByID(companyID, taskID uint) (*models.Task, error) {
	var task models.Task
	if err := r.db.Where("id = ? AND company_id = ?", taskID, companyID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) Update(task *models.Task) error {
	return r.db.Save(task).Error
}

func (r *TaskRepository) Delete(companyID, taskID uint) error {
	return r.db.Where("id = ? AND company_id = ?", taskID, companyID).Delete(&models.Task{}).Error
}
