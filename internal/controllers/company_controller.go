package controllers

import (
	"net/http"

	"cms/internal/middleware"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type CompanyController struct {
	service *services.CompanyService
}

func NewCompanyController(service *services.CompanyService) *CompanyController {
	return &CompanyController{service: service}
}

func (ctl *CompanyController) Get(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	company, err := ctl.service.Get(companyID)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, company)
}

func (ctl *CompanyController) Update(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	var req services.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	company, err := ctl.service.Update(companyID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, company)
}
