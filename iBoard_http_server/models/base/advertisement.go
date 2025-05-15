package base_models

import (
	"time"

	"github.com/The-Healthist/iboard_http_service/utils/field"
)

// Advertisement 广告模型
type Advertisement struct {
	ModelFields
	Title       string                     `json:"title"          gorm:"size:255;not null"`
	Description string                     `json:"description"    gorm:"type:text"`
	Type        field.AdvertisementType    `json:"type"           gorm:"size:50"` // video, image
	Status      field.Status               `json:"status"         gorm:"size:50"` // pending, active, inactive
	Duration    int                        `json:"duration"`
	Priority    int                        `json:"priority"       gorm:"default:0"` //0 - 100, 100 is the highest priority (default 0)
	StartTime   time.Time                  `json:"startTime"      gorm:"type:datetime"`
	EndTime     time.Time                  `json:"endTime"        gorm:"type:datetime"`
	Display     field.AdvertisementDisplay `json:"display"        gorm:"size:50"` // full, top, topfull
	FileID      *uint                      `json:"fileId"         gorm:"default:null"`
	File        *File                      `json:"file,omitempty" gorm:"foreignKey:FileID"`
	IsPublic    bool                       `json:"isPublic"       gorm:"default:true"`
	Buildings   []Building                 `json:"-"              gorm:"many2many:advertisement_buildings;"`
}
