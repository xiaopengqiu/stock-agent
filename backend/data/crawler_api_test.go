package data

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/duke-git/lancet/v2/strutil"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func TestNewTimeOutGuShiTongCrawler(t *testing.T) {
	crawlerAPI := CrawlerApi{}
	timeout := 10
	crawlerBaseInfo := CrawlerBaseInfo{
		Name:        "TestCrawler",
		Description: "Test Crawler Description",
		BaseUrl:     "https://gushitong.baidu.com",
		Headers:     map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0"},
	}

	result := crawlerAPI.NewTimeOutCrawler(timeout, crawlerBaseInfo)
	assert.NotNil(t, result.crawlerCtx)
	assert.Equal(t, crawlerBaseInfo, result.crawlerBaseInfo)
}

func TestNewGuShiTongCrawler(t *testing.T) {
	crawlerAPI := CrawlerApi{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	crawlerBaseInfo := CrawlerBaseInfo{
		Name:        "TestCrawler",
		Description: "Test Crawler Description",
		BaseUrl:     "https://gushitong.baidu.com",
		Headers:     map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0"},
	}

	result := crawlerAPI.NewCrawler(ctx, crawlerBaseInfo)
	assert.Equal(t, ctx, result.crawlerCtx)
	assert.Equal(t, crawlerBaseInfo, result.crawlerBaseInfo)
}

func TestGetHtml(t *testing.T) {
	crawlerAPI := CrawlerApi{}
	crawlerBaseInfo := CrawlerBaseInfo{
		Name:        "TestCrawler",
		Description: "Test Crawler Description",
		BaseUrl:     "https://gushitong.baidu.com",
		Headers:     map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	crawlerAPI = crawlerAPI.NewCrawler(ctx, crawlerBaseInfo)

	url := "https://www.cls.cn/searchPage?type=depth&keyword=%E6%96%B0%E5%B8%8C%E6%9C%9B"
	waitVisible := ".search-telegraph-list,.subject-interest-list"

	//url = "https://gushitong.baidu.com/stock/ab-600745"
	//waitVisible = "div.news-item"
	htmlContent, success := crawlerAPI.GetHtml(url, waitVisible, true)
	if success {
		document, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			logger.SugaredLogger.Error(err.Error())
		}
		var messages []string
		document.Find(waitVisible).Each(func(i int, selection *goquery.Selection) {
			text := strutil.RemoveNonPrintable(selection.Text())
			messages = append(messages, text)
			logger.SugaredLogger.Infof("搜索到消息-%s: %s", "", text)
		})
	}
	//logger.SugaredLogger.Infof("htmlContent:%s", htmlContent)
}

func TestGetHtmlWithActions(t *testing.T) {
	crawlerAPI := CrawlerApi{}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	crawlerAPI = crawlerAPI.NewCrawler(ctx, CrawlerBaseInfo{
		Name:        "百度股市通",
		Description: "Test Crawler Description",
		BaseUrl:     "https://gushitong.baidu.com",
		Headers:     map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0"},
	})
	actions := []chromedp.Action{
		chromedp.Navigate("https://gushitong.baidu.com/stock/ab-600745"),
		chromedp.WaitVisible("div.cos-tab"),
		chromedp.Click(".header div.cos-tab:nth-child(6)", chromedp.ByQuery),
		chromedp.ScrollIntoView("div.finance-container >div.row:nth-child(3)"),
		chromedp.WaitVisible("div.cos-tabs-header-container"),
		chromedp.Click(".page-content .cos-tabs-header-container .cos-tabs-header .cos-tab:nth-child(1)", chromedp.ByQuery),
		chromedp.WaitVisible(".page-content .finance-container .report-col-content", chromedp.ByQuery),
		chromedp.Click(".page-content .cos-tabs-header-container .cos-tabs-header .cos-tab:nth-child(4)", chromedp.ByQuery),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(1 * time.Second),
	}
	htmlContent, success := crawlerAPI.GetHtmlWithActions(&actions, false)
	if success {
		document, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			logger.SugaredLogger.Error(err.Error())
		}
		var messages []string
		document.Find("div.report-table-list-container,div.report-row").Each(func(i int, selection *goquery.Selection) {
			text := strutil.RemoveWhiteSpace(selection.Text(), false)
			messages = append(messages, text)
			logger.SugaredLogger.Infof("搜索到消息-%s: %s", "", text)
		})
		logger.SugaredLogger.Infof("messages:%d", len(messages))
	}
	//logger.SugaredLogger.Infof("htmlContent:%s", htmlContent)
}

func TestHk(t *testing.T) {
	//https://stock.finance.sina.com.cn/hkstock/quotes/00001.html
	db.Init("../../data/stock.db")
	hks := &[]models.StockInfoHK{}
	db.Dao.Model(&models.StockInfoHK{}).Limit(1).Find(hks)

	crawlerAPI := CrawlerApi{}
	crawlerBaseInfo := CrawlerBaseInfo{
		Name:        "TestCrawler",
		Description: "Test Crawler Description",
		BaseUrl:     "https://stock.finance.sina.com.cn",
		Headers:     map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()
	crawlerAPI = crawlerAPI.NewCrawler(ctx, crawlerBaseInfo)

	for _, hk := range *hks {
		logger.SugaredLogger.Infof("hk: %+v", hk)
		url := fmt.Sprintf("https://stock.finance.sina.com.cn/hkstock/quotes/%s.html", strings.ReplaceAll(hk.Code, ".HK", ""))
		htmlContent, ok := crawlerAPI.GetHtml(url, "#stock_cname", true)
		if !ok {
			continue
		}
		//logger.SugaredLogger.Infof("htmlContent: %s", htmlContent)
		document, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			logger.SugaredLogger.Error(err.Error())
		}
		document.Find("#stock_cname").Each(func(i int, selection *goquery.Selection) {
			text := strutil.RemoveNonPrintable(selection.Text())
			logger.SugaredLogger.Infof("股票名称-:%s", text)
		})

		document.Find("#mts_stock_hk_price").Each(func(i int, selection *goquery.Selection) {
			text := strutil.RemoveNonPrintable(selection.Text())
			logger.SugaredLogger.Infof("股票名称-现价: %s", text)
		})

		document.Find(".deta_hqContainer >.deta03 li").Each(func(i int, selection *goquery.Selection) {
			text := strutil.RemoveNonPrintable(selection.Text())
			logger.SugaredLogger.Infof("股票名称-%s: %s", "", text)
		})

	}
}
