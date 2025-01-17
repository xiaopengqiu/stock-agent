package data

import (
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"gorm.io/gorm"
)

type Settings struct {
	gorm.Model
	LocalPushEnable        bool   `json:"localPushEnable"`
	DingPushEnable         bool   `json:"dingPushEnable"`
	DingRobot              string `json:"dingRobot"`
	UpdateBasicInfoOnStart bool   `json:"updateBasicInfoOnStart"`
	RefreshInterval        int64  `json:"refreshInterval"`

	OpenAiEnable      bool    `json:"openAiEnable"`
	OpenAiBaseUrl     string  `json:"openAiBaseUrl"`
	OpenAiApiKey      string  `json:"openAiApiKey"`
	OpenAiModelName   string  `json:"openAiModelName"`
	OpenAiMaxTokens   int     `json:"openAiMaxTokens"`
	OpenAiTemperature float64 `json:"openAiTemperature"`
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
	count := int64(0)
	db.Dao.Model(s.Config).Count(&count)
	if count > 0 {
		db.Dao.Model(s.Config).Where("id=?", s.Config.ID).Updates(map[string]any{
			"local_push_enable":          s.Config.LocalPushEnable,
			"ding_push_enable":           s.Config.DingPushEnable,
			"ding_robot":                 s.Config.DingRobot,
			"update_basic_info_on_start": s.Config.UpdateBasicInfoOnStart,
			"refresh_interval":           s.Config.RefreshInterval,
			"open_ai_enable":             s.Config.OpenAiEnable,
			"open_ai_base_url":           s.Config.OpenAiBaseUrl,
			"open_ai_api_key":            s.Config.OpenAiApiKey,
			"open_ai_model_name":         s.Config.OpenAiModelName,
			"open_ai_max_tokens":         s.Config.OpenAiMaxTokens,
			"open_ai_temperature":        s.Config.OpenAiTemperature,
		})
	} else {
		logger.SugaredLogger.Infof("未找到配置，创建默认配置:%+v", s.Config)
		db.Dao.Model(s.Config).Create(&Settings{
			LocalPushEnable:        s.Config.LocalPushEnable,
			DingPushEnable:         s.Config.DingPushEnable,
			DingRobot:              s.Config.DingRobot,
			UpdateBasicInfoOnStart: s.Config.UpdateBasicInfoOnStart,
			RefreshInterval:        s.Config.RefreshInterval,
			OpenAiEnable:           s.Config.OpenAiEnable,
			OpenAiBaseUrl:          s.Config.OpenAiBaseUrl,
			OpenAiApiKey:           s.Config.OpenAiApiKey,
			OpenAiModelName:        s.Config.OpenAiModelName,
			OpenAiMaxTokens:        s.Config.OpenAiMaxTokens,
			OpenAiTemperature:      s.Config.OpenAiTemperature,
		})
	}
	return "保存成功！"
}
func (s SettingsApi) GetConfig() *Settings {
	var settings Settings
	db.Dao.Model(&Settings{}).First(&settings)
	return &settings
}
