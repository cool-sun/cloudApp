package model

import (
	myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type ContainerApi struct {
	//版本号
	Tag string `json:"tag" binding:"required"`
	//cpu核心数，例如1，2，3，4
	CPU string `json:"cpu" binding:"required"`
	//内存大小，例如1，2，3，4
	Mem string `json:"mem" binding:"required"`
	//环境变量，格式key:value,例如[{"ROOT_PASSWORD:2222"},{"AGE:4444"}]
	Env []*myappv1.EnvVar `json:"env"`
}

type AppCategoryTmplRes struct {
	Code int                       `json:"code"`
	Msg  string                    `json:"msg"`
	Data []*AppCategoryTmplBaseArr `json:"data"`
}

type AppCategoryTmplBaseArr struct {
	ShowName string                 `json:"show_name"`
	Name     string                 `json:"name"`
	Apps     []*AppCategoryTmplBase `json:"apps"`
}
type AppCategoryTmplBase struct {
	AppLittleCategory string            `json:"app_little_category"` //标题
	ShowName          string            `json:"show_name"`           //名称
	Icon              string            `json:"icon"`                //图标
	Summary           string            `json:"summary"`             //简介
	IsNeedVolume      bool              `json:"is_need_volume"`      //是否需要磁盘
	Env               []*myappv1.EnvVar `yaml:"env" json:"env"`      //必备的环境变量
}

type AppLittleCategoryTagRes struct {
	Code int                       `json:"code"`
	Msg  string                    `json:"msg"`
	Data *AppLittleCategoryTagBase `json:"data"`
}
type AppLittleCategoryTagBase struct {
	Count int64  `json:"count"`
	Tags  []*Tag `json:"tags"`
}

type AppDetailRes struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data *AppDetailBase `json:"data"`
}

type AppDetailBase struct {
	*App          `json:",inline"`
	Replicas      int                `json:"replicas"`
	MainContainer *myappv1.Container `json:"main_container"`
	Metrics       interface{}        `json:"metrics"`
}

type PodInfo struct {
	Name  []string `json:"name"`
	Limit string   `json:"limit"`
	Used  string   `json:"used"`
}

type AppConfigRes struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data *AppConfigBase `json:"data"`
}
type AppConfigBase struct {
	Path string            `json:"path"`
	File map[string]string `json:"file"`
}

type UserLoginRes struct {
	Code int               `json:"code"`
	Msg  string            `json:"msg"`
	Data *UserLoginResBase `json:"data"`
}

type UserLoginResBase struct {
	*User     `json:",inline"`
	AvatarUrl string `json:"avatar_url" xorm:"-"`
	RoleName  string `json:"role_name"`
}

type UserListRes struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data *UserListBase `json:"data"`
}

type UserListBase struct {
	Count int64   `json:"count"`
	List  []*User `json:"list"`
}

type AppListRes struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data *AppListBase `json:"data"`
}

type HelmReleaseListRes struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data *HelmReleaseList `json:"data"`
}
type HelmReleaseList struct {
	Count int64      `json:"count"`
	List  []*HelmApp `json:"list"`
}
type AppListBase struct {
	Count int64      `json:"count"`
	List  []*AppInfo `json:"list"`
}

type AppInfo struct {
	*App                      `json:",inline"`
	Status                    string      `json:"status"` //空字符串""表示创建中，1表示停止，2和3表示正常运行(2说明副本还没有全部就绪,3说明副本全部就绪)
	AppBigCategoryShowName    string      `json:"app_big_category_show_name"`
	AppLittleCategoryShowName string      `json:"app_little_category_show_name"`
	EndPoint                  interface{} `json:"end_point"`
}

type EndPoint struct {
	In  []string `json:"in"`  //集群内访问地址
	Out []string `json:"out"` //集群外访问地址
}

type SCListRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []*SC  `json:"data"`
}

type SC struct {
	Name                 string `json:"name"`
	AllowVolumeExpansion bool   `json:"allow_volume_expansion"` //是否允许扩容
}

type SnapshotListRes struct {
	Code int               `json:"code"`
	Msg  string            `json:"msg"`
	Data *SnapshotListBase `json:"data"`
}
type SnapshotListBase struct {
	Count int             `json:"count"`
	List  []*SnapshotInfo `json:"list"`
}

type SnapshotInfo struct {
	*Snapshot   `json:",inline"`
	AppShowName string `json:"app_show_name"`
}
