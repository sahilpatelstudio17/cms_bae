package controllers

import (
	"net/http"
	"strconv"

	"cms/internal/middleware"
	"cms/internal/repositories"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type EmployeeController struct {
	service           *services.EmployeeService
	userRepo          *repositories.UserRepository
	userManagementSvc *services.UserManagementService
}

func NewEmployeeController(service *services.EmployeeService, userRepo *repositories.UserRepository, userManagementSvc *services.UserManagementService) *EmployeeController {
	return &EmployeeController{service: service, userRepo: userRepo, userManagementSvc: userManagementSvc}
}

func (ctl *EmployeeController) List(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)

	// Fallback: if companyID not in token, get from user record
	if !ok {
		userID, userOk := middleware.UserIDFromContext(c)
		if userOk {
			// Get user to extract company ID
			userData, err := ctl.userRepo.GetUserByID(userID)
			if err == nil && userData != nil && userData.CompanyID > 0 {
				companyID = userData.CompanyID
				ok = true
			}
		}
	}

	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	p := utils.NewPagination(c)
	items, total, err := ctl.service.List(companyID, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"items": items,
		"meta":  utils.BuildPaginationMeta(p.Page, p.Limit, total),
	})
}

func (ctl *EmployeeController) Create(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	var req services.UpsertEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	item, err := ctl.service.Create(companyID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, item)
}

func (ctl *EmployeeController) Update(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	employeeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid employee id")
		return
	}

	var req services.UpsertEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	item, err := ctl.service.Update(companyID, uint(employeeID), req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, item)
}

// CreateWithUser - Create employee and auto-generate user account
func (ctl *EmployeeController) CreateWithUser(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	var req services.CreateEmployeeWithUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := ctl.userManagementSvc.CreateEmployeeWithUser(companyID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, result)
}

func (ctl *EmployeeController) Delete(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	employeeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid employee id")
		return
	}

	if err := ctl.service.Delete(companyID, uint(employeeID)); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
