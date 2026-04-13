package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page    int   `json:"page"`
	Limit   int   `json:"limit"`
	Total   int64 `json:"total"`
	Pages   int   `json:"pages"`
	HasNext bool  `json:"has_next"`
	HasPrev bool  `json:"has_prev"`
	Offset  int   `json:"-"`
}

func NewPagination(c *gin.Context) Pagination {
	page := parsePositiveInt(c.Query("page"), 1)
	limit := parsePositiveInt(c.Query("limit"), 20)
	if limit > 100 {
		limit = 100
	}

	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func BuildPaginationMeta(page, limit int, total int64) Pagination {
	pages := int(math.Ceil(float64(total) / float64(limit)))
	if pages == 0 {
		pages = 1
	}
	return Pagination{
		Page:    page,
		Limit:   limit,
		Total:   total,
		Pages:   pages,
		HasNext: page < pages,
		HasPrev: page > 1,
	}
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
