package controllers

import (
	"errors"
	"net/http"

	"cms/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Error(c, http.StatusNotFound, "resource not found")
		return
	}
	utils.Error(c, http.StatusBadRequest, err.Error())
}
