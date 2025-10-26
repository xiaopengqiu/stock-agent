package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
	"os"
	"time"
)

var Dao *gorm.DB

func Init(sqlitePath string) {
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second * 3,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			LogLevel:                  logger.Info,
		},
	)
	var openDb *gorm.DB
	var err error
	if sqlitePath == "" {
		sqlitePath = "data/stock.db?cache=shared&mode=rwc&_journal_mode=WAL"
	}
	openDb, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{
		Logger:                                   dbLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		PrepareStmt:                              true,
	})
	//读写分离提高sqlite效率，防止锁库
	openDb.Use(dbresolver.Register(
		dbresolver.Config{
			Replicas: []gorm.Dialector{sqlite.Open(sqlitePath)}},
	))

	if err != nil {
		log.Fatalf("db connection error is %s", err.Error())
	}

	dbCon, err := openDb.DB()
	if err != nil {
		log.Fatalf("openDb.DB error is  %s", err.Error())
	}
	dbCon.SetMaxIdleConns(10)
	dbCon.SetMaxOpenConns(100)
	dbCon.SetConnMaxLifetime(time.Hour)
	Dao = openDb
}

func CheckTableEmpty(tableName string) bool {
	var count int64
	Dao.Table(tableName).Count(&count)
	return count == 0
}
