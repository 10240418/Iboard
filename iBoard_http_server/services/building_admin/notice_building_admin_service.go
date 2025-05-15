package building_admin_services

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"gorm.io/gorm"
)

type InterfaceBuildingAdminNoticeService interface {
	Create(notice *base_models.Notice, email string) error
	Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Notice, base_models.PaginationResult, error)
	Update(id uint, email string, updates map[string]interface{}) error
	Delete(id uint, email string) error
	GetByID(id uint, email string) (*base_models.Notice, error)
}

type BuildingAdminNoticeService struct {
	db                   *gorm.DB
	buildingAdminService relationship_service.InterfaceBuildingAdminBuildingService
	fileService          base_services.InterfaceFileService
}

func NewBuildingAdminNoticeService(
	db *gorm.DB,
	buildingAdminService relationship_service.InterfaceBuildingAdminBuildingService,
	fileService base_services.InterfaceFileService,
) InterfaceBuildingAdminNoticeService {
	return &BuildingAdminNoticeService{
		db:                   db,
		buildingAdminService: buildingAdminService,
		fileService:          fileService,
	}
}

func (s *BuildingAdminNoticeService) Create(notice *base_models.Notice, email string) error {
	// 获取管理员的建筑物
	buildings, err := s.buildingAdminService.GetBuildingsByAdminEmail(email)
	if err != nil {
		return err
	}

	// 验证文件是否存在且上传者类型是否正确
	if notice.FileID != nil {
		var file base_models.File
		if err := s.db.First(&file, notice.FileID).Error; err != nil {
			return errors.New("file not found")
		}
		if file.UploaderType != field.UploaderTypeBuildingAdmin {
			return errors.New("file uploader type must be buildingAdmin")
		}
	}

	// 设置 IsPublic 为 false
	notice.IsPublic = false

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(notice).Error; err != nil {
			return err
		}

		// 自动关联到管理员的所有建筑物
		if err := tx.Model(notice).Association("Buildings").Append(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *BuildingAdminNoticeService) Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Notice, base_models.PaginationResult, error) {
	// 获取管理员的建筑物
	buildings, err := s.buildingAdminService.GetBuildingsByAdminEmail(email)
	if err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	var buildingIDs []uint
	for _, b := range buildings {
		buildingIDs = append(buildingIDs, b.ID)
	}

	var notices []base_models.Notice
	var total int64
	db := s.db.Model(&base_models.Notice{}).
		Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id IN ?", buildingIDs).
		Group("notices.id")

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
		Find(&notices).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return notices, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingAdminNoticeService) Update(id uint, email string, updates map[string]interface{}) error {
	// 检查通知是否存在且属于该管理员
	notice, err := s.GetByID(id, email)
	if err != nil {
		return err
	}

	// 检查 isPublic
	if notice.IsPublic {
		return errors.New("cannot update public notice")
	}

	// 检查文件上传者类型
	if notice.FileID != nil {
		var file base_models.File
		if err := s.db.First(&file, notice.FileID).Error; err != nil {
			return err
		}
		if file.UploaderType != field.UploaderTypeBuildingAdmin {
			return errors.New("cannot update notice with non-buildingAdmin file")
		}
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
		if notice.FileID != nil {
			if err := s.checkAndDeleteFile(*notice.FileID); err != nil {
				return err
			}
		}
	}

	// 不允许更新 isPublic 为 true
	if isPublic, ok := updates["is_public"].(bool); ok && isPublic {
		return errors.New("cannot set isPublic to true")
	}

	return s.db.Model(&base_models.Notice{}).Where("id = ?", id).Updates(updates).Error
}

func (s *BuildingAdminNoticeService) Delete(id uint, email string) error {
	// 检查通知是否存在且属于该管理员
	notice, err := s.GetByID(id, email)
	if err != nil {
		return err
	}

	// 检查 isPublic
	if notice.IsPublic {
		return errors.New("cannot delete public notice")
	}

	// 检查文件上传者类型
	if notice.FileID != nil {
		var file base_models.File
		if err := s.db.First(&file, notice.FileID).Error; err != nil {
			return err
		}
		if file.UploaderType != field.UploaderTypeBuildingAdmin {
			return errors.New("cannot delete notice with non-buildingAdmin file")
		}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除通知
		if err := tx.Delete(&base_models.Notice{}, id).Error; err != nil {
			return err
		}

		// 检查并可能删除文件
		if notice.FileID != nil {
			if err := s.checkAndDeleteFile(*notice.FileID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *BuildingAdminNoticeService) GetByID(id uint, email string) (*base_models.Notice, error) {
	var notice base_models.Notice

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
		Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notices.id = ? AND notice_buildings.building_id IN ?", id, buildingIDs).
		First(&notice).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notice not found or no permission")
		}
		return nil, err
	}

	return &notice, nil
}

func (s *BuildingAdminNoticeService) checkAndDeleteFile(fileID uint) error {
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
