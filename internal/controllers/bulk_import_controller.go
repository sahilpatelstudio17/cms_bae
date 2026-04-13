package controllers

import (
	"net/http"

	"cms/internal/middleware"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type BulkImportController struct {
	bulkImportService *services.BulkImportService
	userImportService *services.UserImportService
}

func NewBulkImportController(bulkImportService *services.BulkImportService, userImportService *services.UserImportService) *BulkImportController {
	return &BulkImportController{
		bulkImportService: bulkImportService,
		userImportService: userImportService,
	}
}

func (c *BulkImportController) ImportEmployees(ctx *gin.Context) {
	companyID, ok := ctx.Get(middleware.ContextCompanyIDKey)
	if !ok {
		utils.Error(ctx, http.StatusUnauthorized, "company_id not found in context")
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "Failed to get file from request")
		return
	}

	result, err := c.bulkImportService.ImportEmployeesFromExcel(file, companyID.(uint))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "Failed to process Excel file")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    result,
		"message": "Import completed",
	})
}

func (c *BulkImportController) ImportUsers(ctx *gin.Context) {
	companyID, ok := ctx.Get(middleware.ContextCompanyIDKey)
	if !ok {
		utils.Error(ctx, http.StatusUnauthorized, "company_id not found in context")
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		utils.Error(ctx, http.StatusBadRequest, "Failed to get file from request")
		return
	}

	results, err := c.userImportService.ImportUsersFromFile(file, companyID.(uint))
	if err != nil {
		utils.Error(ctx, http.StatusInternalServerError, "Failed to process file: "+err.Error())
		return
	}

	// Count successful and failed
	successCount := 0
	failedCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success_count": successCount,
		"failed_count":  failedCount,
		"results":       results,
		"message":       "Import completed",
	})
}
