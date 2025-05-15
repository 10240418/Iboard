package http_base_controller

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
)

type InterfaceFileController interface {
	Create()
	CreateMany()
	Get()
	Update()
	Delete()
	GetOne()
}

type FileController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewFileController(ctx *gin.Context, container *container.ServiceContainer) *FileController {
	return &FileController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncFile returns a gin.HandlerFunc for the specified method
func HandleFuncFile(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.Create()
		}
	case "createMany":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.CreateMany()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *FileController) Create() {
	var form struct {
		Path         string                 `json:"path" binding:"required"`
		Size         int64                  `json:"size" binding:"required"`
		MimeType     string                 `json:"mimeType" binding:"required"`
		Oss          string                 `json:"oss" binding:"required"`
		UploaderType field.FileUploaderType `json:"uploaderType" binding:"required"`
		UploaderID   uint                   `json:"uploaderId" binding:"required"`
		Md5          string                 `json:"md5"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	if form.Md5 == "" {
		data := form.Path + time.Now().String()
		hash := md5.Sum([]byte(data))
		form.Md5 = hex.EncodeToString(hash[:])
	}

	file := &base_models.File{
		Path:         form.Path,
		Size:         form.Size,
		MimeType:     form.MimeType,
		Oss:          form.Oss,
		UploaderType: form.UploaderType,
		UploaderID:   form.UploaderID,
		Md5:          form.Md5,
	}

	if err := c.Container.GetService("file").(base_services.InterfaceFileService).Create(file); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create file failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create file success",
		"data":    file,
	})
}

func (c *FileController) CreateMany() {
	var forms []struct {
		Path         string                 `json:"path" binding:"required"`
		Size         int64                  `json:"size" binding:"required"`
		MimeType     string                 `json:"mimeType" binding:"required"`
		Oss          string                 `json:"oss" binding:"required"`
		UploaderType field.FileUploaderType `json:"uploaderType" binding:"required"`
		UploaderID   uint                   `json:"uploaderId" binding:"required"`
		Md5          string                 `json:"md5"`
	}

	if err := c.Ctx.ShouldBindJSON(&forms); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	var files []*base_models.File
	for _, form := range forms {
		if form.Md5 == "" {
			data := form.Path + time.Now().String()
			hash := md5.Sum([]byte(data))
			form.Md5 = hex.EncodeToString(hash[:])
		}

		file := &base_models.File{
			Path:         form.Path,
			Size:         form.Size,
			MimeType:     form.MimeType,
			Oss:          form.Oss,
			UploaderType: form.UploaderType,
			UploaderID:   form.UploaderID,
			Md5:          form.Md5,
		}
		files = append(files, file)
	}

	for _, file := range files {
		if err := c.Container.GetService("file").(base_services.InterfaceFileService).Create(file); err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "create file failed",
				"file":    file,
			})
			return
		}
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create files success",
		"data":    files,
	})
}

func (c *FileController) Get() {
	var searchQuery struct {
		MimeType     string `form:"mimeType"`
		Oss          string `form:"oss"`
		UploaderType string `form:"uploaderType"`
	}
	if err := c.Ctx.ShouldBindQuery(&searchQuery); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	pagination := struct {
		PageSize int  `form:"pageSize"`
		PageNum  int  `form:"pageNum"`
		Desc     bool `form:"desc"`
	}{
		PageSize: 10,
		PageNum:  1,
		Desc:     false,
	}

	if err := c.Ctx.ShouldBindQuery(&pagination); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	queryMap := utils.StructToMap(searchQuery)
	paginationMap := map[string]interface{}{
		"pageSize": pagination.PageSize,
		"pageNum":  pagination.PageNum,
		"desc":     pagination.Desc,
	}

	files, paginationResult, err := c.Container.GetService("file").(base_services.InterfaceFileService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       files,
		"pagination": paginationResult,
	})
}

func (c *FileController) Update() {
	var form struct {
		ID           uint   `json:"id" binding:"required"`
		Path         string `json:"path"`
		Size         int64  `json:"size"`
		MimeType     string `json:"mimeType"`
		Oss          string `json:"oss"`
		Uploader     string `json:"uploader"`
		UploaderType string `json:"uploaderType"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Path != "" {
		updates["path"] = form.Path
	}
	if form.Size != 0 {
		updates["size"] = form.Size
	}
	if form.MimeType != "" {
		updates["mime_type"] = form.MimeType
	}
	if form.Oss != "" {
		updates["oss"] = form.Oss
	}
	if form.Uploader != "" {
		updates["uploader"] = form.Uploader
	}
	if form.UploaderType != "" {
		updates["uploader_type"] = form.UploaderType
	}

	if err := c.Container.GetService("file").(base_services.InterfaceFileService).Update(form.ID, updates); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "update file success"})
}

func (c *FileController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("file").(base_services.InterfaceFileService).Delete(form.IDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete file success"})
}

func (c *FileController) GetOne() {
	if c.Container.GetService("jwt") == nil {
		c.Ctx.JSON(500, gin.H{
			"error":   "jwt service is nil",
			"message": "internal server error",
		})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Invalid file ID",
			"message": "Please check the ID format",
		})
		return
	}

	file, err := c.Container.GetService("file").(base_services.InterfaceFileService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get file",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get file success",
		"data":    file,
	})
}
