package building_admin_controllers

import (
	"strconv"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/The-Healthist/iboard_http_service/utils/response"
	"github.com/gin-gonic/gin"
)

type BuildingAdminNoticeController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingAdminNoticeController(
	ctx *gin.Context,
	container *container.ServiceContainer,
) *BuildingAdminNoticeController {
	return &BuildingAdminNoticeController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuildingAdminNotice returns a gin.HandlerFunc for the specified method
func HandleFuncBuildingAdminNotice(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "getNotices":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminNoticeController(ctx, container)
			controller.GetNotices()
		}
	case "getNotice":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminNoticeController(ctx, container)
			controller.GetNotice()
		}
	case "createNotice":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminNoticeController(ctx, container)
			controller.CreateNotice()
		}
	case "updateNotice":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminNoticeController(ctx, container)
			controller.UpdateNotice()
		}
	case "deleteNotice":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminNoticeController(ctx, container)
			controller.DeleteNotice()
		}
	case "getUploadParams":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminNoticeController(ctx, container)
			controller.GetUploadParams()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *BuildingAdminNoticeController) GetNotices() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// 处理分页参数
	pageSize, _ := strconv.Atoi(c.Ctx.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.Ctx.DefaultQuery("pageNum", "1"))
	desc := c.Ctx.DefaultQuery("desc", "true") == "true"

	// 处理查询参数
	query := make(map[string]interface{})
	if noticeType := c.Ctx.Query("type"); noticeType != "" {
		query["type"] = field.NoticeType(noticeType)
	}
	if status := c.Ctx.Query("status"); status != "" {
		query["status"] = field.Status(status)
	}
	if fileID := c.Ctx.Query("fileId"); fileID != "" {
		if id, err := strconv.ParseUint(fileID, 10, 64); err == nil {
			query["fileId"] = uint(id)
		}
	}
	if fileType := c.Ctx.Query("fileType"); fileType != "" {
		query["fileType"] = field.FileType(fileType)
	}

	paginate := map[string]interface{}{
		"pageSize": pageSize,
		"pageNum":  pageNum,
		"desc":     desc,
	}

	notices, pagination, err := c.Container.GetService("buildingAdminNotice").(building_admin_services.InterfaceBuildingAdminNoticeService).Get(email, query, paginate)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       notices,
		"pagination": pagination,
	})
}

func (c *BuildingAdminNoticeController) GetNotice() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid notice ID"})
		return
	}

	notice, err := c.Container.GetService("buildingAdminNotice").(building_admin_services.InterfaceBuildingAdminNoticeService).GetByID(uint(id), email)
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": notice})
}

type CreateNoticeRequest struct {
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description"`
	Type        field.NoticeType `json:"type" binding:"required"`
	Status      field.Status     `json:"status" binding:"required"`
	StartTime   *time.Time       `json:"startTime" binding:"required"`
	EndTime     *time.Time       `json:"endTime" binding:"required"`
	IsPublic    bool             `json:"isPublic"`
	FileID      *uint            `json:"fileId"`
	FileType    field.FileType   `json:"fileType"`
}

func (c *BuildingAdminNoticeController) CreateNotice() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateNoticeRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c.Ctx, err)
		return
	}

	notice := &base_models.Notice{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Status:      req.Status,
		StartTime:   *req.StartTime,
		EndTime:     *req.EndTime,
		FileID:      req.FileID,
		FileType:    req.FileType,
		IsPublic:    false, // 强制设置为 false
	}

	if err := c.Container.GetService("buildingAdminNotice").(building_admin_services.InterfaceBuildingAdminNoticeService).Create(notice, email); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Notice created successfully",
		"data":    notice,
	})
}

func (c *BuildingAdminNoticeController) UpdateNotice() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ID      uint64                 `json:"id" binding:"required"`
		Updates map[string]interface{} `json:"updates"`
	}
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 强制设置 isPublic 为 false
	if _, ok := req.Updates["is_public"]; ok {
		req.Updates["is_public"] = false
	}

	if err := c.Container.GetService("buildingAdminNotice").(building_admin_services.InterfaceBuildingAdminNoticeService).Update(uint(req.ID), email, req.Updates); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "Notice updated successfully"})
}

func (c *BuildingAdminNoticeController) DeleteNotice() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid notice ID"})
		return
	}

	if err := c.Container.GetService("buildingAdminNotice").(building_admin_services.InterfaceBuildingAdminNoticeService).Delete(uint(id), email); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "Notice deleted successfully"})
}

func (c *BuildingAdminNoticeController) GetUploadParams() {
	var req struct {
		FileName string `json:"fileName" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Missing required parameters"})
		return
	}

	// Get upload policy
	policy, err := c.Container.GetService("uploadService").(base_services.IUploadService).GetUploadParams(req.FileName)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": policy})
}
