package http_relationship_controller

import (
	"strconv"

	"github.com/The-Healthist/iboard_http_service/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceFileAdvertisementController interface {
	BindFile()
	UnbindFile()
	GetAdvertisementByFile()
	GetFileByAdvertisement()
}

type FileAdvertisementController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewFileAdvertisementController(ctx *gin.Context, container *container.ServiceContainer) *FileAdvertisementController {
	return &FileAdvertisementController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncFileAdvertisement returns a gin.HandlerFunc for the specified method
func HandleFuncFileAdvertisement(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewFileAdvertisementController(ctx, container)
		switch method {
		case "bindFile":
			controller.BindFile()
		case "unbindFile":
			controller.UnbindFile()
		case "getAdvertisementByFile":
			controller.GetAdvertisementByFile()
		case "getFileByAdvertisement":
			controller.GetFileByAdvertisement()
		default:
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *FileAdvertisementController) getService() relationship_service.InterfaceFileAdvertisementService {
	return c.Container.GetService("fileAdvertisement").(relationship_service.InterfaceFileAdvertisementService)
}

func (c *FileAdvertisementController) BindFile() {
	var form struct {
		AdvertisementID uint `json:"advertisementId" binding:"required"`
		FileID          uint `json:"fileId" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters"})
		return
	}

	exists, err := c.getService().AdvertisementExists(form.AdvertisementID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Advertisement not found"})
		return
	}

	exists, err = c.getService().FileExists(form.FileID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "File not found"})
		return
	}

	if err := c.getService().BindFile(form.AdvertisementID, form.FileID); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "File bound successfully"})
}

func (c *FileAdvertisementController) UnbindFile() {
	advertisementIDStr := c.Ctx.Query("advertisementId")
	if advertisementIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "advertisementId is required"})
		return
	}

	advertisementID, err := strconv.ParseUint(advertisementIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid advertisementId"})
		return
	}

	if err := c.getService().UnbindFile(uint(advertisementID)); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "File unbound successfully"})
}

func (c *FileAdvertisementController) GetAdvertisementByFile() {
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

	advertisement, err := c.getService().GetAdvertisementByFileID(uint(fileID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch advertisement"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": advertisement})
}

func (c *FileAdvertisementController) GetFileByAdvertisement() {
	advertisementIDStr := c.Ctx.Query("advertisementId")
	if advertisementIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "advertisementId is required"})
		return
	}

	advertisementID, err := strconv.ParseUint(advertisementIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid advertisementId"})
		return
	}

	file, err := c.getService().GetFileByAdvertisementID(uint(advertisementID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch file"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": file})
}
