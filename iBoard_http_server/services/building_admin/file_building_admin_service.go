package building_admin_services

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceBuildingAdminFileService interface {
	Create(file *base_models.File, email string) error
	Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.File, base_models.PaginationResult, error)
	Update(id uint, email string, updates map[string]interface{}) error
	Delete(id uint, email string) error
	GetByID(id uint, email string) (*base_models.File, error)
}

type BuildingAdminFileService struct {
	db *gorm.DB
}

func NewBuildingAdminFileService(db *gorm.DB) InterfaceBuildingAdminFileService {
	return &BuildingAdminFileService{db: db}
}

func (s *BuildingAdminFileService) Create(file *base_models.File, email string) error {
	file.Uploader = email
	file.UploaderType = "building_admin"
	return s.db.Create(file).Error
}

func (s *BuildingAdminFileService) Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.File, base_models.PaginationResult, error) {
	var files []base_models.File
	var total int64

	// 首先获取 building admin 关联的所有 building IDs
	var buildingAdmin base_models.BuildingAdmin
	if err := s.db.Where("email = ?", email).Preload("Buildings").First(&buildingAdmin).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	var buildingIDs []uint
	for _, building := range buildingAdmin.Buildings {
		buildingIDs = append(buildingIDs, building.ID)
	}

	// 获取广告相关的文件 IDs
	var adFileIDs []uint
	adQuery := s.db.Table("advertisements").
		Select("DISTINCT file_id").
		Joins("JOIN advertisement_buildings ON advertisements.id = advertisement_buildings.advertisement_id").
		Where("advertisement_buildings.building_id IN ?", buildingIDs).
		Where("file_id IS NOT NULL")
	if err := adQuery.Pluck("file_id", &adFileIDs).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	// 获取通知相关的文件 IDs
	var noticeFileIDs []uint
	noticeQuery := s.db.Table("notices").
		Select("DISTINCT file_id").
		Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id IN ?", buildingIDs).
		Where("file_id IS NOT NULL")
	if err := noticeQuery.Pluck("file_id", &noticeFileIDs).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	// 合并所有文件 IDs
	allFileIDs := append(adFileIDs, noticeFileIDs...)

	// 构建主查询
	db := s.db.Model(&base_models.File{}).
		Where(
			s.db.Where("(uploader = ? AND uploader_type = ?)", email, "building_admin").
				Or("id IN ?", allFileIDs),
		)

	// 应用过滤条件
	if mimeType, ok := query["mimeType"].(string); ok && mimeType != "" {
		db = db.Where("mime_type = ?", mimeType)
	}

	if oss, ok := query["oss"].(string); ok && oss != "" {
		db = db.Where("oss = ?", oss)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	// 应用分页
	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if desc, ok := paginate["desc"].(bool); ok && desc {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	// Debug: 打印 SQL 语句
	db = db.Debug()

	if err := db.Limit(pageSize).Offset(offset).Find(&files).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return files, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingAdminFileService) Update(id uint, email string, updates map[string]interface{}) error {
	result := s.db.Model(&base_models.File{}).
		Where("id = ? AND uploader = ? AND uploader_type = ?", id, email, "building_admin").
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found or no permission")
	}
	return nil
}

func (s *BuildingAdminFileService) Delete(id uint, email string) error {
	result := s.db.Where("id = ? AND uploader = ? AND uploader_type = ?", id, email, "building_admin").
		Delete(&base_models.File{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found or no permission")
	}
	return nil
}

func (s *BuildingAdminFileService) GetByID(id uint, email string) (*base_models.File, error) {
	var file base_models.File

	// 首先获取 building admin 关联的所有 building IDs
	var buildingAdmin base_models.BuildingAdmin
	if err := s.db.Where("email = ?", email).Preload("Buildings").First(&buildingAdmin).Error; err != nil {
		return nil, err
	}

	var buildingIDs []uint
	for _, building := range buildingAdmin.Buildings {
		buildingIDs = append(buildingIDs, building.ID)
	}

	// 获取广告相关的文件 IDs
	var adFileIDs []uint
	adQuery := s.db.Table("advertisements").
		Select("DISTINCT file_id").
		Joins("JOIN advertisement_buildings ON advertisements.id = advertisement_buildings.advertisement_id").
		Where("advertisement_buildings.building_id IN ?", buildingIDs).
		Where("file_id IS NOT NULL")
	if err := adQuery.Pluck("file_id", &adFileIDs).Error; err != nil {
		return nil, err
	}

	// 获取通知相关的文件 IDs
	var noticeFileIDs []uint
	noticeQuery := s.db.Table("notices").
		Select("DISTINCT file_id").
		Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id IN ?", buildingIDs).
		Where("file_id IS NOT NULL")
	if err := noticeQuery.Pluck("file_id", &noticeFileIDs).Error; err != nil {
		return nil, err
	}

	// 合并所有文件 IDs
	allFileIDs := append(adFileIDs, noticeFileIDs...)

	// Debug: 打印 SQL 语句
	db := s.db.Debug().Model(&base_models.File{}).
		Where("id = ?", id).
		Where(
			s.db.Where("(uploader = ? AND uploader_type = ?)", email, "building_admin").
				Or("id IN ?", allFileIDs),
		)

	if err := db.First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("file not found or no permission")
		}
		return nil, err
	}

	return &file, nil
}
