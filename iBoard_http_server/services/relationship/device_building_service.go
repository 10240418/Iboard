package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceDeviceBuildingService interface {
	BindDevices(buildingID uint, deviceIDs []uint) error
	UnbindDevice(deviceID uint) error
	GetDevicesByBuilding(buildingID uint, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Device, base_models.PaginationResult, error)
	GetBuildingByDevice(deviceID uint) (*base_models.Building, error)
}

type DeviceBuildingService struct {
	db *gorm.DB
}

func NewDeviceBuildingService(db *gorm.DB) InterfaceDeviceBuildingService {
	return &DeviceBuildingService{db: db}
}

func (s *DeviceBuildingService) BindDevices(buildingID uint, deviceIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Check if building exists
		var building base_models.Building
		if err := tx.First(&building, buildingID).Error; err != nil {
			return errors.New("building not found")
		}

		// Check if all devices exist and are not bound to other buildings
		var devices []base_models.Device
		if err := tx.Where("id IN ?", deviceIDs).Find(&devices).Error; err != nil {
			return err
		}

		if len(devices) != len(deviceIDs) {
			return errors.New("one or more devices not found")
		}

		// Check if any device is already bound to another building
		for _, device := range devices {
			if device.BuildingID != 0 && device.BuildingID != buildingID {
				return errors.New("one or more devices are already bound to another building")
			}
		}

		// Update all devices' building ID
		if err := tx.Model(&base_models.Device{}).Where("id IN ?", deviceIDs).Update("building_id", buildingID).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *DeviceBuildingService) UnbindDevice(deviceID uint) error {
	// Check if device exists
	var device base_models.Device
	if err := s.db.First(&device, deviceID).Error; err != nil {
		return errors.New("device not found")
	}

	// Set building_id to null
	return s.db.Model(&device).Update("building_id", nil).Error
}

func (s *DeviceBuildingService) GetDevicesByBuilding(buildingID uint, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Device, base_models.PaginationResult, error) {
	var devices []base_models.Device
	var total int64
	db := s.db.Model(&base_models.Device{}).Where("building_id = ?", buildingID)

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("device_id LIKE ?", "%"+search+"%")
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

	if err := db.Limit(pageSize).Offset(offset).Find(&devices).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return devices, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *DeviceBuildingService) GetBuildingByDevice(deviceID uint) (*base_models.Building, error) {
	var device base_models.Device
	if err := s.db.Preload("Building").First(&device, deviceID).Error; err != nil {
		return nil, errors.New("device not found")
	}

	if device.BuildingID == 0 {
		return nil, errors.New("device is not bound to any building")
	}

	return &device.Building, nil
}
