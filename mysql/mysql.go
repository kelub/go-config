package mysql

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

type MysqlEngine struct {
	*xorm.Engine
}

var GmysqlEngine *MysqlEngine

func init() {
	GmysqlEngine, _ = NewMysqlEngine(nil)
	//gMysqlengine = mysqlEngine
}

func NewMysqlEngine(conf *mysql.Config) (*MysqlEngine, error) {
	conf = &mysql.Config{
		User:   "config",
		Passwd: "dell+ikbc",
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "config",
	}
	dsn := conf.FormatDSN()
	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		logrus.Errorf("get new engine error err:%v", err)
		return nil, fmt.Errorf("get new engine error err:%v", err)
	}
	err = engine.Ping()
	if err != nil {
		logrus.Errorf("ping engine error err:%v", err)
		return nil, fmt.Errorf("get new engine error err:%v", err)
	}
	return &MysqlEngine{
		engine,
	}, nil
}

func GetMysqlEngine() (*MysqlEngine, error) {
	err := GmysqlEngine.Ping()
	if err != nil {
		logrus.Errorf("ping engine error err:%v", err)
		return nil, fmt.Errorf("get new engine error err:%v", err)
	}
	return GmysqlEngine, nil
}

func (e *MysqlEngine) Table(tableNameOrBean interface{}) *xorm.Session {
	session := e.Engine.Table(tableNameOrBean)
	return session
}

func (e *MysqlEngine) SQL(query interface{}, args ...interface{}) *xorm.Session {
	session := e.Engine.SQL(query, args...)
	return session
}

func (e *MysqlEngine) Select(str string) *xorm.Session {
	session := e.Engine.Select(str)
	return session
}

func (e *MysqlEngine) Query(sqlorArgs ...interface{}) ([]map[string][]byte, error) {
	resultsSlice, err := e.Engine.Query(sqlorArgs...)
	if err != nil {
		logrus.Errorf("sql error func: [Query],sql: [%s],err: [%v]", sqlorArgs, err)
	}
	return resultsSlice, err
}

func (e *MysqlEngine) QueryString(sqlorArgs ...interface{}) ([]map[string]string, error) {
	records, err := e.Engine.QueryString(sqlorArgs...)
	if err != nil {
		logrus.Errorf("sql error func: [QueryString],sql: [%s],err: [%v]", sqlorArgs, err)
	}
	return records, err
}

func (e *MysqlEngine) Exec(sqlorArgs ...interface{}) (sql.Result, error) {
	result, err := e.Engine.Exec(sqlorArgs...)
	if err != nil {
		logrus.Errorf("sql error func: [Exec],sql: [%s],err: [%v]", sqlorArgs, err)
	}
	return result, err
}

//ORM
func (e *MysqlEngine) Insert(beans ...interface{}) (int64, error) {
	affected, err := e.Engine.Insert(beans...)
	if err != nil {
		logrus.Errorf("sql error func: [Insert],sql: [%s],err: [%v]", affected, err)
	}
	return affected, err
}

func (e *MysqlEngine) InsertOne(beans interface{}) (int64, error) {
	affected, err := e.Engine.InsertOne(beans)
	if err != nil {
		logrus.Errorf("sql error func: [InsertOne],sql: [%s],err: [%v]", affected, err)
	}
	return affected, err
}

func (e *MysqlEngine) Update(bean interface{}, condiBeans ...interface{}) (int64, error) {
	affected, err := e.Engine.Update(bean, condiBeans...)
	if err != nil {
		logrus.Errorf("sql error func: [Update],sql: [%s],err: [%v]", affected, err)
	}
	return affected, err
}

func (e *MysqlEngine) Delete(bean interface{}) (int64, error) {
	affected, err := e.Engine.Delete(bean)
	if err != nil {
		logrus.Errorf("sql error func: [Update],sql: [%s],err: [%v]", affected, err)
	}
	return affected, err
}
