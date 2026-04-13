package repositories

import (
	"cms/internal/models"

	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *AuthRepository) CreateCompanyWithAdmin(company *models.Company, user *models.User, subscription *models.Subscription) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(company).Error; err != nil {
			return err
		}

		user.CompanyID = company.ID
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		subscription.CompanyID = company.ID
		if err := tx.Create(subscription).Error; err != nil {
			return err
		}

		return nil
	})
}
