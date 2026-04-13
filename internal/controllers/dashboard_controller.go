package controllers

import (
	"cms/internal/middleware"
	"cms/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	dashboardService *services.DashboardService
}

func NewDashboardController(dashboardService *services.DashboardService) *DashboardController {
	return &DashboardController{
		dashboardService: dashboardService,
	}
}

type DashboardResponse struct {
	Data interface{} `json:"data"`
}

func (c *DashboardController) GetStats(ctx *gin.Context) {
	companyID, ok := ctx.Get(middleware.ContextCompanyIDKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "company_id not found in context"})
		return
	}

	stats, err := c.dashboardService.GetDashboardStats(companyID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch dashboard stats"})
		return
	}

	ctx.JSON(http.StatusOK, DashboardResponse{Data: stats})
}
