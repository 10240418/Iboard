package base_models

// Building 建筑模型
type Building struct {
	ModelFields
	Name           string          `json:"name" gorm:"size:255;not null"`
	IsmartID       string          `json:"ismartId" gorm:"size:255;not null;unique"`
	Remark         string          `json:"remark" gorm:"type:text"`
	Location       string          `json:"location" gorm:"size:255"`
	Devices        []Device        `json:"devices" gorm:"foreignKey:BuildingID"`
	BuildingAdmins []BuildingAdmin `json:"-" gorm:"many2many:building_admins_buildings;"`
	Notices        []Notice        `json:"notices" gorm:"many2many:notice_buildings;"`
	Advertisements []Advertisement `json:"advertisements" gorm:"many2many:advertisement_buildings;"`
}
