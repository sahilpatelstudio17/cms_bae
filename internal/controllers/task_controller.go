package controllers

import (
	"net/http"
	"strconv"

	"cms/internal/middleware"
	"cms/internal/services"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	service *services.TaskService
}

func NewTaskController(service *services.TaskService) *TaskController {
	return &TaskController{service: service}
}

func (ctl *TaskController) List(c *gin.Context) {
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

func (ctl *TaskController) Create(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	var req services.UpsertTaskRequest
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

func (ctl *TaskController) Update(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid task id")
		return
	}

	var req services.UpsertTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	item, err := ctl.service.Update(companyID, uint(taskID), req)
	if err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, item)
}

func (ctl *TaskController) Delete(c *gin.Context) {
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := ctl.service.Delete(companyID, uint(taskID)); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
