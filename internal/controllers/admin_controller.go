package controllers

import (
	"net/http"
	"strconv"

	"cms/internal/middleware"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	service *services.UserManagementService
}

func NewAdminController(service *services.UserManagementService) *AdminController {
	return &AdminController{service: service}
}

// ListAdmins - Super Admin lists all admins
func (ctl *AdminController) ListAdmins(c *gin.Context) {
	// Get user role from context
	role, ok := c.Get(middleware.ContextRoleKey)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "role not found in token")
		return
	}

	roleStr, ok := role.(string)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "invalid role")
		return
	}

	var admins interface{}

	// If super admin, show all admins across all companies
	if roleStr == "super_admin" {
		adminList, err := ctl.service.ListAllAdmins()
		if err != nil {
			handleError(c, err)
			return
		}
		admins = adminList
	} else {
		// Regular admin - show admins from their company only
		companyID, ok := middleware.CompanyIDFromContext(c)
		if !ok {
			utils.Error(c, http.StatusUnauthorized, "company not found in token")
			return
		}

		adminList, err := ctl.service.ListAdminsByCompany(companyID)
		if err != nil {
			handleError(c, err)
			return
		}
		admins = adminList
	}

	utils.Success(c, http.StatusOK, gin.H{
		"items": admins,
	})
}

// CreateAdmin - Super Admin creates a new admin
func (ctl *AdminController) CreateAdmin(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	var req services.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	admin, err := ctl.service.CreateAdmin(companyID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, admin)
}

// DeleteAdmin - Super Admin deletes an admin
func (ctl *AdminController) DeleteAdmin(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	adminID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid admin id")
		return
	}

	if err := ctl.service.DeleteAdminUser(companyID, uint(adminID)); err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "admin deleted successfully",
	})
}
