package data

import (
	"github.com/go-resty/resty/v2"
	"go-stock/backend/logger"
)

// @Author spark
// @Date 2025/1/3 13:53
// @Desc
//-----------------------------------------------------------------------------------

const dingding_robot_url = "https://oapi.dingtalk.com/robot/send?access_token=0237527b404598f37ae5d83ef36e936860c7ba5d3892cd43f64c4159d3ed7cb1"

type DingDingAPI struct {
	client *resty.Client
}

func NewDingDingAPI() *DingDingAPI {
	return &DingDingAPI{
		client: resty.New(),
	}
}

func (DingDingAPI) SendDingDingMessage(message string) string {
	// 发送钉钉消息
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(message).
		Post(dingding_robot_url)
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return "发送钉钉消息失败"
	}
	logger.SugaredLogger.Infof("send dingding message: %s", resp.String())
	return "发送钉钉消息成功"
}

func (DingDingAPI) SendToDingDing(title, message string) string {
	// 发送钉钉消息
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(&Message{
			Msgtype: "markdown",
			Markdown: Markdown{
				Title: "go-stock " + title,
				Text:  message,
			},
			At: At{
				IsAtAll: true,
			},
		}).
		Post(dingding_robot_url)
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return "发送钉钉消息失败"
	}
	logger.SugaredLogger.Infof("send dingding message: %s", resp.String())
	return "发送钉钉消息成功"
}

type Message struct {
	Msgtype  string   `json:"msgtype"`
	Markdown Markdown `json:"markdown"`
	At       At       `json:"at"`
}

type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type At struct {
	AtMobiles []string `json:"atMobiles"`
	AtUserIds []string `json:"atUserIds"`
	IsAtAll   bool     `json:"isAtAll"`
}
