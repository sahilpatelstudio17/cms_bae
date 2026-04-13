package controllers

import (
	"net/http"
	"strconv"

	"cms/internal/middleware"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type SalesController struct {
	service *services.SalesService
}

func NewSalesController(service *services.SalesService) *SalesController {
	return &SalesController{service: service}
}

// ListSales - GET /api/sales
// Returns all sales for the logged-in user's company
func (ctl *SalesController) ListSales(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	p := utils.NewPagination(c)
	items, total, err := ctl.service.ListSales(companyID, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"items": items,
		"meta":  utils.BuildPaginationMeta(p.Page, p.Limit, total),
	})
}

// CreateSale - POST /api/sales
// Creates a new sale request
func (ctl *SalesController) CreateSale(c *gin.Context) {
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

	var req services.CreateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	sale, err := ctl.service.CreateSale(companyID, userIDUint, req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, sale)
}

// GetSale - GET /api/sales/:id
// Gets a specific sale
func (ctl *SalesController) GetSale(c *gin.Context) {
	_, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	idStr := c.Param("id")
	_, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid sale id")
		return
	}

	// This would need to be implemented in the service
	utils.Success(c, http.StatusOK, gin.H{"message": "not implemented"})
}

// UpdateSale - PUT /api/sales/:id
// Updates a sale
func (ctl *SalesController) UpdateSale(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	idStr := c.Param("id")
	saleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid sale id")
		return
	}

	var req services.UpsertSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	sale, err := ctl.service.UpdateSale(companyID, uint(saleID), req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, sale)
}

// DeleteSale - DELETE /api/sales/:id
// Deletes a sale
func (ctl *SalesController) DeleteSale(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	idStr := c.Param("id")
	saleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid sale id")
		return
	}

	if err := ctl.service.DeleteSale(companyID, uint(saleID)); err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{"message": "sale deleted successfully"})
}

// ListPendingApprovals - GET /api/sales/pending
// Returns pending sales awaiting admin approval
func (ctl *SalesController) ListPendingApprovals(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	sales, err := ctl.service.ListPendingApprovals(companyID)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"items": sales,
	})
}

// ApproveSale - POST /api/sales/:id/approve
// Admin approves a pending sale
func (ctl *SalesController) ApproveSale(c *gin.Context) {
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

	idStr := c.Param("id")
	saleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid sale id")
		return
	}

	sale, err := ctl.service.ApproveSale(companyID, uint(saleID), userIDUint)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, sale)
}

// RejectSale - POST /api/sales/:id/reject
// Admin rejects a pending sale
func (ctl *SalesController) RejectSale(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found")
		return
	}

	idStr := c.Param("id")
	saleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid sale id")
		return
	}

	sale, err := ctl.service.RejectSale(companyID, uint(saleID))
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, sale)
}
