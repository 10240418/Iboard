package pagination

import (
	"github.com/gin-gonic/gin"
)

type Pagination struct {
	PageNum  int `form:"pageNum" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

func GetPaginationParams(c *gin.Context) Pagination {
	pagination := Pagination{
		PageNum:  1,
		PageSize: 10,
	}
	c.ShouldBindQuery(&pagination)
	return pagination
}
