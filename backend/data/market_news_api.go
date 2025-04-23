package data

import (
	"github.com/samber/lo"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
)

// @Author spark
// @Date 2025/4/23 14:54
// @Desc
// -----------------------------------------------------------------------------------
type MarketNewsApi struct {
}

func NewMarketNewsApi() *MarketNewsApi {
	return &MarketNewsApi{}
}

func (m MarketNewsApi) GetTelegraphList() *[]*models.Telegraph {
	news := &[]*models.Telegraph{}
	db.Dao.Model(news).Preload("TelegraphTags").Order("id desc").Limit(20).Find(news)
	for _, item := range *news {
		tags := &[]models.Tags{}
		db.Dao.Model(&models.Tags{}).Where("id in ?", lo.Map(item.TelegraphTags, func(item models.TelegraphTags, index int) uint {
			return item.TagId
		})).Find(&tags)
		tagNames := lo.Map(*tags, func(item models.Tags, index int) string {
			return item.Name
		})
		item.SubjectTags = tagNames
		logger.SugaredLogger.Infof("tagNames %v ，SubjectTags：%s", tagNames, item.SubjectTags)
	}
	return news
}
