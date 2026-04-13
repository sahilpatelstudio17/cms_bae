package repositories

import (
	"cms/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) DeleteUser(id uint) error {
	return r.db.Where("id = ?", id).Delete(&models.User{}).Error
}

func (r *UserRepository) ListUsersByRole(companyID uint, role string) ([]models.User, error) {
	var users []models.User
	if err := r.db.Where("company_id = ? AND role = ?", companyID, role).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) ListUsersByRoleGlobal(role string) ([]models.User, error) {
	var users []models.User
	if err := r.db.Where("role = ?", role).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) ListUsersByCompany(companyID uint) ([]models.User, error) {
	var users []models.User
	if err := r.db.Where("company_id = ?", companyID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) WithTx(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}
