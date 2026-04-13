package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"cms/internal/middleware"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type ExpenseController struct {
	service *services.ExpenseService
}

func NewExpenseController(service *services.ExpenseService) *ExpenseController {
	return &ExpenseController{service: service}
}

// ListExpenses - GET /api/expenses
// Returns all expenses for the logged-in user (employees see their own, admins see company expenses)
func (ctl *ExpenseController) ListExpenses(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	p := utils.NewPagination(c)
	items, total, err := ctl.service.ListExpenses(companyID, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"items": items,
		"meta":  utils.BuildPaginationMeta(p.Page, p.Limit, total),
	})
}

// CreateExpense - POST /api/expenses
// Creates a new expense request
func (ctl *ExpenseController) CreateExpense(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "user not found")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "invalid user id")
		return
	}

	var req services.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// CreateExpense passes userID to service, which will find/create employee if needed
	expense, err := ctl.service.CreateExpense(companyID, userIDUint, req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, expense)
}

// ListPendingApprovals - GET /api/expenses/pending
// Returns pending expenses awaiting admin approval
func (ctl *ExpenseController) ListPendingApprovals(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	expenses, err := ctl.service.ListPendingApprovals(companyID)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"items": expenses,
	})
}

// ApproveExpense - POST /api/expenses/:id/approve
// Admin approves a pending expense
func (ctl *ExpenseController) ApproveExpense(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "user not found")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "invalid user id")
		return
	}

	expenseIDStr := c.Param("id")
	expenseID, err := strconv.ParseUint(expenseIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid expense id")
		return
	}

	expense, err := ctl.service.ApproveExpense(companyID, uint(expenseID), userIDUint)
	if err != nil {
		if errors.Is(err, errors.New("expense not found")) {
			utils.Error(c, http.StatusNotFound, "expense not found")
		} else {
			handleError(c, err)
		}
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"success": true,
		"data":    expense,
		"message": "expense approved successfully",
	})
}

// RejectExpense - POST /api/expenses/:id/reject
// Admin rejects a pending expense
func (ctl *ExpenseController) RejectExpense(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "user not found")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "invalid user id")
		return
	}

	var req services.ApprovalActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	expenseIDStr := c.Param("id")
	expenseID, err := strconv.ParseUint(expenseIDStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid expense id")
		return
	}

	expense, err := ctl.service.RejectExpense(companyID, uint(expenseID), userIDUint)
	if err != nil {
		if errors.Is(err, errors.New("expense not found")) {
			utils.Error(c, http.StatusNotFound, "expense not found")
		} else {
			handleError(c, err)
		}
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"success": true,
		"data":    expense,
		"message": "expense rejected successfully",
	})
}
