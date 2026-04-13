package controllers

import (
	"net/http"
	"strconv"

	"cms/internal/middleware"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type RoleAssignmentController struct {
	roleAssignmentService *services.RoleAssignmentService
}

func NewRoleAssignmentController(roleAssignmentService *services.RoleAssignmentService) *RoleAssignmentController {
	return &RoleAssignmentController{
		roleAssignmentService: roleAssignmentService,
	}
}

// RequestRoleAssignment creates a role assignment approval request
func (c *RoleAssignmentController) RequestRoleAssignment(ctx *gin.Context) {
	companyID, ok := ctx.Get(middleware.ContextCompanyIDKey)
	if !ok {
		utils.Error(ctx, http.StatusUnauthorized, "company_id not found in context")
		return
	}

	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(ctx, http.StatusUnauthorized, "user_id not found in context")
		return
	}

	var req services.RoleAssignmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	approval, err := c.roleAssignmentService.RequestRoleAssignment(
		req,
		companyID.(uint),
		userID.(uint),
	)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data":    approval,
		"message": "Role assignment request created. Waiting for approval.",
	})
}

// ApproveRoleAssignment approves a role assignment request
func (c *RoleAssignmentController) ApproveRoleAssignment(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(ctx, http.StatusUnauthorized, "user_id not found in context")
		return
	}

	approvalID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "Invalid approval ID")
		return
	}

	user, err := c.roleAssignmentService.ApproveRoleAssignment(uint(approvalID), userID.(uint))
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    user,
		"message": "Role assignment approved. User account created successfully.",
	})
}

// RejectRoleAssignment rejects a role assignment request
func (c *RoleAssignmentController) RejectRoleAssignment(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(ctx, http.StatusUnauthorized, "user_id not found in context")
		return
	}

	approvalID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "Invalid approval ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Error(ctx, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	err = c.roleAssignmentService.RejectRoleAssignment(uint(approvalID), userID.(uint), req.Reason)
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Role assignment request rejected.",
	})
}

// GetPendingRoleAssignments gets all pending role assignment requests
func (c *RoleAssignmentController) GetPendingRoleAssignments(ctx *gin.Context) {
	companyID, ok := ctx.Get(middleware.ContextCompanyIDKey)
	if !ok {
		utils.Error(ctx, http.StatusUnauthorized, "company_id not found in context")
		return
	}

	approvals, err := c.roleAssignmentService.GetPendingRoleAssignments(companyID.(uint))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "Failed to fetch pending requests")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": approvals,
	})
}
