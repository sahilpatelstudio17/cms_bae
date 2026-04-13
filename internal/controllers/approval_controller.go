package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"cms/internal/middleware"
	"cms/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApprovalController struct {
	approvalService *services.ApprovalService
}

func NewApprovalController(approvalService *services.ApprovalService) *ApprovalController {
	return &ApprovalController{
		approvalService: approvalService,
	}
}

// ListPendingApprovals - GET /api/approvals
// Get all pending approval requests for the logged-in admin's company
// Role-based filtering: Admin role sees all approvals, other roles see limited data
func (c *ApprovalController) ListPendingApprovals(ctx *gin.Context) {
	// Get company ID from context (set by auth middleware)
	companyID, ok := middleware.CompanyIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "company id not found"})
		return
	}

	// Get user role from context
	roleAny, ok := ctx.Get(middleware.ContextRoleKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "role not found"})
		return
	}

	role, ok := roleAny.(string)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid role"})
		return
	}

	approvals, err := c.approvalService.ListPendingApprovals(companyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve approvals"})
		return
	}

	// Role-based filtering
	// Admin role: sees all approval types
	// Manager, Staff, Developer, Employee, Salesman: restricted access (only see user/employee approvals relevant to them)
	if role != "admin" && role != "super_admin" {
		// Filter approvals for non-admin roles to only see user and employee request types
		var filteredApprovals []services.ApprovalRequestResponse
		for _, approval := range approvals {
			if approval.RequestType == "employee" || approval.RequestType == "user" {
				// Only non-admin users can see employee and user approvals
				filteredApprovals = append(filteredApprovals, approval)
			}
		}
		approvals = filteredApprovals
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    approvals,
	})
}

// ApproveUser - POST /api/approvals/:id/approve
// Admin approves a pending user registration
func (c *ApprovalController) ApproveUser(ctx *gin.Context) {
	// Get user ID and company ID from context
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "company id not found"})
		return
	}

	// Get approval ID from URL
	approvalIDStr := ctx.Param("id")
	approvalID, err := strconv.ParseUint(approvalIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}

	// Approve the user
	if err := c.approvalService.ApproveUser(companyID, uint(approvalID), userIDUint); err != nil {
		var statusCode int
		var message string

		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, errors.New("approval request not found")) {
			statusCode = http.StatusNotFound
			message = "approval request not found"
		} else if errors.Is(err, errors.New("approval request does not belong to this company")) {
			statusCode = http.StatusForbidden
			message = "approval request does not belong to this company"
		} else if errors.Is(err, errors.New("approval request is not pending")) {
			statusCode = http.StatusBadRequest
			message = "approval request is not pending"
		} else {
			statusCode = http.StatusInternalServerError
			message = "failed to approve user"
		}

		ctx.JSON(statusCode, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user approved successfully",
	})
}

// RejectUser - POST /api/approvals/:id/reject
// Admin rejects a pending user registration
func (c *ApprovalController) RejectUser(ctx *gin.Context) {
	// Get user ID and company ID from context
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	companyID, ok := middleware.CompanyIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "company id not found"})
		return
	}

	// Parse request body
	var req services.ApprovalActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get approval ID from URL
	approvalIDStr := ctx.Param("id")
	approvalID, err := strconv.ParseUint(approvalIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}

	// Reject the user
	if err := c.approvalService.RejectUser(companyID, uint(approvalID), userIDUint, req.Message); err != nil {
		var statusCode int
		var message string

		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, errors.New("approval request not found")) {
			statusCode = http.StatusNotFound
			message = "approval request not found"
		} else if errors.Is(err, errors.New("approval request does not belong to this company")) {
			statusCode = http.StatusForbidden
			message = "approval request does not belong to this company"
		} else if errors.Is(err, errors.New("approval request is not pending")) {
			statusCode = http.StatusBadRequest
			message = "approval request is not pending"
		} else {
			statusCode = http.StatusInternalServerError
			message = "failed to reject user"
		}

		ctx.JSON(statusCode, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user rejected successfully",
	})
}

// GetPendingAdminRequests - GET /api/approvals/admin/pending
// Get all pending admin approval requests (super admin only)
func (c *ApprovalController) GetPendingAdminRequests(ctx *gin.Context) {
	approvals, err := c.approvalService.GetPendingAdminRequests()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve admin approvals"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    approvals,
	})
}

// ApproveAdminRequest - POST /api/approvals/admin/:id/approve
// Super admin approves admin user creation
func (c *ApprovalController) ApproveAdminRequest(ctx *gin.Context) {
	// Get user ID from context (super admin)
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	// Get approval ID from URL
	approvalIDStr := ctx.Param("id")
	approvalID, err := strconv.ParseUint(approvalIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}

	// Approve the admin request
	user, err := c.approvalService.ApproveAdminRequest(uint(approvalID), userIDUint)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "admin request approved successfully",
		"data":    user,
	})
}

