package service

import (
	"context"
	"fmt"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/builder"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/scanner"
	"github.com/pkg/errors"
)

type Snapshot struct {
}

func (s Snapshot) ListInfo(search, userName, appName string, current, pageSize int) (listInfo []*model.SnapshotInfo, count int64, err error) {
	listInfo = make([]*model.SnapshotInfo, 0)
	list, count, err := s.List(search, userName, appName, current, pageSize)
	if err != nil {
		return
	}
	appNameArr := make([]string, 0)
	for _, v := range list {
		appNameArr = append(appNameArr, v.AppName)
	}
	appM, err := App{}.getAppShowNameByName(appNameArr)
	if err != nil {
		return
	}
	for _, v := range list {
		listInfo = append(listInfo, &model.SnapshotInfo{
			Snapshot:    v,
			AppShowName: appM[v.AppName],
		})
	}
	return
}
func (s Snapshot) List(search, userName, appName string, current, pageSize int) (list []*model.Snapshot, count int64, err error) {
	skipLines := (current - 1) * pageSize
	likeValue := "%" + search + "%"
	where := map[string]interface{}{
		"_or": []map[string]interface{}{
			{
				"name like": likeValue,
			},
			{
				"show_name like": likeValue,
			},
			{
				"app_name like": likeValue,
			},
		},
		"user_name": userName,
		"app_name":  appName,
		"_orderby":  "create_time desc",
		"_limit":    []uint{uint(skipLines), uint(pageSize)},
	}
	finalWhere := builder.OmitEmpty(where, []string{"user_name", "app_name"})
	selectFields := []string{"*"}
	cond, values, err := builder.BuildSelect("snapshot", finalWhere, selectFields)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	rows, err := engine.DB().Query(cond, values...)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	list = make([]*model.Snapshot, 0, 0)
	err = scanner.Scan(rows, &list)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	delete(finalWhere, "_limit") //如果不删除这个key的话,当查询第2页的时候，count会变成0
	rlt, err := builder.AggregateQuery(context.TODO(), engine.DB().DB, "snapshot", finalWhere, builder.AggregateCount("*"))
	count = rlt.Int64()
	return
}

func (s Snapshot) Delete(userName string, snapshotId int) (err error) {
	snapshot := &model.Snapshot{Id: snapshotId}
	has, err := engine.Get(snapshot)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if !has {
		err = errors.New(model.E10022)
		return
	}
	//todo 使用中的快照无法删除
	session := engine.NewSession()
	session.Begin()
	defer session.Close()
	session.Delete(snapshot)
	err = plat.App().SnapshotDelete(userName, snapshot.Name)
	if err != nil {
		return
	}
	session.Commit()
	return
}
func (s Snapshot) Create(userName, showName, appName string) (err error) {
	//先往数据库中插入快照记录
	name := fmt.Sprintf("%v-%v-%v", userName, appName, utils.RandStringRunes(4))
	session := engine.NewSession()
	session.Begin()
	defer session.Close()
	session.Insert(&model.Snapshot{
		UserName: userName,
		Name:     name,
		ShowName: showName,
		AppName:  appName,
	})
	err = plat.App().SnapshotCreate(userName, appName, name)
	if err != nil {
		return
	}
	session.Commit()
	return
}
