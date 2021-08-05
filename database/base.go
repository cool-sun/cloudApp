package database

import (
	"fmt"
	"github.com/coolsun/cloud-app/model"
	mylog "github.com/coolsun/cloud-app/utils/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"time"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

type Engine struct {
	*xorm.Engine
}
type MyLog struct {
	showSql bool
	lev     log.LogLevel
}

func (m *MyLog) Debug(v ...interface{}) {
	if m.lev <= log.LOG_DEBUG {
		mylog.Debug(v...)
	}
}

func (m *MyLog) Debugf(format string, v ...interface{}) {
	if m.lev <= log.LOG_DEBUG {
		mylog.Debugf(format, v...)
	}
}

func (m *MyLog) Error(v ...interface{}) {
	if m.lev <= log.LOG_ERR {
		mylog.Error(v...)
	}
}

func (m *MyLog) Errorf(format string, v ...interface{}) {
	if m.lev <= log.LOG_ERR {
		mylog.Errorf(format, v...)
	}
}

func (m *MyLog) Info(v ...interface{}) {
	if m.lev <= log.LOG_INFO {
		mylog.Info(v...)
	}
}

func (m *MyLog) Infof(format string, v ...interface{}) {
	if m.lev <= log.LOG_INFO {
		mylog.Infof(format, v...)
	}
}

func (m *MyLog) Warn(v ...interface{}) {
	if m.lev <= log.LOG_WARNING {
		mylog.Warn(v...)
	}
}

func (m *MyLog) Warnf(format string, v ...interface{}) {
	if m.lev <= log.LOG_WARNING {
		mylog.Warnf(format, v...)
	}
}

func (m *MyLog) Level() (l log.LogLevel) {
	return m.lev
}

func (m *MyLog) SetLevel(l log.LogLevel) {
	m.lev = l
}

func (m *MyLog) ShowSQL(show ...bool) {
	if len(show) == 0 {
		m.showSql = true
	} else {
		m.showSql = show[0]
	}
}

func (m *MyLog) IsShowSQL() bool {
	return m.showSql
}

var engine *xorm.Engine

//监控数据库连接，实现断线重连
func monitorConnection(url string) {
	for {
		time.Sleep(time.Second)
		if engine == nil {
			continue
		}
		if engine.Ping() != nil {
			var err error
			engine, err = xorm.NewEngine("mysql", url)
			if err != nil {
				mylog.Error("数据库连接断开，重新连接失败 : ", err)
				continue
			}
			engineSet(engine)
		}
	}
}

func engineSet(engine *xorm.Engine) {
	if engine != nil {
		engine.SetLogger(&MyLog{})
		engine.ShowSQL(true)
		engine.SetMaxIdleConns(200)
		engine.SetLogLevel(log.LOG_ERR)
	}
}
func NewEngine(cfg *model.Config) (e *Engine, err error) {
	dbName := "cloud-app"
	mysqlUrl := fmt.Sprintf("%v:%v@tcp(%v:%v)/mysql?charset=utf8mb4&parseTime=True&loc=Local", cfg.MysqlUser, cfg.MysqlPassword, cfg.MysqlHost, cfg.MysqlPort)
	engine, err = xorm.NewEngine("mysql", mysqlUrl)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	_, err = engine.Exec(fmt.Sprintf("create database  if not exists `%v` charset utf8mb4 collate utf8mb4_general_ci", dbName))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	onlineUrl := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", cfg.MysqlUser, cfg.MysqlPassword, cfg.MysqlHost, cfg.MysqlPort, dbName)
	engine, err = xorm.NewEngine("mysql", onlineUrl)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	engineSet(engine)
	err = engine.Ping()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	//同步数据库结构
	err = engine.Sync2(
		new(model.User),
		new(model.App),
		new(model.Role),
		new(model.Tag),
		new(model.HelmApp),
		new(model.Snapshot),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	e = &Engine{Engine: engine}

	go monitorConnection(onlineUrl)
	return
}

//xorm的事务操作,要先调session.Begin()，最后要调session.Commit(),
//Rollback()一般可以不用调，调用session.Close的时候，如果没调过session.Commit()，则Rollback()会被自动调。

func (engine *Engine) RowExist(bean interface{}) (b bool, err error) {
	b, err = engine.Where("is_delete = ?", "0").Exist(bean)
	return b, errors.WithStack(err)
}

func (engine *Engine) RowInsert(bean interface{}) (id int64, err error) {
	id, err = engine.Insert(bean)
	return id, errors.WithStack(err)
}

func (engine *Engine) RowUpdate(id, bean interface{}) (affected int64, err error) {
	affected, err = engine.ID(id).Cols().Update(bean)
	return affected, errors.WithStack(err)
}

func (engine *Engine) RowsDelete(bean interface{}, id []int, isHard ...bool) (affected int64, err error) {
	var hard bool
	if len(isHard) > 0 {
		hard = isHard[0]
	}
	//物理删除
	if hard {
		affected, err = engine.In("id", id).Delete(bean)
		return affected, errors.WithStack(err)
	}
	//逻辑删除
	affected, err = engine.Table(bean).In("id", id).Update(map[string]interface{}{"is_delete": 1})
	return affected, errors.WithStack(err)
}

//封装一个方法用来查询单条信息,bean的非空字段用作查询条件
func (engine *Engine) RowGet(bean interface{}, cols ...string) (b bool, err error) {
	if len(cols) == 0 {
		b, err = engine.Where("is_delete = ?", "0").Get(bean)
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	} else {
		b, err = engine.Where("is_delete = ?", "0").Cols(cols...).Get(bean)
		if err != nil {
			err = errors.WithStack(err)
			return
		}
	}
	return
}

func (engine *Engine) Count(bean interface{}) (count int64, err error) {
	count, err = engine.Where("is_delete = ?", "0").Count(bean)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
