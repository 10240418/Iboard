package base_models

// Device represents a display device in a building
type Device struct {
	ModelFields
	DeviceID   string         `json:"deviceId" gorm:"size:255;not null;unique"`
	Building   Building       `json:"building" gorm:"foreignKey:BuildingID"`
	BuildingID uint           `json:"buildingId" `
	Settings   DeviceSettings `json:"settings" gorm:"embedded"`
	Status     string         `json:"status" gorm:"-"` // 设备在线状态，不存储在数据库中
}

// DeviceSettings contains all settings for a device
type DeviceSettings struct {
	// Update durations (in minutes)
	ArrearageUpdateDuration     int `json:"arrearageUpdateDuration" gorm:"default:5"`      // 欠款更新时间间隔
	NoticeUpdateDuration        int `json:"noticeUpdateDuration" gorm:"default:10"`        // 通知更新时间间隔
	AdvertisementUpdateDuration int `json:"advertisementUpdateDuration" gorm:"default:15"` // 广告更新时间间隔

	// Play durations (in seconds)
	AdvertisementPlayDuration int `json:"advertisementPlayDuration" gorm:"default:30"` // 广告播放时间
	NoticePlayDuration        int `json:"noticePlayDuration" gorm:"default:30"`        // 通知播放时间
	SpareDuration             int `json:"spareDuration" gorm:"default:5"`              // 备用时间
	NoticeStayDuration        int `json:"noticeStayDuration" gorm:"default:10"`        // 通知停留时间
}
