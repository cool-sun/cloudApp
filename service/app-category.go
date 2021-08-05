package service

import (
	"context"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/builder"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/scanner"
	"github.com/pkg/errors"
)

type AppCategory struct {
}

func (ac AppCategory) GetTmpl() (data []*model.AppCategoryTmplBaseArr) {
	data = make([]*model.AppCategoryTmplBaseArr, 0)
	cts := plat.App().GetTemplates()
	for _, v := range cts {
		var exist bool
		fillInfo := plat.App().GetLittleCategoryFillInfo(v.AppLittleCategory)
		obj := &model.AppCategoryTmplBase{
			AppLittleCategory: v.AppLittleCategory,
			Icon:              fillInfo.Icon,
			Summary:           fillInfo.Summary,
			ShowName:          fillInfo.ShowName,
			Env:               v.Pod[myappv1.Main].Env,
		}
		if isNeedVolume(v.Pod) {
			obj.IsNeedVolume = true
		}
		for _, val := range data {
			if val.Name == v.AppBigCategory {
				val.Apps = append(val.Apps, obj)
				exist = true
				break
			}
		}
		if !exist {
			data = append(data, &model.AppCategoryTmplBaseArr{
				ShowName: plat.App().GetBigCategoryFillInfo(v.AppBigCategory).ShowName,
				Name:     v.AppBigCategory,
				Apps:     []*model.AppCategoryTmplBase{obj},
			})
		}
	}
	return
}

func (ac AppCategory) GetTagsList(pageNum, pageSize int, littleCategory, search string) (list []*model.Tag, count int64, err error) {
	skipLines := (pageNum - 1) * pageSize
	likeValue := "%" + search + "%"
	where := map[string]interface{}{
		"app_little_category": littleCategory,
		"version like":        likeValue,
		"_orderby":            "create_time desc",
		"_limit":              []uint{uint(skipLines), uint(pageSize)},
	}
	finalWhere := builder.OmitEmpty(where, []string{})
	selectFields := []string{"version"}
	cond, values, err := builder.BuildSelect("tag", finalWhere, selectFields)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	rows, err := engine.DB().Query(cond, values...)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	list = make([]*model.Tag, 0, 0)
	err = scanner.Scan(rows, &list)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	delete(finalWhere, "_limit") //如果不删除这个key的话,当查询第2页的时候，count会变成0
	rlt, err := builder.AggregateQuery(context.TODO(), engine.DB().DB, "tag", finalWhere, builder.AggregateCount("*"))
	count = rlt.Int64()
	return
}
