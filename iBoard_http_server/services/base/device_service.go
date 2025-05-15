package base_services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"gorm.io/gorm"
)

// getDeviceHealthTimeout returns the device health timeout in seconds from environment variables
func getDeviceHealthTimeout() int {
	timeout := os.Getenv("DEVICE_HEALTH_TIMEOUT")
	if timeout == "" {
		return 600 // default to 1 hour if not set
	}

	timeoutInt, err := strconv.Atoi(timeout)
	if err != nil {
		return 600 // default to 1 hour if invalid value
	}

	return timeoutInt
}

type InterfaceDeviceService interface {
	Create(device *base_models.Device) error
	CreateMany(devices []*base_models.Device) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Device, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) (*base_models.Device, error)
	Delete(ids []uint) error
	GetByID(id uint) (*base_models.Device, error)
	GetByDeviceID(deviceID string) (*base_models.Device, error)
	GetDeviceAdvertisements(deviceId string) ([]base_models.Advertisement, error)
	GetDeviceNotices(deviceId string) ([]base_models.Notice, error)
	UpdateDeviceHealth(deviceID uint) error
	CheckDeviceStatus(deviceID uint) string
	GetWithStatus(query map[string]interface{}, pagination map[string]interface{}) ([]DeviceWithStatus, base_models.PaginationResult, error)
	GetByIDWithStatus(id uint) (*DeviceWithStatus, error)
	GetDevicesByBuildingWithStatus(buildingID uint) ([]DeviceWithStatus, error)
}

type DeviceService struct {
	db *gorm.DB
}

func NewDeviceService(db *gorm.DB) InterfaceDeviceService {
	return &DeviceService{
		db: db,
	}
}

func (s *DeviceService) Create(device *base_models.Device) error {
	// Verify building exists
	var building base_models.Building
	if err := s.db.First(&building, device.BuildingID).Error; err != nil {
		return errors.New("building not found")
	}
	return s.db.Create(device).Error
}

