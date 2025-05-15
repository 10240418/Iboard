package http_relationship_controller

import (
	"strconv"

	"github.com/The-Healthist/iboard_http_service/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceFileNoticeController interface {
	BindFile()
	UnbindFile()
	GetNoticeByFile()
	GetFileByNotice()
}

type FileNoticeController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewFileNoticeController(ctx *gin.Context, container *container.ServiceContainer) *FileNoticeController {
	return &FileNoticeController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncFileNotice returns a gin.HandlerFunc for the specified method
func HandleFuncFileNotice(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewFileNoticeController(ctx, container)
		switch method {
		case "bindFile":
			controller.BindFile()
		case "unbindFile":
			controller.UnbindFile()
		case "getNoticeByFile":
			controller.GetNoticeByFile()
		case "getFileByNotice":
			controller.GetFileByNotice()
		default:
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *FileNoticeController) getService() relationship_service.InterfaceFileNoticeService {
	return c.Container.GetService("fileNotice").(relationship_service.InterfaceFileNoticeService)
}

func (c *FileNoticeController) BindFile() {
	var form struct {
		NoticeID uint `json:"noticeId" binding:"required"`
		FileID   uint `json:"fileId" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters"})
		return
	}

	// 检查通知是否存在
	exists, err := c.getService().NoticeExists(form.NoticeID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Notice not found"})
		return
	}

	// 检查文件是否存在
	exists, err = c.getService().FileExists(form.FileID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "File not found"})
		return
	}

	if err := c.getService().BindFile(form.NoticeID, form.FileID); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "File bound successfully"})
}

func (c *FileNoticeController) UnbindFile() {
	noticeIDStr := c.Ctx.Query("noticeId")
	if noticeIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "noticeId is required"})
		return
	}

	noticeID, err := strconv.ParseUint(noticeIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid noticeId"})
		return
	}

	if err := c.getService().UnbindFile(uint(noticeID)); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "File unbound successfully"})
}

func (c *FileNoticeController) GetNoticeByFile() {
	fileIDStr := c.Ctx.Query("fileId")
	if fileIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "fileId is required"})
		return
	}

	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid fileId"})
		return
	}

	notice, err := c.getService().GetNoticeByFileID(uint(fileID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch notice"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": notice})
}

func (c *FileNoticeController) GetFileByNotice() {
	noticeIDStr := c.Ctx.Query("noticeId")
	if noticeIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "noticeId is required"})
		return
	}

	noticeID, err := strconv.ParseUint(noticeIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid noticeId"})
		return
	}

	file, err := c.getService().GetFileByNoticeID(uint(noticeID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch file"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": file})
}
