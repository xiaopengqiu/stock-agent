package data

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/duke-git/lancet/v2/strutil"
	"go-stock/backend/logger"
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
		chromedp.Click("div.cos-tab:nth-child(5)", chromedp.ByQuery),
		chromedp.ScrollIntoView("div.body-box"),
		chromedp.WaitVisible("div.body-col"),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(1 * time.Second),
	}
	htmlContent, success := crawlerAPI.GetHtmlWithActions(&actions, true)
	if success {
		document, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			logger.SugaredLogger.Error(err.Error())
		}
		var messages []string
		document.Find("div.finance-hover,div.list-date").Each(func(i int, selection *goquery.Selection) {
			text := strutil.RemoveWhiteSpace(selection.Text(), false)
			messages = append(messages, text)
			logger.SugaredLogger.Infof("搜索到消息-%s: %s", "", text)
		})
		logger.SugaredLogger.Infof("messages:%d", len(messages))
	}
	//logger.SugaredLogger.Infof("htmlContent:%s", htmlContent)
}
