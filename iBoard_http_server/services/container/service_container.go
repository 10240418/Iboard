package container

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"gorm.io/gorm"
)

// ServiceContainer holds all services
type ServiceContainer struct {
	db *gorm.DB

	// Base Services
	advertisementService base_services.InterfaceAdvertisementService
	buildingService      base_services.InterfaceBuildingService
	buildingAdminService base_services.InterfaceBuildingAdminService
	noticeService        base_services.InterfaceNoticeService
	fileService          base_services.InterfaceFileService
	jwtService           base_services.IJWTService
	emailService         base_services.IEmailService
	superAdminService    base_services.InterfaceSuperAdminService
	uploadService        base_services.IUploadService
	deviceService        base_services.InterfaceDeviceService
	noticeSyncService    base_services.InterfaceNoticeSyncService

	// Building Admin Services
	buildingAdminAdvertisementService building_admin_services.InterfaceBuildingAdminAdvertisementService
	buildingAdminNoticeService        building_admin_services.InterfaceBuildingAdminNoticeService
	buildingAdminFileService          building_admin_services.InterfaceBuildingAdminFileService

	// Relationship Services
	advertisementBuildingService relationship_service.InterfaceAdvertisementBuildingService
	noticeBuildingService        relationship_service.InterfaceNoticeBuildingService
	buildingAdminBuildingService relationship_service.InterfaceBuildingAdminBuildingService
	fileAdvertisementService     relationship_service.InterfaceFileAdvertisementService
	fileNoticeService            relationship_service.InterfaceFileNoticeService
	deviceBuildingService        relationship_service.InterfaceDeviceBuildingService

	mu sync.RWMutex
}

// NewServiceContainer creates a new service container
func NewServiceContainer(db *gorm.DB) *ServiceContainer {
	if db == nil {
		panic("database connection is nil")
	}

	// Verify Redis connection
	if databases.REDIS_CONN == nil {
		log.Println("Redis connection not initialized, attempting to initialize...")
		if err := databases.InitRedis(); err != nil {
			panic(fmt.Sprintf("failed to initialize redis: %v", err))
		}
	}

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := databases.REDIS_CONN.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("redis connection test failed: %v", err))
	}

	container := &ServiceContainer{
		db: db,
	}
	container.initializeServices()
	return container
}

// initializeServices initializes all services
func (c *ServiceContainer) initializeServices() {
	if c.db == nil {
		panic("database connection is nil")
	}

	// Ensure Redis is initialized
	if databases.REDIS_CONN == nil {
		if err := databases.InitRedis(); err != nil {
			panic(fmt.Sprintf("failed to initialize redis: %v", err))
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize Base Services
	c.jwtService = base_services.NewJWTService()
	c.fileService = base_services.NewFileService(c.db)
	c.buildingService = base_services.NewBuildingService(c.db)
	c.advertisementService = base_services.NewAdvertisementService(c.db)
	c.noticeService = base_services.NewNoticeService(c.db)
	c.buildingAdminService = base_services.NewBuildingAdminService(c.db)
	c.superAdminService = base_services.NewSuperAdminService(c.db)
	c.deviceService = base_services.NewDeviceService(c.db)

	// Use global Redis connection
	c.uploadService = base_services.NewUploadService(c.db, databases.REDIS_CONN)

	// Initialize Notice Sync Service
	c.noticeSyncService = base_services.NewNoticeSyncService(
		c.db,
		databases.REDIS_CONN,
		c.buildingService,
		c.uploadService,
		c.fileService,
	)

	// Initialize Relationship Services
	c.buildingAdminBuildingService = relationship_service.NewBuildingAdminBuildingService(c.db)
	c.advertisementBuildingService = relationship_service.NewAdvertisementBuildingService(c.db)
	c.noticeBuildingService = relationship_service.NewNoticeBuildingService(c.db)
	c.fileAdvertisementService = relationship_service.NewFileAdvertisementService(c.db)
	c.fileNoticeService = relationship_service.NewFileNoticeService(c.db)
	c.deviceBuildingService = relationship_service.NewDeviceBuildingService(c.db)

	// Initialize Building Admin Services
	c.buildingAdminAdvertisementService = building_admin_services.NewBuildingAdminAdvertisementService(
		c.db,
		c.buildingAdminBuildingService,
		c.fileService,
	)
	c.buildingAdminNoticeService = building_admin_services.NewBuildingAdminNoticeService(
		c.db,
		c.buildingAdminBuildingService,
		c.fileService,
	)
	c.buildingAdminFileService = building_admin_services.NewBuildingAdminFileService(c.db)
}

// GetService returns the requested service based on the service name
func (c *ServiceContainer) GetService(name string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	switch name {
	// Base services
	case "advertisement":
		return c.advertisementService
	case "building":
		return c.buildingService
	case "buildingAdmin":
		return c.buildingAdminService
	case "notice":
		return c.noticeService
	case "file":
		return c.fileService
	case "jwt":
		return c.jwtService
	case "superAdmin":
		return c.superAdminService
	case "email":
		return c.emailService
	case "upload":
		return c.uploadService
	case "device":
		return c.deviceService
	case "noticeSync":
		return c.noticeSyncService

	// Building admin services
	case "buildingAdminAdvertisement":
		return c.buildingAdminAdvertisementService
	case "buildingAdminNotice":
		return c.buildingAdminNoticeService
	case "buildingAdminFile":
		return c.buildingAdminFileService

	// Relationship services
	case "advertisementBuilding":
		return c.advertisementBuildingService
	case "noticeBuilding":
		return c.noticeBuildingService
	case "buildingAdminBuilding":
		return c.buildingAdminBuildingService
	case "fileAdvertisement":
		return c.fileAdvertisementService
	case "fileNotice":
		return c.fileNoticeService
	case "deviceBuilding":
		return c.deviceBuildingService
	default:
		return nil
	}
}
