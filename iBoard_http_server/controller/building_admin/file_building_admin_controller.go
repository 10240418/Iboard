package building_admin_controllers

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
)

type BuildingAdminFileController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingAdminFileController(
	ctx *gin.Context,
	container *container.ServiceContainer,
) *BuildingAdminFileController {
	return &BuildingAdminFileController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuildingAdminFile returns a gin.HandlerFunc for the specified method
func HandleFuncBuildingAdminFile(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "getFiles":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminFileController(ctx, container)
			controller.GetFiles()
		}
	case "getFile":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminFileController(ctx, container)
			controller.GetFile()
		}
	case "uploadFile":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminFileController(ctx, container)
			controller.UploadFile()
		}
	case "updateFile":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminFileController(ctx, container)
			controller.UpdateFile()
		}
	case "deleteFile":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminFileController(ctx, container)
			controller.DeleteFile()
		}
	case "downloadFile":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminFileController(ctx, container)
			controller.DownloadFile()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *BuildingAdminFileController) GetFiles() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var searchQuery struct {
		Search   string `form:"search"`
		MimeType string `form:"mimeType"`
		Oss      string `form:"oss"`
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

	files, paginationResult, err := c.Container.GetService("buildingAdminFile").(building_admin_services.InterfaceBuildingAdminFileService).Get(email, queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       files,
		"pagination": paginationResult,
	})
}

func (c *BuildingAdminFileController) GetFile() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := c.Container.GetService("buildingAdminFile").(building_admin_services.InterfaceBuildingAdminFileService).GetByID(uint(id), email)
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get file success",
		"data":    file,
	})
}

func (c *BuildingAdminFileController) UploadFile() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var form struct {
		Path     string                 `json:"path" binding:"required"`
		Size     int64                  `json:"size" binding:"required"`
		Md5      string                 `json:"md5" binding:"required"`
		MimeType string                 `json:"mimeType" binding:"required"`
		Oss      string                 `json:"oss"`
		Type     field.FileUploaderType `json:"type"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	// 计算文件的 MD5
	hasher := md5.New()
	hasher.Write([]byte(form.Path))
	calculatedMd5 := hex.EncodeToString(hasher.Sum(nil))

	// 验证 MD5
	if calculatedMd5 != form.Md5 {
		c.Ctx.JSON(400, gin.H{
			"error":   "MD5 mismatch",
			"message": "File integrity check failed",
		})
		return
	}

	file := &base_models.File{
		Path:         form.Path,
		Size:         form.Size,
		Md5:          form.Md5,
		MimeType:     form.MimeType,
		Oss:          form.Oss,
		Uploader:     email,
		UploaderType: "building_admin",
	}

	if err := c.Container.GetService("buildingAdminFile").(building_admin_services.InterfaceBuildingAdminFileService).Create(file, email); err != nil {
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

func (c *BuildingAdminFileController) UpdateFile() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var form struct {
		ID       uint   `json:"id" binding:"required"`
		Path     string `json:"path"`
		MimeType string `json:"mimeType"`
		Oss      string `json:"oss"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Path != "" {
		updates["path"] = form.Path
	}
	if form.MimeType != "" {
		updates["mime_type"] = form.MimeType
	}
	if form.Oss != "" {
		updates["oss"] = form.Oss
	}

	if err := c.Container.GetService("buildingAdminFile").(building_admin_services.InterfaceBuildingAdminFileService).Update(form.ID, email, updates); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取更新后的文件信息
	updatedFile, err := c.Container.GetService("buildingAdminFile").(building_admin_services.InterfaceBuildingAdminFileService).GetByID(form.ID, email)
	if err != nil {
		c.Ctx.JSON(200, gin.H{
			"message": "update file success, but failed to load updated data",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "update file success",
		"data":    updatedFile,
	})
}

func (c *BuildingAdminFileController) DeleteFile() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	if err := c.Container.GetService("buildingAdminFile").(building_admin_services.InterfaceBuildingAdminFileService).Delete(uint(id), email); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete file success"})
}

func (c *BuildingAdminFileController) DownloadFile() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := c.Container.GetService("buildingAdminFile").(building_admin_services.InterfaceBuildingAdminFileService).GetByID(uint(id), email)
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.File(file.Path)
}
