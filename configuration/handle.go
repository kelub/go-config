package configuration

import (
	"fmt"
	db "go-config/mysql"
)

type dbhandle struct {
	engine *db.MysqlEngine
}

func NewDBHandle() *dbhandle {
	engine, err := db.NewMysqlEngine(nil)
	if err != nil {
		return nil
	}
	return &dbhandle{
		engine: engine,
	}
}

func GetKeyDB(key string) ([]map[string][]byte, error) {
	e, _ := db.GetMysqlEngine()
	col := "value, lastindex"
	sql := fmt.Sprintf("select %s from config where key=%s", col, key)
	return e.Query(sql)
}

func SetKeyDB(key string, value string) (int64, error) {
	e, _ := db.GetMysqlEngine()
	config := new(db.Config)
	config.Key = key
	config.Value = value
	return e.InsertOne(config)
}
