package controllers

import (
	"net/http"

	"cms/internal/middleware"
	"cms/internal/models"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service         *services.AuthService
	approvalService *services.ApprovalService
}

func NewAuthController(service *services.AuthService, approvalService *services.ApprovalService) *AuthController {
	return &AuthController{
		service:         service,
		approvalService: approvalService,
	}
}

func (ctl *AuthController) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := ctl.service.Register(req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, response)
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := ctl.service.Login(req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, response)
}

func (ctl *AuthController) RegisterWithApproval(c *gin.Context) {
	var req services.RegisterUserWithApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := ctl.approvalService.RegisterUserWithApproval(req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, gin.H{
		"message": "registration successful, waiting for admin approval",
		"data":    response,
	})
}

func (ctl *AuthController) RequestAdminApproval(c *gin.Context) {
	var req services.AdminApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := ctl.approvalService.RequestAdminApproval(req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, gin.H{
		"message": "admin request created, waiting for super admin approval",
		"data":    response,
	})
}

func (ctl *AuthController) GetMe(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "user not found in token")
		return
	}

	user, err := ctl.service.GetUserByID(userID)
	if err != nil {
		handleError(c, err)
		return
	}

	// If user has CreatedBy, fetch admin info
	var adminInfo *models.User
	if user.CreatedBy > 0 {
		adminInfo, _ = ctl.service.GetUserByID(user.CreatedBy)
	}

	// Return user with admin info
	response := gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"status":     user.Status,
		"company_id": user.CompanyID,
		"created_by": user.CreatedBy,
		"created_at": user.CreatedAt,
		"admin":      nil,
	}

	if adminInfo != nil {
		response["admin"] = gin.H{
			"id":    adminInfo.ID,
			"name":  adminInfo.Name,
			"email": adminInfo.Email,
			"role":  adminInfo.Role,
		}
	}

	utils.Success(c, http.StatusOK, response)
}

type UpdateProfileRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=120"`
	Email string `json:"email" binding:"required,email"`
}

func (ctl *AuthController) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "user not found in token")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := ctl.service.UpdateUserProfile(userID, req.Name, req.Email)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, user)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

func (ctl *AuthController) ChangePassword(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "user not found in token")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := ctl.service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "password changed successfully",
	})
}
