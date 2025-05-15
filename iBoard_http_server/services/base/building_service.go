package base_services

import (
	"errors"
	"fmt"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"gorm.io/gorm"
)

type InterfaceBuildingService interface {
	Create(building *base_models.Building) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Building, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(ids []uint) error
	GetByID(id uint) (*base_models.Building, error)
	GetByIDWithDevices(id uint) (*base_models.Building, error)
	GetByIsmartID(ismartID string) (*base_models.Building, error)
	GetBuildingAdvertisements(buildingId uint) ([]base_models.Advertisement, error)
	GetBuildingNotices(buildingId uint) ([]base_models.Notice, error)
}

type BuildingService struct {
	db *gorm.DB
}

func NewBuildingService(db *gorm.DB) InterfaceBuildingService {
	return &BuildingService{db: db}
}

func (s *BuildingService) Create(building *base_models.Building) error {
	return s.db.Create(building).Error
}

func (s *BuildingService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Building, base_models.PaginationResult, error) {
	var buildings []base_models.Building
	var total int64
	db := s.db.Model(&base_models.Building{})

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("name LIKE ? OR ismart_id LIKE ? OR remark LIKE ? OR location LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
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

	if err := db.Select("id, created_at, updated_at, deleted_at, name, ismart_id, remark, location").
		Limit(pageSize).Offset(offset).
		Find(&buildings).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return buildings, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingService) Update(id uint, updates map[string]interface{}) error {
	result := s.db.Model(&base_models.Building{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("building not found")
	}
	return nil
}

func (s *BuildingService) Delete(ids []uint) error {
	result := s.db.Delete(&base_models.Building{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *BuildingService) GetByID(id uint) (*base_models.Building, error) {
	var building base_models.Building

	// 使用 Debug 模式查看 SQL 查询
	db := s.db.Debug()

	// 使用一次查询加载所有关联数据
	if err := db.
		Preload("Devices", func(db *gorm.DB) *gorm.DB {
			return db.Select("devices.*, " +
				"devices.id, " +
				"devices.created_at, " +
				"devices.updated_at, " +
				"devices.deleted_at, " +
				"devices.device_id, " +
				"devices.building_id, " +
				"devices.arrearage_update_duration, " +
				"devices.notice_update_duration, " +
				"devices.advertisement_update_duration, " +
				"devices.advertisement_play_duration, " +
				"devices.notice_play_duration, " +
				"devices.spare_duration, " +
				"devices.notice_stay_duration")
		}).
		Preload("Notices.File").
		Preload("Advertisements.File").
		Where("id = ?", id).
		First(&building).Error; err != nil {
		return nil, err
	}

	// 初始化空切片
	if building.BuildingAdmins == nil {
		building.BuildingAdmins = []base_models.BuildingAdmin{}
	}
	if building.Notices == nil {
		building.Notices = []base_models.Notice{}
	}
	if building.Advertisements == nil {
		building.Advertisements = []base_models.Advertisement{}
	}
	if building.Devices == nil {
		building.Devices = []base_models.Device{}
	}

	// 获取每个设备的在线状态
	deviceService := NewDeviceService(s.db)
	for i := range building.Devices {
		status := deviceService.CheckDeviceStatus(building.Devices[i].ID)
		building.Devices[i].Status = status
		fmt.Printf("Device %d: ID=%d, DeviceID=%s, Status=%s\n",
			i, building.Devices[i].ID, building.Devices[i].DeviceID, status)
	}

	// 设置通知默认值
	for i := range building.Notices {
		if building.Notices[i].StartTime.IsZero() {
			building.Notices[i].StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Notices[i].EndTime.IsZero() {
			building.Notices[i].EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Notices[i].Status == "" {
			building.Notices[i].Status = field.Status("active")
		}
		if building.Notices[i].FileType == "" {
			building.Notices[i].FileType = field.FileTypePdf
		}
		if building.Notices[i].File != nil && building.Notices[i].File.ID == 0 {
			building.Notices[i].File = nil
		}
	}

	// 设置广告默认值
	for i := range building.Advertisements {
		if building.Advertisements[i].StartTime.IsZero() {
			building.Advertisements[i].StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Advertisements[i].EndTime.IsZero() {
			building.Advertisements[i].EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Advertisements[i].Status == "" {
			building.Advertisements[i].Status = field.Status("active")
		}
		if building.Advertisements[i].File != nil && building.Advertisements[i].File.ID == 0 {
			building.Advertisements[i].File = nil
		}
	}

	return &building, nil
}

func (s *BuildingService) GetByIDWithDevices(id uint) (*base_models.Building, error) {
	var building base_models.Building

	// 使用 Debug 模式查看 SQL 查询
	db := s.db.Debug()

	// 使用一次查询加载所有关联数据
	if err := db.
		Preload("Devices").
		Preload("Notices").
		Preload("Notices.File").
		Preload("Advertisements").
		Preload("Advertisements.File").
		Where("id = ?", id).
		First(&building).Error; err != nil {
		return nil, err
	}

	// 初始化空切片
	if building.BuildingAdmins == nil {
		building.BuildingAdmins = []base_models.BuildingAdmin{}
	}
	if building.Notices == nil {
		building.Notices = []base_models.Notice{}
	}
	if building.Advertisements == nil {
		building.Advertisements = []base_models.Advertisement{}
	}
	if building.Devices == nil {
		building.Devices = []base_models.Device{}
	}

	// 获取每个设备的在线状态
	deviceService := NewDeviceService(s.db)
	for i := range building.Devices {
		status := deviceService.CheckDeviceStatus(building.Devices[i].ID)
		building.Devices[i].Status = status
	}

	// 设置通知默认值
	for i := range building.Notices {
		if building.Notices[i].StartTime.IsZero() {
			building.Notices[i].StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Notices[i].EndTime.IsZero() {
			building.Notices[i].EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Notices[i].Status == "" {
			building.Notices[i].Status = field.Status("active")
		}
		if building.Notices[i].FileType == "" {
			building.Notices[i].FileType = field.FileTypePdf
		}
		if building.Notices[i].File != nil && building.Notices[i].File.ID == 0 {
			building.Notices[i].File = nil
		}
	}

	// 设置广告默认值
	for i := range building.Advertisements {
		if building.Advertisements[i].StartTime.IsZero() {
			building.Advertisements[i].StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Advertisements[i].EndTime.IsZero() {
			building.Advertisements[i].EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Advertisements[i].Status == "" {
			building.Advertisements[i].Status = field.Status("active")
		}
		if building.Advertisements[i].File != nil && building.Advertisements[i].File.ID == 0 {
			building.Advertisements[i].File = nil
		}
	}

	return &building, nil
}

func (s *BuildingService) GetByIsmartID(ismartID string) (*base_models.Building, error) {
	var building base_models.Building
	if err := s.db.Where("ismart_id = ?", ismartID).
		Preload("Notices").
		Preload("Notices.File").
		Preload("Advertisements", "status = ?", "active").
		Preload("Advertisements.File").
		First(&building).Error; err != nil {
		return nil, fmt.Errorf("invalid ismartId: %v", err)
	}

	if building.Notices == nil {
		building.Notices = []base_models.Notice{}
	}
	if building.Advertisements == nil {
		building.Advertisements = []base_models.Advertisement{}
	}

	for i := range building.Notices {
		if building.Notices[i].File != nil && building.Notices[i].File.ID == 0 && building.Notices[i].FileID != nil {
			var file base_models.File
			if err := s.db.First(&file, building.Notices[i].FileID).Error; err == nil {
				building.Notices[i].File = &file
			} else {
				building.Notices[i].File = nil
			}
		}
	}

	for i := range building.Advertisements {
		if building.Advertisements[i].File != nil && building.Advertisements[i].File.ID == 0 && building.Advertisements[i].FileID != nil {
			var file base_models.File
			if err := s.db.First(&file, building.Advertisements[i].FileID).Error; err == nil {
				building.Advertisements[i].File = &file
			} else {
				building.Advertisements[i].File = nil
			}
		}
	}
	return &building, nil
}

func (s *BuildingService) GetBuildingAdvertisements(buildingId uint) ([]base_models.Advertisement, error) {
	var building base_models.Building
	if err := s.db.
		Preload("Advertisements", "is_public = ?", true).
		Preload("Advertisements.File").
		First(&building, buildingId).Error; err != nil {
		return nil, fmt.Errorf("building not found: %v", err)
	}

	if building.Advertisements == nil {
		building.Advertisements = []base_models.Advertisement{}
	}

	// Set default values and handle File field
	for i := range building.Advertisements {
		if building.Advertisements[i].StartTime.IsZero() {
			building.Advertisements[i].StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Advertisements[i].EndTime.IsZero() {
			building.Advertisements[i].EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Advertisements[i].Status == "" {
			building.Advertisements[i].Status = field.Status("active")
		}
		// Handle File field
		if building.Advertisements[i].FileID != nil && building.Advertisements[i].File != nil && building.Advertisements[i].File.ID == 0 {
			var file base_models.File
			if err := s.db.First(&file, building.Advertisements[i].FileID).Error; err == nil {
				building.Advertisements[i].File = &file
			} else {
				building.Advertisements[i].File = nil
			}
		} else if building.Advertisements[i].FileID == nil {
			building.Advertisements[i].File = nil
		}
	}

	return building.Advertisements, nil
}

func (s *BuildingService) GetBuildingNotices(buildingId uint) ([]base_models.Notice, error) {
	var building base_models.Building
	if err := s.db.
		Preload("Notices", "is_public = ?", true).
		Preload("Notices.File").
		First(&building, buildingId).Error; err != nil {
		return nil, fmt.Errorf("building not found: %v", err)
	}

	if building.Notices == nil {
		building.Notices = []base_models.Notice{}
	}

	// Set default values and handle File field
	for i := range building.Notices {
		if building.Notices[i].StartTime.IsZero() {
			building.Notices[i].StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Notices[i].EndTime.IsZero() {
			building.Notices[i].EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if building.Notices[i].Status == "" {
			building.Notices[i].Status = field.Status("active")
		}
		if building.Notices[i].FileType == "" {
			building.Notices[i].FileType = field.FileTypePdf
		}
		// Handle File field
		if building.Notices[i].FileID != nil && building.Notices[i].File != nil && building.Notices[i].File.ID == 0 {
			var file base_models.File
			if err := s.db.First(&file, building.Notices[i].FileID).Error; err == nil {
				building.Notices[i].File = &file
			} else {
				building.Notices[i].File = nil
			}
		} else if building.Notices[i].FileID == nil {
			building.Notices[i].File = nil
		}
	}

	return building.Notices, nil
}
