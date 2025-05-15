package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceAdvertisementBuildingService interface {
	BindBuildings(advertisementID uint, buildingIDs []uint) error
	UnbindBuildings(advertisementID uint, buildingIDs []uint) error
	GetBuildingsByAdvertisementID(advertisementID uint) ([]base_models.Building, error)
	GetAdvertisementsByBuildingID(buildingID uint) ([]base_models.Advertisement, error)
	BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error)
	BuildingExists(buildingID uint) (bool, error)
	AdvertisementExists(advertisementID uint) (bool, error)
}

type AdvertisementBuildingService struct {
	db *gorm.DB
}

func NewAdvertisementBuildingService(db *gorm.DB) InterfaceAdvertisementBuildingService {
	return &AdvertisementBuildingService{db: db}
}

func (s *AdvertisementBuildingService) BindBuildings(advertisementID uint, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// check if the advertisement exists
		exists, err := s.AdvertisementExists(advertisementID)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("advertisement not found")
		}

		// check if all buildings exist
		missingIDs, err := s.BulkCheckBuildingsExistence(buildingIDs)
		if err != nil {
			return err
		}
		if len(missingIDs) > 0 {
			return errors.New("some buildings not found")
		}

		// check if the relationship already exists
		for _, buildingID := range buildingIDs {
			var count int64
			if err := tx.Table("advertisement_buildings").
				Where("advertisement_id = ? AND building_id = ?", advertisementID, buildingID).
				Count(&count).Error; err != nil {
				return err
			}

			// if not exists, insert
			if count == 0 {
				if err := tx.Exec(
					"INSERT INTO advertisement_buildings (advertisement_id, building_id) VALUES (?, ?)",
					advertisementID, buildingID,
				).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (s *AdvertisementBuildingService) UnbindBuildings(advertisementID uint, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var advertisement base_models.Advertisement
		if err := tx.Preload("Buildings").First(&advertisement, advertisementID).Error; err != nil {
			return err
		}

		var buildings []base_models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(&advertisement).Association("Buildings").Delete(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *AdvertisementBuildingService) GetBuildingsByAdvertisementID(advertisementID uint) ([]base_models.Building, error) {
	var advertisement base_models.Advertisement
	if err := s.db.Preload("Buildings").First(&advertisement, advertisementID).Error; err != nil {
		return nil, err
	}
	if advertisement.Buildings == nil {
		return []base_models.Building{}, nil
	}
	return advertisement.Buildings, nil
}

func (s *AdvertisementBuildingService) GetAdvertisementsByBuildingID(buildingID uint) ([]base_models.Advertisement, error) {
	var building base_models.Building
	if err := s.db.
		Preload("Advertisements.File").
		Preload("Advertisements").
		First(&building, buildingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []base_models.Advertisement{}, nil
		}
		return nil, err
	}
	return building.Advertisements, nil
}

func (s *AdvertisementBuildingService) AdvertisementExists(advertisementID uint) (bool, error) {
	var advertisement base_models.Advertisement
	err := s.db.First(&advertisement, advertisementID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *AdvertisementBuildingService) BuildingExists(buildingID uint) (bool, error) {
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

func (s *AdvertisementBuildingService) BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error) {
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