func (s *DeviceService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Device, base_models.PaginationResult, error) {
	var devices []base_models.Device
	var total int64
	db := s.db.Model(&base_models.Device{})

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("device_id LIKE ?", "%"+search+"%")
	}

	if buildingID, ok := query["buildingId"].(uint); ok && buildingID != 0 {
		db = db.Where("building_id = ?", buildingID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if desc, ok := paginate["desc"].(bool); ok && desc {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	// 选择所有字段，包括嵌入的设置字段
	if err := db.Preload("Building").
		Select("devices.*, " +
			"arrearage_update_duration, " +
			"notice_update_duration, " +
			"advertisement_update_duration, " +
			"advertisement_play_duration, " +
			"notice_play_duration, " +
			"spare_duration, " +
			"notice_stay_duration").
		Limit(pageSize).Offset(offset).
		Find(&devices).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return devices, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *DeviceService) Update(id uint, updates map[string]interface{}) (*base_models.Device, error) {
	var updatedDevice *base_models.Device

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 获取当前设备信息
		var device base_models.Device
		if err := tx.First(&device, id).Error; err != nil {
			return err
		}

		// 如果要更新 buildingId
		if newBuildingID, ok := updates["buildingId"].(float64); ok {
			// 如果新的 buildingId 不为 0，先验证建筑是否存在
			if newBuildingID != 0 {
				var building base_models.Building
				if err := tx.First(&building, uint(newBuildingID)).Error; err != nil {
					return fmt.Errorf("new building not found: %v", err)
				}

				// 检查设备是否已经绑定到其他建筑物
				if device.BuildingID != 0 && device.BuildingID != uint(newBuildingID) {
					// 先解绑
					if err := tx.Model(&device).Update("building_id", nil).Error; err != nil {
						return fmt.Errorf("failed to unbind old building: %v", err)
					}
				}

				// 绑定到新建筑物
				if err := tx.Model(&device).Update("building_id", uint(newBuildingID)).Error; err != nil {
					return fmt.Errorf("failed to bind to new building: %v", err)
				}
			} else {
				// 如果新的 buildingId 为 0，则只进行解绑
				if device.BuildingID != 0 {
					if err := tx.Model(&device).Update("building_id", nil).Error; err != nil {
						return fmt.Errorf("failed to unbind building: %v", err)
					}
				}
			}

			// 从更新映射中删除 buildingId，因为已经处理过了
			delete(updates, "buildingId")
		}

		// 更新设备基本信息
		if deviceID, ok := updates["device_id"]; ok {
			if err := tx.Model(&device).Update("device_id", deviceID).Error; err != nil {
				return err
			}
			delete(updates, "device_id")
		}

		// 更新设置字段
		settingsUpdates := make(map[string]interface{})
		settingsFields := []string{
			"arrearage_update_duration",
			"notice_update_duration",
			"advertisement_update_duration",
			"advertisement_play_duration",
			"notice_play_duration",
			"spare_duration",
			"notice_stay_duration",
		}

		for _, field := range settingsFields {
			if value, ok := updates[field]; ok {
				settingsUpdates[field] = value
				delete(updates, field)
			}
		}

		// 如果有设置字段需要更新
		if len(settingsUpdates) > 0 {
			if err := tx.Model(&device).Updates(settingsUpdates).Error; err != nil {
				return fmt.Errorf("failed to update settings: %v", err)
			}
		}

		// 获取更新后的设备信息，确保获取最新状态
		if err := tx.Preload("Building").First(&device, id).Error; err != nil {
			return err
		}

		updatedDevice = &device
		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedDevice, nil
}

func (s *DeviceService) Delete(ids []uint) error {
	result := s.db.Delete(&base_models.Device{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *DeviceService) GetByID(id uint) (*base_models.Device, error) {
	var device base_models.Device

	// 使用原始SQL查询获取设备信息
	if err := s.db.Raw(`
		SELECT 
			d.*,
			d.arrearage_update_duration,
			d.notice_update_duration,
			d.advertisement_update_duration,
			d.advertisement_play_duration,
			d.notice_play_duration,
			d.spare_duration,
			d.notice_stay_duration
		FROM devices d
		WHERE d.id = ?
	`, id).Scan(&device).Error; err != nil {
		return nil, err
	}

	// 加载关联的建筑信息
	if err := s.db.Preload("Building").First(&device, id).Error; err != nil {
		return nil, err
	}

	return &device, nil
}

func (s *DeviceService) GetByDeviceID(deviceID string) (*base_models.Device, error) {
	var device base_models.Device
	if err := s.db.Preload("Building").Where("device_id = ?", deviceID).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (s *DeviceService) CreateMany(devices []*base_models.Device) error {
	// Start a transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify all buildings exist first
		buildingIDs := make(map[uint]bool)
		buildingIDList := make([]uint, 0)
		for _, device := range devices {
			if !buildingIDs[device.BuildingID] {
				buildingIDs[device.BuildingID] = true
				buildingIDList = append(buildingIDList, device.BuildingID)
			}
		}

		var count int64
		if err := tx.Model(&base_models.Building{}).Where("id IN ?", buildingIDList).Count(&count).Error; err != nil {
			return err
		}
		if int(count) != len(buildingIDs) {
			return errors.New("one or more buildings not found")
		}

		// Verify device IDs are unique
		deviceIDs := make(map[string]bool)
		deviceIDList := make([]string, 0)
		for _, device := range devices {
			if deviceIDs[device.DeviceID] {
				return errors.New("duplicate device ID found")
			}
			deviceIDs[device.DeviceID] = true
			deviceIDList = append(deviceIDList, device.DeviceID)
		}

		// Check if any device IDs already exist in database
		var existingCount int64
		if err := tx.Model(&base_models.Device{}).Where("device_id IN ?", deviceIDList).Count(&existingCount).Error; err != nil {
			return err
		}
		if existingCount > 0 {
			return errors.New("one or more device IDs already exist")
		}

		// Create all devices
		for _, device := range devices {
			if err := tx.Create(device).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *DeviceService) GetDeviceAdvertisements(deviceId string) ([]base_models.Advertisement, error) {
	var device base_models.Device
	if err := s.db.Where("device_id = ?", deviceId).First(&device).Error; err != nil {
		return nil, fmt.Errorf("device not found: %v", err)
	}

	if device.BuildingID == 0 {
		return nil, fmt.Errorf("device is not bound to any building")
	}

	var advertisements []base_models.Advertisement
	if err := s.db.
		Joins("JOIN advertisement_buildings ON advertisements.id = advertisement_buildings.advertisement_id").
		Where("advertisement_buildings.building_id = ? AND advertisements.status = ?",
			device.BuildingID, "active").
		Preload("File").
		Find(&advertisements).Error; err != nil {
		return nil, fmt.Errorf("failed to get advertisements: %v", err)
	}

	// Clean up empty files and check endTime
	now := time.Now()
	for i := range advertisements {
		if advertisements[i].File != nil && advertisements[i].File.ID == 0 {
			advertisements[i].File = nil
		}

		// Check if advertisement has expired
		if advertisements[i].EndTime.Before(now) && advertisements[i].Status == field.Status("active") {
			// Update status in database
			if err := s.db.Model(&base_models.Advertisement{}).Where("id = ?", advertisements[i].ID).Update("status", field.Status("inactive")).Error; err != nil {
				return nil, fmt.Errorf("failed to update advertisement status: %v", err)
			}
			// Update status in memory
			advertisements[i].Status = field.Status("inactive")
		}
	}

	return advertisements, nil
}

func (s *DeviceService) GetDeviceNotices(deviceId string) ([]base_models.Notice, error) {
	var device base_models.Device
	if err := s.db.Where("device_id = ?", deviceId).First(&device).Error; err != nil {
		return nil, fmt.Errorf("device not found: %v", err)
	}

	if device.BuildingID == 0 {
		return nil, fmt.Errorf("device is not bound to any building")
	}

	var notices []base_models.Notice
	if err := s.db.
		Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id = ? AND notices.status = ?",
			device.BuildingID, "active").
		Select("notices.*, notices.is_ismart_notice as is_ismart_notice").
		Preload("File").
		Find(&notices).Error; err != nil {
		return nil, fmt.Errorf("failed to get notices: %v", err)
	}

	// Clean up empty files, set default values and check endTime
	now := time.Now()
	for i := range notices {
		if notices[i].FileType == "" {
			notices[i].FileType = field.FileTypePdf
		}
		if notices[i].File != nil && notices[i].File.ID == 0 {
			notices[i].File = nil
		}

		// Check if notice has expired
		if notices[i].EndTime.Before(now) && notices[i].Status == field.Status("active") {
			// Update status in database
			if err := s.db.Model(&base_models.Notice{}).Where("id = ?", notices[i].ID).Update("status", field.Status("inactive")).Error; err != nil {
				return nil, fmt.Errorf("failed to update notice status: %v", err)
			}
			// Update status in memory
			notices[i].Status = field.Status("inactive")
		}
	}

	return notices, nil
}

// DeviceWithStatus 用于返回带状态的设备信息
type DeviceWithStatus struct {
	base_models.Device
	Status string `json:"status"`
}

// UpdateDeviceHealth 更新设备健康状态
func (s *DeviceService) UpdateDeviceHealth(deviceID uint) error {
	key := fmt.Sprintf("device:online:%d", deviceID)
	ctx := context.Background()

	// 设置设备在线状态，使用环境变量中的超时时间
	timeout := getDeviceHealthTimeout()
	err := databases.REDIS_CONN.Set(ctx, key, "true", time.Duration(timeout)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to update device health: %v", err)
	}

	return nil
}

// CheckDeviceStatus 检查设备状态
func (s *DeviceService) CheckDeviceStatus(deviceID uint) string {
	key := fmt.Sprintf("device:online:%d", deviceID)
	ctx := context.Background()

	exists, err := databases.REDIS_CONN.Exists(ctx, key).Result()
	if err != nil || exists == 0 {
		return "inactive"
	}
	return "active"
}

// GetWithStatus 获取带状态的设备信息
func (s *DeviceService) GetWithStatus(query map[string]interface{}, pagination map[string]interface{}) ([]DeviceWithStatus, base_models.PaginationResult, error) {
	devices, paginationResult, err := s.Get(query, pagination)
	if err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	// 转换为带状态的设备信息
	devicesWithStatus := make([]DeviceWithStatus, len(devices))
	for i, device := range devices {
		devicesWithStatus[i] = DeviceWithStatus{
			Device: device,
			Status: s.CheckDeviceStatus(device.ID),
		}
	}

	return devicesWithStatus, paginationResult, nil
}

// GetByIDWithStatus 获取带状态的单个设备信息
func (s *DeviceService) GetByIDWithStatus(id uint) (*DeviceWithStatus, error) {
	device, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &DeviceWithStatus{
		Device: *device,
		Status: s.CheckDeviceStatus(device.ID),
	}, nil
}

// GetDevicesByBuildingWithStatus 获取建筑物关联的所有设备（带状态）
func (s *DeviceService) GetDevicesByBuildingWithStatus(buildingID uint) ([]DeviceWithStatus, error) {
	var devices []base_models.Device

	if err := databases.DB_CONN.
		Joins("JOIN device_buildings ON devices.id = device_buildings.device_id").
		Where("device_buildings.building_id = ?", buildingID).
		Find(&devices).Error; err != nil {
		return nil, err
	}

	// 转换为带状态的设备信息
	devicesWithStatus := make([]DeviceWithStatus, len(devices))
	for i, device := range devices {
		devicesWithStatus[i] = DeviceWithStatus{
			Device: device,
			Status: s.CheckDeviceStatus(device.ID),
		}
	}

	return devicesWithStatus, nil
}
