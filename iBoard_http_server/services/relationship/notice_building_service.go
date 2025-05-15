package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceNoticeBuildingService interface {
	BindBuildings(noticeID uint, buildingIDs []uint) error
	UnbindBuildings(noticeID uint, buildingIDs []uint) error
	GetBuildingsByNoticeID(noticeID uint) ([]base_models.Building, error)
	GetNoticesByBuildingID(buildingID uint) ([]base_models.Notice, error)
	BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error)
	BuildingExists(buildingID uint) (bool, error)
	NoticeExists(noticeID uint) (bool, error)
}

type NoticeBuildingService struct {
	db *gorm.DB
}

func NewNoticeBuildingService(db *gorm.DB) InterfaceNoticeBuildingService {
	return &NoticeBuildingService{db: db}
}

func (s *NoticeBuildingService) BindBuildings(noticeID uint, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var notice base_models.Notice
		if err := tx.Preload("Buildings").First(&notice, noticeID).Error; err != nil {
			return err
		}

		var buildings []base_models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(&notice).Association("Buildings").Append(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *NoticeBuildingService) UnbindBuildings(noticeID uint, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var notice base_models.Notice
		if err := tx.Preload("Buildings").First(&notice, noticeID).Error; err != nil {
			return err
		}

		var buildings []base_models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(&notice).Association("Buildings").Delete(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *NoticeBuildingService) GetBuildingsByNoticeID(noticeID uint) ([]base_models.Building, error) {
	var notice base_models.Notice
	if err := s.db.Preload("Buildings").First(&notice, noticeID).Error; err != nil {
		return nil, err
	}
	if notice.Buildings == nil {
		return []base_models.Building{}, nil
	}
	return notice.Buildings, nil
}

func (s *NoticeBuildingService) GetNoticesByBuildingID(buildingID uint) ([]base_models.Notice, error) {
	var building base_models.Building
	if err := s.db.
		Preload("Notices.File").
		Preload("Notices").
		First(&building, buildingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []base_models.Notice{}, nil
		}
		return nil, err
	}
	return building.Notices, nil
}

func (s *NoticeBuildingService) NoticeExists(noticeID uint) (bool, error) {
	var notice base_models.Notice
	err := s.db.First(&notice, noticeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *NoticeBuildingService) BuildingExists(buildingID uint) (bool, error) {
	var building base_models.Building
	err := s.db.First(&building, buildingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *NoticeBuildingService) BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error) {
	var count int64
	err := s.db.Model(&base_models.Building{}).Where("id IN ?", buildingIDs).Count(&count).Error
	if err != nil {
		return nil, err
	}
	if int(count) == len(buildingIDs) {
		return []uint{}, nil
	}

	var existingIDs []uint
	err = s.db.Model(&base_models.Building{}).Where("id IN ?", buildingIDs).Pluck("id", &existingIDs).Error
	if err != nil {
		return nil, err
	}

	existingMap := make(map[uint]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	var missingIDs []uint
	for _, id := range buildingIDs {
		if !existingMap[id] {
			missingIDs = append(missingIDs, id)
		}
	}

	return missingIDs, nil
}
