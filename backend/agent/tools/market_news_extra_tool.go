package tools

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/tidwall/gjson"
	"go-stock/backend/data"
	"strings"
)

// GetNewsList2Tool 获取新闻列表（带刷新）
func GetNewsList2Tool() tool.InvokableTool {
	return &NewsList2Tool{}
}

type NewsList2Tool struct {
}

func (t NewsList2Tool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetNewsList2",
		Desc: "获取新闻列表（会自动刷新最新电报数据），支持按来源筛选和数量限制",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"source": {
				Type:     "string",
				Desc:     "新闻来源，如'财联社电报'，留空则获取所有来源",
				Required: false,
			},
			"limit": {
				Type:     "integer",
				Desc:     "返回数量限制，默认50",
				Required: false,
			},
		}),
	}, nil
}

func (t NewsList2Tool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	source := gjson.Get(argumentsInJSON, "source").String()
	limit := int(gjson.Get(argumentsInJSON, "limit").Int())
	
	if limit <= 0 {
		limit = 50
	}

	news := data.NewMarketNewsApi().GetNewsList2(source, limit)
	
	md := strings.Builder{}
	md.WriteString("## 最新市场资讯\n\n")
	
	for _, item := range *news {
		md.WriteString("### " + item.Time + "\n")
		if item.Title != "" {
			md.WriteString("**" + item.Title + "**\n\n")
		}
		md.WriteString(item.Content + "\n\n")
		if len(item.SubjectTags) > 0 {
			md.WriteString("标签: " + strings.Join(item.SubjectTags, ", ") + "\n\n")
		}
		if item.Url != "" {
			md.WriteString("[查看原文](" + item.Url + ")\n\n")
		}
		md.WriteString("---\n\n")
	}
	
	return md.String(), nil
}

// GetTelegraphListWithPagingTool 分页获取电报列表
func GetTelegraphListWithPagingTool() tool.InvokableTool {
	return &TelegraphListWithPagingTool{}
}

type TelegraphListWithPagingTool struct {
}

func (t TelegraphListWithPagingTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetTelegraphListWithPaging",
		Desc: "分页获取电报列表，支持按来源筛选",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"source": {
				Type:     "string",
				Desc:     "新闻来源，如'财联社电报'，留空则获取所有来源",
				Required: false,
			},
			"page": {
				Type:     "integer",
				Desc:     "页码，从1开始，默认1",
				Required: false,
			},
			"pageSize": {
				Type:     "integer",
				Desc:     "每页数量，默认20",
				Required: false,
			},
		}),
	}, nil
}

func (t TelegraphListWithPagingTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	source := gjson.Get(argumentsInJSON, "source").String()
	page := int(gjson.Get(argumentsInJSON, "page").Int())
	pageSize := int(gjson.Get(argumentsInJSON, "pageSize").Int())
	
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	news := data.NewMarketNewsApi().GetTelegraphListWithPaging(source, page, pageSize)
	
	md := strings.Builder{}
	md.WriteString(fmt.Sprintf("## 电报列表 (第%d页，每页%d条)\n\n", page, pageSize))
	
	for _, item := range *news {
		md.WriteString("### " + item.Time + "\n")
		if item.Title != "" {
			md.WriteString("**" + item.Title + "**\n\n")
		}
		md.WriteString(item.Content + "\n\n")
		if len(item.SubjectTags) > 0 {
			md.WriteString("标签: " + strings.Join(item.SubjectTags, ", ") + "\n\n")
		}
		if item.Url != "" {
			md.WriteString("[查看原文](" + item.Url + ")\n\n")
		}
		md.WriteString("---\n\n")
	}
	
	return md.String(), nil
}

// GetTradingViewNewsDetailTool 获取TradingView新闻详情
func GetTradingViewNewsDetailTool() tool.InvokableTool {
	return &TradingViewNewsDetailTool{}
}

type TradingViewNewsDetailTool struct {
}

func (t TradingViewNewsDetailTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "TradingViewNewsDetail",
		Desc: "获取TradingView新闻的详细内容",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"id": {
				Type:     "string",
				Desc:     "新闻ID，从TradingViewNews列表中获取",
				Required: true,
			},
		}),
	}, nil
}

func (t TradingViewNewsDetailTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	id := gjson.Get(argumentsInJSON, "id").String()

	detail := data.NewMarketNewsApi().TradingViewNewsDetail(id)
	
	md := strings.Builder{}
	if detail.Title != "" {
		md.WriteString("# " + detail.Title + "\n\n")
	}
	if detail.ShortDescription != "" {
		md.WriteString(detail.ShortDescription + "\n\n")
	}
	if detail.Copyright != "" {
		md.WriteString("---\n\n*" + detail.Copyright + "*\n")
	}
	
	return md.String(), nil
}

// GetSecuritiesCompanyOpinionContentTool 获取券商观点内容
func GetSecuritiesCompanyOpinionContentTool() tool.InvokableTool {
	return &SecuritiesCompanyOpinionContentTool{}
}

type SecuritiesCompanyOpinionContentTool struct {
}

func (t SecuritiesCompanyOpinionContentTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetSecuritiesCompanyOpinionContent",
		Desc: "获取券商观点的详细内容",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"orgSName": {
				Type:     "string",
				Desc:     "券商名称",
				Required: true,
			},
			"encodeUrl": {
				Type:     "string",
				Desc:     "编码的URL，从券商观点列表中获取",
				Required: true,
			},
		}),
	}, nil
}

func (t SecuritiesCompanyOpinionContentTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	orgSName := gjson.Get(argumentsInJSON, "orgSName").String()
	encodeUrl := gjson.Get(argumentsInJSON, "encodeUrl").String()

	content := data.NewMarketNewsApi().GetSecuritiesCompanyOpinionContent(orgSName, encodeUrl)
	
	return content, nil
}

// GetNews24HoursListTool 获取24小时内的新闻列表
func GetNews24HoursListTool() tool.InvokableTool {
	return &News24HoursListTool{}
}

type News24HoursListTool struct {
}

func (t News24HoursListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "GetNews24HoursList",
		Desc: "获取过去24小时内的新闻列表，支持按来源筛选和数量限制，自动去重",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"source": {
				Type:     "string",
				Desc:     "新闻来源，如'财联社电报'，留空则获取所有来源",
				Required: false,
			},
			"limit": {
				Type:     "integer",
				Desc:     "返回数量限制，默认50",
				Required: false,
			},
		}),
	}, nil
}

func (t News24HoursListTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	source := gjson.Get(argumentsInJSON, "source").String()
	limit := int(gjson.Get(argumentsInJSON, "limit").Int())
	
	if limit <= 0 {
		limit = 50
	}

	news := data.NewMarketNewsApi().GetNews24HoursList(source, limit)
	
	md := strings.Builder{}
	md.WriteString("## 24小时内最新资讯\n\n")
	
	for _, item := range *news {
		md.WriteString("### " + item.Time + "\n")
		if item.Title != "" {
			md.WriteString("**" + item.Title + "**\n\n")
		}
		md.WriteString(item.Content + "\n\n")
		if len(item.SubjectTags) > 0 {
			md.WriteString("标签: " + strings.Join(item.SubjectTags, ", ") + "\n\n")
		}
		if item.Url != "" {
			md.WriteString("[查看原文](" + item.Url + ")\n\n")
		}
		md.WriteString("---\n\n")
	}
	
	return md.String(), nil
}
