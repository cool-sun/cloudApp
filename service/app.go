package service

import (
	"context"
	"fmt"
	"github.com/coolsun/cloud-app/k8s"
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/builder"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/scanner"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type App struct {
}

//获取支持的存储类
func (a App) GetSc() ([]*model.SC, error) {
	return plat.App().GetScList()
}

//从快照恢复
func (a App) Restore(namespace string, appName string, snapshotId int) (err error) {
	snapshot := &model.Snapshot{Id: snapshotId}
	has, _ := engine.Get(snapshot)
	if !has {
		err = errors.New(model.E10022)
		return
	}
	return plat.App().UpdatePvcDataSource(namespace, appName, snapshot.Name)
}

//app详情信息
func (a App) Detail(user *model.User, appName string) (detail *model.AppDetailBase, err error) {
	dbApp, err := a.getAppByName(appName)
	if err != nil {
		return
	}
	crdApp, err := plat.App().Get(user.Name, appName)
	if err != nil {
		return
	}
	metrics, err := plat.App().GetMetric(user.Name, appName)
	detail = &model.AppDetailBase{
		App:           dbApp,
		Replicas:      int(crdApp.Spec.Replicas),
		MainContainer: crdApp.Spec.Pod[myappv1.Main],
		Metrics:       metrics,
	}
	return
}

func (a App) List(userName, showName, category string, current, pageSize int) (list []*model.AppInfo, count int64, err error) {
	list = make([]*model.AppInfo, 0, 0)
	apps, count, err := a.getAppFromDB(showName, category, userName, current, pageSize)
	for _, v := range apps {
		endPoint := plat.App().GetEndPoint(v.UserName, v.Name)
		list = append(list, &model.AppInfo{
			App:                       v,
			AppBigCategoryShowName:    plat.App().GetBigCategoryFillInfo(v.AppBigCategory).ShowName,
			AppLittleCategoryShowName: plat.App().GetLittleCategoryFillInfo(v.AppLittleCategory).ShowName,
			Status:                    plat.App().GetStatus(userName, v.Name),
			EndPoint:                  map[string]interface{}{"in": strings.Join(endPoint.In, ","), "out": strings.Join(endPoint.Out, ",")},
		})
	}
	return
}

//app列表信息
func (a App) getAppFromDB(showName, category, userName string, current, pageSize int) (list []*model.App, count int64, err error) {
	skipLines := (current - 1) * pageSize
	likeValue := "%" + showName + "%"
	where := map[string]interface{}{
		"show_name like": likeValue,
		"user_name":      userName,
		"is_delete":      0,
		"_orderby":       "create_time desc",
		"_limit":         []uint{uint(skipLines), uint(pageSize)},
	}
	if category != "" {
		where["_or"] = []map[string]interface{}{
			{
				"app_big_category": category,
			},
			{
				"app_little_category": category,
			},
		}
	}
	omitEmptyArr := []string{"user_name"}
	finalWhere := builder.OmitEmpty(where, omitEmptyArr)
	selectFields := []string{"id", "user_name", "name", "show_name", "app_big_category", "app_little_category", "is_delete", "create_time", "modify_time"}
	cond, values, err := builder.BuildSelect("app", finalWhere, selectFields)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	rows, err := engine.DB().Query(cond, values...)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	list = make([]*model.App, 0, 0)
	err = scanner.Scan(rows, &list)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	delete(finalWhere, "_limit") //如果不删除这个key的话,当查询第2页的时候，count会变成0
	rlt, err := builder.AggregateQuery(context.TODO(), engine.DB().DB, "app", finalWhere, builder.AggregateCount("*"))
	count = rlt.Int64()
	return
}

func (a App) getAppShowNameByName(names []string) (appMap map[string]string, err error) {
	apps := make([]*model.App, 0)
	appMap = make(map[string]string, 0)
	engine.In("name", names).Cols("name", "show_name").Find(&apps)
	for _, v := range apps {
		appMap[v.Name] = v.ShowName
	}
	return
}
func (a App) getAppByName(name string) (app *model.App, err error) {
	app = &model.App{Name: name}
	exist, err := engine.Get(app)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if !exist {
		err = errors.New(model.E10012)
		return
	}
	return
}

//app开启或关闭集群外访问
func (a App) Open(user *model.User, appName string, isOpen bool) (err error) {
	app, err := a.getAppByName(appName)
	if err != nil {
		return
	}
	//StatefulSet部署的应用采用的headlessService,不支持将类型改成NodePort
	if app.Controller == string(k8s.StatefulSet) {
		err = errors.New(model.E10016)
		return
	}
	return plat.App().OpenOrClosePort(user.Name, appName, isOpen)
}

//app重启
func (a App) Restart(namespace, appName string) (err error) {
	return plat.App().DeleteAppPods(namespace, appName)
}

//app内存和cpu扩缩容量
func (a App) Scale(user *model.User, appName, CPU, Mem string) (err error) {
	app, err := plat.App().Get(user.Name, appName)
	if err != nil {
		return
	}
	app.Spec.Pod[myappv1.Main].CPU = CPU
	app.Spec.Pod[myappv1.Main].Mem = Mem
	err = plat.App().Create(app)
	return
}

//获取app配置
func (a App) Config(user *model.User, appName string) (path string, config map[string]string, err error) {
	app := &model.App{Name: appName}
	b, err := engine.RowGet(app)
	if err != nil {
		return
	}
	if !b {
		return
	}
	template, err := plat.App().GetTemplateByLittleCategory(app.AppLittleCategory)
	if err != nil {
		return
	}
	if template.Config.Path == "" {
		return
	}
	path = template.Config.Path
	config, err = plat.App().GetConfig(user.Name, appName, template.Config.Path, template.Config.File)
	return
}

//更新app配置
func (a App) ConfigUpdate(user *model.User, appName string, data map[string]string) (err error) {
	app := &model.App{Name: appName}
	b, err := engine.RowGet(app)
	if err != nil {
		return
	}
	if !b {
		return
	}
	template, err := plat.App().GetTemplateByLittleCategory(app.AppLittleCategory)
	if err != nil {
		return
	}
	return plat.App().UpdateConfig(user.Name, appName, template.Config.Path, data)
}

//删除app
func (a App) Delete(user *model.User, appName string) (err error) {
	session := engine.NewSession()
	defer session.Close()
	session.Begin()
	session.Delete(&model.App{Name: appName})
	err = plat.App().Delete(user.Name, appName)
	if err != nil {
		return
	}
	session.Commit()
	return
}

//创建app
func (a App) Create(user *model.User, para *model.AppCreateReq) (err error) {
	session := engine.NewSession()
	defer session.Close()
	session.Begin()
	app, err := getCompleteTemplate(user.Name, para)
	if err != nil {
		return
	}
	_, err = session.Insert(&model.App{
		UserName:          user.Name,
		Name:              app.Name,
		ShowName:          para.ShowName,
		Controller:        app.Spec.Controller,
		AppBigCategory:    app.Spec.AppBigCategory,
		AppLittleCategory: app.Spec.AppLittleCategory,
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = plat.App().Create(app)
	if err != nil {
		return
	}
	session.Commit()
	return
}

func isNeedVolume(m map[myappv1.ContainerType]*myappv1.Container) bool {
	for _, v := range m {
		if v != nil && len(v.VolumeMount) > 0 {
			return true
		}
	}
	return false
}

//构建完整的Template
func getCompleteTemplate(namespace string, para *model.AppCreateReq) (app *myappv1.App, err error) {
	tmpl, err := plat.App().GetTemplateByLittleCategory(para.AppLittleCategory)
	if err != nil {
		return
	}
	err = getEnv(tmpl.Pod[myappv1.Main].Env, para.Env)
	if err != nil {
		return
	}
	app = &myappv1.App{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-%v-%v", tmpl.AppBigCategory, tmpl.AppLittleCategory, utils.RandStringRunes(4)),
			Namespace: namespace,
		},
		Spec: myappv1.AppSpec{
			AppBigCategory:    tmpl.AppBigCategory,
			AppLittleCategory: tmpl.AppLittleCategory,
			Replicas:          tmpl.Replicas,
			Strategy:          tmpl.Strategy,
			Pod:               getPod(tmpl, para),
			Controller:        tmpl.Controller,
			Config: myappv1.ConfigInfo{
				Path: "",
				File: []string{},
			},
		},
	}
	if isNeedVolume(tmpl.Pod) {
		app.Spec.Pvc = &myappv1.Pvc{
			Size: para.VolumeSize,
			SC:   para.SC,
		}
	}
	return
}

//构建完整的pod里的容器
func getPod(tmpl *myappv1.AppSpec, para *model.AppCreateReq) (pod map[myappv1.ContainerType]*myappv1.Container) {
	pod = map[myappv1.ContainerType]*myappv1.Container{
		myappv1.Main: {
			Image:       tmpl.Pod[myappv1.Main].Image,
			Tag:         para.Tag,
			CPU:         para.CPU,
			Mem:         para.Mem,
			Env:         append(para.Env),
			Ports:       tmpl.Pod[myappv1.Main].Ports,
			VolumeMount: tmpl.Pod[myappv1.Main].VolumeMount,
			Command:     tmpl.Pod[myappv1.Main].Command,
			Args:        tmpl.Pod[myappv1.Main].Args,
		},
	}
	if tmpl.Pod[myappv1.Init] != nil {
		pod[myappv1.Init] = tmpl.Pod[myappv1.Init]
	}
	if tmpl.Pod[myappv1.Sidecar] != nil {
		pod[myappv1.Sidecar] = tmpl.Pod[myappv1.Sidecar]
	}
	return
}

//从paraEnv拿到需要填充的env值，拼凑出完整的env
func getEnv(tmplEnv []*myappv1.EnvVar, paraEnv []*myappv1.EnvVar) (err error) {
	for i, t := range tmplEnv {
		//如果tmplEnv中某个env的值为空, 就从传入的paraEnv中查找相应的值，如果没有找到就报错
		if t.GetValue() == "" {
			for _, p := range paraEnv {
				if p.GetName() == t.GetName() {
					tmplEnv[i].SetValue(p.GetValue())
				}
			}
			if tmplEnv[i].GetValue() == "" {
				err = errors.New(model.E10011)
				return
			}
		}
	}
	return
}