// RejectAdminRequest - POST /api/approvals/admin/:id/reject
// Super admin rejects admin user creation
func (c *ApprovalController) RejectAdminRequest(ctx *gin.Context) {
	// Get user ID from context (super admin)
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	// Parse request body
	var req struct {
		Reason string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get approval ID from URL
	approvalIDStr := ctx.Param("id")
	approvalID, err := strconv.ParseUint(approvalIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}

	// Reject the admin request
	if err := c.approvalService.RejectAdminRequest(uint(approvalID), userIDUint, req.Reason); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "admin request rejected successfully",
	})
}

// ApproveCompanySignup - POST /api/approvals/company/:id/approve
// Admin approves a company signup request and creates company, user, subscription
func (c *ApprovalController) ApproveCompanySignup(ctx *gin.Context) {
	// Get user ID from context (admin)
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	// Get approval ID from URL
	approvalIDStr := ctx.Param("id")
	approvalID, err := strconv.ParseUint(approvalIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}

	// Approve the company signup
	response, err := c.approvalService.ApproveCompanySignup(uint(approvalID), userIDUint)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "company signup approved! company and user account created",
		"data":    response,
	})
}

// GetPendingCompanySignups - GET /api/approvals/company/pending
// Get all pending company signup requests (admin only)
func (c *ApprovalController) GetPendingCompanySignups(ctx *gin.Context) {
	approvals, err := c.approvalService.GetPendingCompanySignups()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve company signups"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    approvals,
	})
}

// RequestEmployeeApproval - POST /api/approvals/employee/request
// Request approval for a new employee
func (c *ApprovalController) RequestEmployeeApproval(ctx *gin.Context) {
	// Get company ID from context
	companyID, ok := middleware.CompanyIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "company id not found"})
		return
	}

	var req struct {
		Name     string  `json:"name" binding:"required"`
		Position string  `json:"position" binding:"required"`
		Salary   float64 `json:"salary" binding:"required,gt=0"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	response, err := c.approvalService.RequestEmployeeApproval(companyID, services.EmployeeApprovalRequest{
		Name:     req.Name,
		Position: req.Position,
		Salary:   req.Salary,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "employee request submitted for approval",
		"data":    response,
	})
}

// GetPendingEmployeeRequests - GET /api/approvals/employee/pending
// Get all pending employee approval requests for admin
func (c *ApprovalController) GetPendingEmployeeRequests(ctx *gin.Context) {
	// Get company ID from context
	companyID, ok := middleware.CompanyIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "company id not found"})
		return
	}

	approvals, err := c.approvalService.GetPendingEmployeeRequests(companyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve employee requests"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    approvals,
	})
}

// ApproveEmployeeRequest - POST /api/approvals/employee/:id/approve
// Admin approves employee request
func (c *ApprovalController) ApproveEmployeeRequest(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	approvalIDStr := ctx.Param("id")
	approvalID, err := strconv.ParseUint(approvalIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}

	employee, err := c.approvalService.ApproveEmployeeRequest(uint(approvalID), userIDUint)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "employee request approved",
		"data":    employee,
	})
}

// RejectEmployeeRequest - POST /api/approvals/employee/:id/reject
// Admin rejects employee request
func (c *ApprovalController) RejectEmployeeRequest(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	approvalIDStr := ctx.Param("id")
	approvalID, err := strconv.ParseUint(approvalIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := c.approvalService.RejectEmployeeRequest(uint(approvalID), userIDUint, req.Reason); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "employee request rejected",
	})
}

// RequestUserApproval - POST /api/approvals/user/request
// Request approval for a new user
func (c *ApprovalController) RequestUserApproval(ctx *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "company id not found"})
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Role     string `json:"role" binding:"required"`
		Status   string `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	response, err := c.approvalService.RequestUserApproval(companyID, services.UserApprovalRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		Status:   req.Status,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user request submitted for approval",
		"data":    response,
	})
}

// GetPendingUserApprovals - GET /api/approvals/user/pending
// Get all pending user approval requests for admin
func (c *ApprovalController) GetPendingUserApprovals(ctx *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "company id not found"})
		return
	}

	approvals, err := c.approvalService.GetPendingUserApprovals(companyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user requests"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    approvals,
	})
}

// ApproveUserRequest - POST /api/approvals/user/:id/approve
// Admin approves user request
func (c *ApprovalController) ApproveUserRequest(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	approvalIDStr := ctx.Param("id")
	approvalID, err := strconv.ParseUint(approvalIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}

	user, err := c.approvalService.ApproveUserRequest(uint(approvalID), userIDUint)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user request approved",
		"data":    user,
	})
}

// RejectUserRequest - POST /api/approvals/user/:id/reject
// Admin rejects user request
func (c *ApprovalController) RejectUserRequest(ctx *gin.Context) {
	userID, ok := ctx.Get(middleware.ContextUserIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	approvalIDStr := ctx.Param("id")
	approvalID, err := strconv.ParseUint(approvalIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := c.approvalService.RejectUserRequest(uint(approvalID), userIDUint, req.Reason); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user request rejected",
	})
}
