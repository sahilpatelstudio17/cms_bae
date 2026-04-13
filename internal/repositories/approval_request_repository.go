package repositories

import (
	"cms/internal/models"

	"gorm.io/gorm"
)

type ApprovalRequestRepository struct {
	db *gorm.DB
}

func NewApprovalRequestRepository(db *gorm.DB) *ApprovalRequestRepository {
	return &ApprovalRequestRepository{db: db}
}

func (r *ApprovalRequestRepository) CreateApprovalRequest(approval *models.ApprovalRequest) error {
	return r.db.Create(approval).Error
}

func (r *ApprovalRequestRepository) GetByID(id uint) (*models.ApprovalRequest, error) {
	var approval models.ApprovalRequest
	if err := r.db.Where("id = ?", id).First(&approval).Error; err != nil {
		return nil, err
	}
	return &approval, nil
}

func (r *ApprovalRequestRepository) ListPendingByCompany(companyID uint) ([]models.ApprovalRequest, error) {
	var approvals []models.ApprovalRequest
	if err := r.db.Where("company_id = ? AND status = ?", companyID, "pending").Order("created_at desc").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *ApprovalRequestRepository) UpdateApprovalRequest(approval *models.ApprovalRequest) error {
	return r.db.Save(approval).Error
}

func (r *ApprovalRequestRepository) DeleteApprovalRequest(id uint) error {
	return r.db.Where("id = ?", id).Delete(&models.ApprovalRequest{}).Error
}

func (r *ApprovalRequestRepository) GetByUserID(userID uint) (*models.ApprovalRequest, error) {
	var approval models.ApprovalRequest
	if err := r.db.Where("user_id = ? AND status = ?", userID, "pending").First(&approval).Error; err != nil {
		return nil, err
	}
	return &approval, nil
}

// Create is an alias for CreateApprovalRequest
func (r *ApprovalRequestRepository) Create(approval *models.ApprovalRequest) error {
	return r.db.Create(approval).Error
}

// Update is an alias for UpdateApprovalRequest
func (r *ApprovalRequestRepository) Update(approval *models.ApprovalRequest) error {
	return r.db.Save(approval).Error
}

// GetByCompanyAndType gets approvals by company, type, and status
func (r *ApprovalRequestRepository) GetByCompanyAndType(companyID uint, requestType, status string) ([]models.ApprovalRequest, error) {
	var approvals []models.ApprovalRequest
	query := r.db.Where("company_id = ? AND request_type = ?", companyID, requestType)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at desc").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

// GetByRequestType gets all approvals by request type and preloads user data
func (r *ApprovalRequestRepository) GetByRequestType(requestType string) ([]models.ApprovalRequest, error) {
	var approvals []models.ApprovalRequest
	if err := r.db.Where("request_type = ?", requestType).Order("created_at desc").Find(&approvals).Error; err != nil {
		return nil, err
	}

	// Manually load User data for each approval request that has a UserID
	for i := range approvals {
		if approvals[i].UserID > 0 {
			var user models.User
			if err := r.db.Where("id = ?", approvals[i].UserID).First(&user).Error; err == nil {
				approvals[i].User = &user
			}
		}
		// Manually load Employee data for role assignments
		if approvals[i].EmployeeID > 0 {
			var employee models.Employee
			if err := r.db.Where("id = ?", approvals[i].EmployeeID).First(&employee).Error; err == nil {
				approvals[i].Employee = &employee
			}
		}
	}

	return approvals, nil
}

// CheckDuplicatePendingEmail checks if an email already has a pending approval request
func (r *ApprovalRequestRepository) CheckDuplicatePendingEmail(email, requestType string) (bool, error) {
	var count int64
	if err := r.db.Where("requested_email = ? AND request_type = ? AND status = ?", email, requestType, "pending").Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
