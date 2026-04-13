package services

import (
	"strings"

	"cms/internal/models"
	"cms/internal/repositories"
)

type CompanyService struct {
	repo *repositories.CompanyRepository
}

type UpdateCompanyRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=120"`
	Email string `json:"email" binding:"required,email"`
}

func NewCompanyService(repo *repositories.CompanyRepository) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) Get(companyID uint) (*models.Company, error) {
	return s.repo.GetByID(companyID)
}

func (s *CompanyService) Update(companyID uint, input UpdateCompanyRequest) (*models.Company, error) {
	company, err := s.repo.GetByID(companyID)
	if err != nil {
		return nil, err
	}

	company.Name = strings.TrimSpace(input.Name)
	company.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if err := s.repo.Update(company); err != nil {
		return nil, err
	}

	return company, nil
}
