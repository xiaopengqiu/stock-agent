//go:build windows

package data

import (
	"github.com/go-toast/toast"
	"go-stock/backend/logger"
)

// AlertWindowsApi @Author spark
// @Date 2025/1/8 9:40
// @Desc
// -----------------------------------------------------------------------------------
type alertWindowsApi struct {
	AppID string
	// 窗口标题
	Title string
	// 窗口内容
	Content string
	// 窗口图标
	Icon string
}

func NewAlertWindowsApi(AppID string, Title string, Content string, Icon string) *alertWindowsApi {
	return &alertWindowsApi{
		AppID:   AppID,
		Title:   Title,
		Content: Content,
		Icon:    Icon,
	}
}

func (a alertWindowsApi) SendNotification() bool {
	notification := toast.Notification{
		AppID:    a.AppID,
		Title:    a.Title,
		Message:  a.Content,
		Icon:     a.Icon,
		Duration: "short",
		Audio:    toast.Default,
	}
	err := notification.Push()
	if err != nil {
		logger.SugaredLogger.Error(err)
		return false
	}
	return true
}
