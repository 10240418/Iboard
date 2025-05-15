package building_admin_services

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"gorm.io/gorm"
)

type InterfaceBuildingAdminAdvertisementService interface {
	Create(advertisement *base_models.Advertisement, email string) error
	Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error)
	Update(id uint, email string, updates map[string]interface{}) error
	Delete(id uint, email string) error
	GetByID(id uint, email string) (*base_models.Advertisement, error)
}

type BuildingAdminAdvertisementService struct {
	db                   *gorm.DB
	buildingAdminService relationship_service.InterfaceBuildingAdminBuildingService
	fileService          base_services.InterfaceFileService
}

func NewBuildingAdminAdvertisementService(
	db *gorm.DB,
	buildingAdminService relationship_service.InterfaceBuildingAdminBuildingService,
	fileService base_services.InterfaceFileService,
) InterfaceBuildingAdminAdvertisementService {
	return &BuildingAdminAdvertisementService{
		db:                   db,
		buildingAdminService: buildingAdminService,
		fileService:          fileService,
	}
}

func (s *BuildingAdminAdvertisementService) Create(advertisement *base_models.Advertisement, email string) error {
	// 获取管理员的建筑物
	buildings, err := s.buildingAdminService.GetBuildingsByAdminEmail(email)
	if err != nil {
		return err
	}

	// 验证文件是否存在且上传者类型是否正确
	var file base_models.File
	if err := s.db.First(&file, advertisement.FileID).Error; err != nil {
		return errors.New("file not found")
	}
	if file.UploaderType != field.UploaderTypeBuildingAdmin {
		return errors.New("file uploader type must be buildingAdmin")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 设置 isPublic 为 false
		advertisement.IsPublic = false

		if err := tx.Create(advertisement).Error; err != nil {
			return err
		}

		// 自动关联到管理员的所有建筑物
		if err := tx.Model(advertisement).Association("Buildings").Append(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *BuildingAdminAdvertisementService) Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error) {
	// 获取管理员的建筑物
	buildings, err := s.buildingAdminService.GetBuildingsByAdminEmail(email)
	if err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	var buildingIDs []uint
	for _, b := range buildings {
		buildingIDs = append(buildingIDs, b.ID)
	}

	var advertisements []base_models.Advertisement
	var total int64
	db := s.db.Model(&base_models.Advertisement{}).
		Joins("JOIN advertisement_buildings ON advertisements.id = advertisement_buildings.advertisement_id").
		Where("advertisement_buildings.building_id IN ?", buildingIDs).
		Group("advertisements.id")

	// 添加查询条件
	for key, value := range query {
		db = db.Where(key+" = ?", value)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if err := db.Preload("File").
		Limit(pageSize).
		Offset(offset).
		Find(&advertisements).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return advertisements, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingAdminAdvertisementService) Update(id uint, email string, updates map[string]interface{}) error {
	// 检查广告是否存在且属于该管理员
	advertisement, err := s.GetByID(id, email)
	if err != nil {
		return err
	}

	// 检查 isPublic
	if advertisement.IsPublic {
		return errors.New("cannot update public advertisement")
	}

	// 检查文件上传者类型
	var file base_models.File
	if err := s.db.First(&file, advertisement.FileID).Error; err != nil {
		return err
	}
	if file.UploaderType != field.UploaderTypeBuildingAdmin {
		return errors.New("cannot update advertisement with non-buildingAdmin file")
	}

	// 如果更新包含新的文件ID，检查新文件
	if newFileID, ok := updates["file_id"].(uint); ok {
		var newFile base_models.File
		if err := s.db.First(&newFile, newFileID).Error; err != nil {
			return errors.New("new file not found")
		}
		if newFile.UploaderType != field.UploaderTypeBuildingAdmin {
			return errors.New("new file uploader type must be buildingAdmin")
		}

		// 检查并可能删除旧文件
		if advertisement.FileID != nil {
			if err := s.checkAndDeleteFile(*advertisement.FileID); err != nil {
				return err
			}
		}
	}

	// 不允许更新 isPublic 为 true
	if isPublic, ok := updates["is_public"].(bool); ok && isPublic {
		return errors.New("cannot set isPublic to true")
	}

	return s.db.Model(&base_models.Advertisement{}).Where("id = ?", id).Updates(updates).Error
}

func (s *BuildingAdminAdvertisementService) Delete(id uint, email string) error {
	// 检查广告是否存在且属于该管理员
	advertisement, err := s.GetByID(id, email)
	if err != nil {
		return err
	}

	// 检查 isPublic
	if advertisement.IsPublic {
		return errors.New("cannot delete public advertisement")
	}

	// 检查文件上传者类型
	var file base_models.File
	if err := s.db.First(&file, advertisement.FileID).Error; err != nil {
		return err
	}
	if file.UploaderType != field.UploaderTypeBuildingAdmin {
		return errors.New("cannot delete advertisement with non-buildingAdmin file")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除广告
		if err := tx.Delete(&base_models.Advertisement{}, id).Error; err != nil {
			return err
		}

		// 检查并可能删除文件
		if advertisement.FileID != nil {
			if err := s.checkAndDeleteFile(*advertisement.FileID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *BuildingAdminAdvertisementService) GetByID(id uint, email string) (*base_models.Advertisement, error) {
	var advertisement base_models.Advertisement

	// 获取管理员的建筑物
	buildings, err := s.buildingAdminService.GetBuildingsByAdminEmail(email)
	if err != nil {
		return nil, err
	}

	var buildingIDs []uint
	for _, b := range buildings {
		buildingIDs = append(buildingIDs, b.ID)
	}

	err = s.db.Preload("File").
		Joins("JOIN advertisement_buildings ON advertisements.id = advertisement_buildings.advertisement_id").
		Where("advertisements.id = ? AND advertisement_buildings.building_id IN ?", id, buildingIDs).
		First(&advertisement).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("advertisement not found or no permission")
		}
		return nil, err
	}

	return &advertisement, nil
}

func (s *BuildingAdminAdvertisementService) checkAndDeleteFile(fileID uint) error {
	// 检查文件是否还被其他广告或通知使用
	var advertisementCount int64
	var noticeCount int64

	if err := s.db.Model(&base_models.Advertisement{}).Where("file_id = ?", fileID).Count(&advertisementCount).Error; err != nil {
		return err
	}

	if err := s.db.Model(&base_models.Notice{}).Where("file_id = ?", fileID).Count(&noticeCount).Error; err != nil {
		return err
	}

	// 如果文件没有其他引用，则删除文件
	if advertisementCount == 0 && noticeCount == 0 {
		if err := s.fileService.Delete([]uint{fileID}); err != nil {
			return err
		}
	}

	return nil
}
