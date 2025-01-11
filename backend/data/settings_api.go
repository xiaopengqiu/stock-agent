package data

import (
	"go-stock/backend/db"
	"gorm.io/gorm"
)

type Settings struct {
	gorm.Model
	LocalPushEnable bool   `json:"localPushEnable"`
	DingPushEnable  bool   `json:"dingPushEnable"`
	DingRobot       string `json:"dingRobot"`
}

func (receiver Settings) TableName() string {
	return "settings"
}

type SettingsApi struct {
	Config Settings
}

func NewSettingsApi(settings *Settings) *SettingsApi {
	return &SettingsApi{
		Config: *settings,
	}
}

func (s SettingsApi) UpdateConfig() string {
	err := db.Dao.Model(s.Config).Updates(map[string]any{
		"local_push_enable": s.Config.LocalPushEnable,
		"ding_push_enable":  s.Config.DingPushEnable,
		"ding_robot":        s.Config.DingRobot,
	}).Error
	if err != nil {
		return err.Error()
	}
	return "保存成功！"
}
func (s SettingsApi) GetConfig() *Settings {
	var settings Settings
	db.Dao.Model(&Settings{}).First(&settings)
	return &settings
}
