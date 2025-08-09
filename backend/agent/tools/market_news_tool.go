package tools

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/duke-git/lancet/v2/random"
	"github.com/tidwall/gjson"
	"go-stock/backend/data"
	"go-stock/backend/logger"
	"strings"
)

// @Author spark
// @Date 2025/8/4 16:38
// @Desc
//-----------------------------------------------------------------------------------

func GetQueryMarketNewsTool() tool.InvokableTool {
	return &QueryMarketNews{}
}

type QueryMarketNews struct {
}

func (q QueryMarketNews) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "QueryMarketNews",
		Desc: "国内外市场资讯/电报/会议/事件",
	}, nil
}

func (q QueryMarketNews) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	md := strings.Builder{}
	res := data.NewMarketNewsApi().ClsCalendar()
	for _, a := range res {
		bytes, err := json.Marshal(a)
		if err != nil {
			continue
		}
		//logger.SugaredLogger.Debugf("value: %+v", string(bytes))
		date := gjson.Get(string(bytes), "calendar_day")
		md.WriteString("\n### 事件/会议日期：" + date.String())
		list := gjson.Get(string(bytes), "items")
		//logger.SugaredLogger.Debugf("value: %+v,list: %+v", date.String(), list)
		list.ForEach(func(key, value gjson.Result) bool {
			logger.SugaredLogger.Debugf("key: %+v,value: %+v", key.String(), gjson.Get(value.String(), "title"))
			md.WriteString("\n- " + gjson.Get(value.String(), "title").String())
			return true
		})
	}

	news := data.NewMarketNewsApi().GetNewsList("财联社电报", random.RandInt(100, 500))
	messageText := strings.Builder{}
	for _, telegraph := range *news {
		messageText.WriteString("## " + telegraph.Time + ":" + "\n")
		messageText.WriteString("### " + telegraph.Content + "\n")
	}
	md.WriteString("\n### 市场资讯：\n" + messageText.String())

	resp := data.NewMarketNewsApi().TradingViewNews()
	var newsText strings.Builder
	for _, a := range *resp {
		logger.SugaredLogger.Debugf("TradingViewNews: %s", a.Title)
		newsText.WriteString(a.Title + "\n")
	}
	md.WriteString("\n### 全球新闻资讯：\n" + newsText.String())

	reutersNew := data.NewMarketNewsApi().ReutersNew()
	reutersNewMessageText := strings.Builder{}
	for _, article := range reutersNew.Result.Articles {
		reutersNewMessageText.WriteString("## " + article.Title + "\n")
		reutersNewMessageText.WriteString("### " + article.Description + "\n")
	}
	md.WriteString("\n### 外媒全球新闻资讯：\n" + reutersNewMessageText.String())

	return md.String(), nil
}
