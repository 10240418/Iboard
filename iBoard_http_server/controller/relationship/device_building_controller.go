package http_relationship_controller

import (
	"strconv"

	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

type DeviceBuildingController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewDeviceBuildingController(ctx *gin.Context, container *container.ServiceContainer) *DeviceBuildingController {
	return &DeviceBuildingController{
		Ctx:       ctx,
		Container: container,
	}
}

func HandleFuncDeviceBuilding(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewDeviceBuildingController(ctx, container)

		switch method {
		case "bindDevice":
			controller.BindDevice()
		case "unbindDevice":
			controller.UnbindDevice()
		case "getDevicesByBuilding":
			controller.GetDevicesByBuilding()
		case "getBuildingByDevice":
			controller.GetBuildingByDevice()
		default:
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *DeviceBuildingController) BindDevice() {
	var form struct {
		DeviceIDs  []uint `json:"deviceIds" binding:"required"`
		BuildingID uint   `json:"buildingId" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).BindDevices(form.BuildingID, form.DeviceIDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "bind devices success"})
}

func (c *DeviceBuildingController) UnbindDevice() {
	var form struct {
		DeviceID uint `json:"deviceId" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).UnbindDevice(form.DeviceID); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "unbind device success"})
}

func (c *DeviceBuildingController) GetDevicesByBuilding() {
	buildingIDStr := c.Ctx.Query("buildingId")
	buildingID, err := strconv.ParseUint(buildingIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid building ID"})
		return
	}

	devices, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetDevicesByBuildingWithStatus(uint(buildingID))
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get devices success",
		"data":    devices,
	})
}

func (c *DeviceBuildingController) GetBuildingByDevice() {
	deviceIDStr := c.Ctx.Query("deviceId")
	deviceID, err := strconv.ParseUint(deviceIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid device ID"})
		return
	}

	building, err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).GetBuildingByDevice(uint(deviceID))
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data": building,
	})
}
