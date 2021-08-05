package service

import (
	"context"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/builder"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/scanner"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type Helm struct {
}

func (h Helm) GetReleaseValues(userName, releaseName string) (map[string]interface{}, error) {
	return plat.Helm().GetReleaseValues(userName, releaseName)
}

func (h Helm) ReleaseDelete(userName, releaseName string) (err error) {
	helmApp := &model.HelmApp{
		Name: releaseName,
	}
	has, err := engine.Get(helmApp)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if !has {
		err = errors.New(model.E10012)
		return
	}
	session := engine.NewSession()
	defer session.Close()
	session.Begin()
	_, err = session.ID(helmApp.Id).Delete(&model.HelmApp{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = plat.Helm().UninstallRelease(userName, releaseName)
	if err != nil {
		return
	}
	session.Commit()
	return
}

func (h Helm) ReleaseList(userName string, para *model.HelmReleaseListReq) (list []*model.HelmApp, count int64, err error) {
	list, count, err = h.getReleaseFromDB(userName, para.Search, para.Current, para.PageSize)
	for _, v := range list {
		endPoint, _ := plat.Helm().GetReleaseEndPoint(v.UserName, v.Name)
		v.EndPoint = map[string]interface{}{"in": strings.Join(endPoint.In, ","), "out": strings.Join(endPoint.Out, ",")}
	}
	return
}

func (h Helm) getReleaseFromDB(userName, search string, current, pageSize int) (list []*model.HelmApp, count int64, err error) {
	skipLines := (current - 1) * pageSize
	likeValue := "%" + search + "%"
	where := map[string]interface{}{
		"name like": likeValue,
		"user_name": userName,
		"_orderby":  "create_time desc",
		"_limit":    []uint{uint(skipLines), uint(pageSize)},
	}
	omitEmptyArr := []string{"user_name"}
	finalWhere := builder.OmitEmpty(where, omitEmptyArr)
	//selectFields := []string{"id", "name", "user_name","repo_name","repo_url","chart_name","version", "create_time", "modify_time"}
	selectFields := []string{"*"}
	cond, values, err := builder.BuildSelect("helm_app", finalWhere, selectFields)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	rows, err := engine.DB().Query(cond, values...)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	list = make([]*model.HelmApp, 0, 0)
	err = scanner.Scan(rows, &list)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	delete(finalWhere, "_limit") //如果不删除这个key的话,当查询第2页的时候，count会变成0
	rlt, err := builder.AggregateQuery(context.TODO(), engine.DB().DB, "helm_app", finalWhere, builder.AggregateCount("*"))
	count = rlt.Int64()
	return
}

//func (h Helm) ReleaseUpdate(userName string, para *model.HelmReleaseUpdateReq) (err error) {
//	helmApp := &model.HelmApp{
//		Name:     para.Name,
//		UserName: userName,
//	}
//	has, err := engine.Get(helmApp)
//	if err != nil {
//		err = errors.WithStack(err)
//		return
//	}
//	if !has {
//		err = errors.New(model.E10012)
//		return
//	}
//	session := engine.NewSession()
//	defer session.Close()
//	session.Begin()
//	_, err = session.ID(helmApp.Id).Update(&model.HelmApp{
//		Name:      para.Name,
//		UserName:  userName,
//		RepoName:  para.RepoName,
//		RepoURL:   para.RepoURL,
//		ChartName: para.ChartName,
//		Version:   para.Version,
//		Value:     para.Value,
//	})
//	if err != nil {
//		err = errors.WithStack(err)
//		return
//	}
//	opt, err := parsingHelmV3Command("", "")
//	if err != nil {
//		return
//	}
//	err = plat.Helm().InstallOrUpgradeChart(userName, opt.RepoName, opt.RepoURL, para.Name, opt.ChartName, opt.Version, para.Value)
//	if err != nil {
//		return
//	}
//	session.Commit()
//	return
//}

type helmOption struct {
	RepoName  string `json:"repo_name"`
	RepoURL   string `json:"repo_url"`
	ChartName string `json:"chart_name"`
	Version   string `json:"version"`
}

func parsingHelmV3Command(addRepository, installChart string) (opt *helmOption, err error) {
	addRepositoryArr := strings.Split(addRepository, " ")
	if len(addRepositoryArr) < 5 {
		err = errors.New(model.E10018)
		return
	}
	installChartArr := strings.Split(installChart, " ")
	if len(installChartArr) < 6 {
		err = errors.New(model.E10019)
		return
	}
	opt = &helmOption{
		RepoName:  addRepositoryArr[3],
		RepoURL:   addRepositoryArr[4],
		ChartName: installChartArr[3],
		Version:   installChartArr[5],
	}
	return
}

func (h Helm) ReleaseCreate(userName string, para *model.HelmReleaseCreateReq) (err error) {
	exist, err := engine.Exist(&model.HelmApp{
		UserName: userName,
		Name:     para.Name,
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if exist {
		err = errors.New(model.E10004)
		return
	}
	_, err = engine.Insert(&model.HelmApp{
		Name:      para.Name,
		UserName:  userName,
		RepoName:  para.RepoName,
		RepoURL:   para.RepoURL,
		ChartName: para.ChartName,
		Version:   para.Version,
		Value:     para.Value,
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func loopCreateRelease() {
	for {
		createRelease()
		time.Sleep(time.Second * 10)
	}
}

func createRelease() {
	jobs := make([]*model.HelmApp, 0)
	engine.Where("status = ?", 0).Find(&jobs)
	for _, v := range jobs {
		value := strings.ReplaceAll(v.Value, `\"`, `"`)
		err := plat.Helm().InstallOrUpgradeChart(v.UserName, v.RepoName, v.RepoURL, v.Name, v.RepoName+"/"+v.ChartName, v.Version, value)
		if err != nil {
			engine.ID(v.Id).Cols("status").Update(&model.HelmApp{Status: -1})
			continue
		} else {
			engine.ID(v.Id).Cols("status").Update(&model.HelmApp{Status: 1})
			continue
		}
	}

}
