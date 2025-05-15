package base_models

import (
	"time"

	"github.com/The-Healthist/iboard_http_service/utils/field"
)

// Notice 通知模型
type Notice struct {
	ModelFields
	Title          string           `json:"title"          gorm:"size:255;not null"`
	Description    string           `json:"description"    gorm:"type:text"`
	Type           field.NoticeType `json:"type"           gorm:"size:50"` //urgent , common ,system, government
	IsPublic       bool             `json:"isPublic"       gorm:"default:true"`
	IsIsmartNotice bool             `json:"isIsmartNotice" gorm:"default:false"` // use for sync with ismart notice
	Priority       int              `json:"priority"       gorm:"default:0"`     //0 - 100, 100 is the highest priority (default 0)
	Status         field.Status     `json:"status"         gorm:"size:50"`       // pending, active, inactive
	StartTime      time.Time        `json:"startTime"      gorm:"type:datetime"`
	EndTime        time.Time        `json:"endTime"        gorm:"type:datetime"`
	FileID         *uint            `json:"fileId"         gorm:"default:null"`
	File           *File            `json:"file,omitempty" gorm:"foreignKey:FileID"`
	FileType       field.FileType   `json:"fileType"       gorm:"size:50" default:"pdf"`
	Buildings      []Building       `json:"-"              gorm:"many2many:notice_buildings;"`
}
