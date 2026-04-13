package controllers

import (
	"net/http"

	"cms/internal/middleware"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type AttendanceController struct {
	service *services.AttendanceService
}

func NewAttendanceController(service *services.AttendanceService) *AttendanceController {
	return &AttendanceController{service: service}
}

func (ctl *AttendanceController) List(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
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

func (ctl *AttendanceController) Create(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	var req services.CreateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	record, err := ctl.service.Create(companyID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, record)
}

// MarkIn - POST /api/attendance/in
// Employee marks check-in time for today
func (ctl *AttendanceController) MarkIn(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "user not found in token")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "invalid user id in token")
		return
	}

	record, err := ctl.service.MarkIn(companyID, userIDUint)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusCreated, record)
}

// MarkOut - POST /api/attendance/out
// Employee marks check-out time for today
func (ctl *AttendanceController) MarkOut(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "user not found in token")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "invalid user id in token")
		return
	}

	record, err := ctl.service.MarkOut(companyID, userIDUint)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, record)
}

// GetMyAttendance - GET /api/attendance/mine
// Employee sees their own attendance records
func (ctl *AttendanceController) GetMyAttendance(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "user not found in token")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "invalid user id in token")
		return
	}

	p := utils.NewPagination(c)
	items, total, err := ctl.service.GetMyAttendance(companyID, userIDUint, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"items": items,
		"meta":  utils.BuildPaginationMeta(p.Page, p.Limit, total),
	})
}

// GetByDate - GET /api/attendance/date?date=2026-04-06
// Admin sees attendance for a specific date
func (ctl *AttendanceController) GetByDate(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	date := c.Query("date")
	if date == "" {
		utils.Error(c, http.StatusBadRequest, "date query parameter is required")
		return
	}

	p := utils.NewPagination(c)
	items, total, err := ctl.service.GetByDate(companyID, date, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"items": items,
		"meta":  utils.BuildPaginationMeta(p.Page, p.Limit, total),
	})
}
