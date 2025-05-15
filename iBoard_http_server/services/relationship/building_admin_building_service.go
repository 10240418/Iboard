package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

// 自定义响应结构
type ServiceResponse struct {
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type InterfaceBuildingAdminBuildingService interface {
	BindBuildings(adminID uint, buildingIDs []uint) error
	UnbindBuildings(adminID uint, buildingIDs []uint) error
	GetBuildingsByAdminID(adminID uint) ([]base_models.Building, error)
	GetAdminsByBuildingID(buildingID uint) ([]base_models.BuildingAdmin, error)
	BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error)
	BuildingExists(buildingID uint) (bool, error)
	BuildingAdminExists(adminID uint) (bool, error)
	GetBuildingsByAdminEmail(email string) ([]base_models.Building, error)
}

type BuildingAdminBuildingService struct {
	db *gorm.DB
}

func NewBuildingAdminBuildingService(db *gorm.DB) InterfaceBuildingAdminBuildingService {
	return &BuildingAdminBuildingService{db: db}
}

func (s *BuildingAdminBuildingService) BindBuildings(adminID uint, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var admin base_models.BuildingAdmin
		if err := tx.Preload("Buildings").First(&admin, adminID).Error; err != nil {
			return err
		}

		var buildings []base_models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(&admin).Association("Buildings").Append(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *BuildingAdminBuildingService) UnbindBuildings(adminID uint, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var admin base_models.BuildingAdmin
		if err := tx.Preload("Buildings").First(&admin, adminID).Error; err != nil {
			return err
		}

		var buildings []base_models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(&admin).Association("Buildings").Delete(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *BuildingAdminBuildingService) GetBuildingsByAdminID(adminID uint) ([]base_models.Building, error) {
	var admin base_models.BuildingAdmin
	if err := s.db.Preload("Buildings").First(&admin, adminID).Error; err != nil {
		return nil, err
	}
	if admin.Buildings == nil {
		return []base_models.Building{}, nil
	}
	return admin.Buildings, nil
}

func (s *BuildingAdminBuildingService) GetAdminsByBuildingID(buildingID uint) ([]base_models.BuildingAdmin, error) {
	var building base_models.Building
	if err := s.db.
		Preload("BuildingAdmins").
		First(&building, buildingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []base_models.BuildingAdmin{}, nil
		}
		return nil, err
	}
	return building.BuildingAdmins, nil
}

func (s *BuildingAdminBuildingService) BuildingAdminExists(adminID uint) (bool, error) {
	var admin base_models.BuildingAdmin
	err := s.db.First(&admin, adminID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *BuildingAdminBuildingService) BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error) {
	var count int64
	err := s.db.Model(&base_models.Building{}).Where("id IN ?", buildingIDs).Count(&count).Error
	if err != nil {
		return nil, err
	}
	if int(count) == len(buildingIDs) {
		return []uint{}, nil
	}

	// 找出缺失的 Building IDs
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

func (s *BuildingAdminBuildingService) BuildingExists(buildingID uint) (bool, error) {
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

func (s *BuildingAdminBuildingService) GetBuildingsByAdminEmail(email string) ([]base_models.Building, error) {
	var admin base_models.BuildingAdmin
	if err := s.db.Where("email = ?", email).Preload("Buildings").First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []base_models.Building{}, nil
		}
		return nil, err
	}
	if admin.Buildings == nil {
		return []base_models.Building{}, nil
	}
	return admin.Buildings, nil
}
