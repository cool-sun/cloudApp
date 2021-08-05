package model

import myappv1 "github.com/coolsun/cloud-app/k8s/crd/pkg/apis/app/v1"

//app部分
type AppCreateReq struct {
	AppLittleCategory string `json:"app_little_category" binding:"required"` //app小类
	ShowName          string `json:"show_name" binding:"required"`           //名称
	VolumeSize        string `json:"volume_size"`                            //磁盘大小
	SC                string `json:"sc"`                                     //储存类型
	*ContainerApi     `json:",inline" binding:"required"`
}

type HelmReleaseCreateReq struct {
	Name      string `json:"name" binding:"required"`        //名称,需要用户输入,后端做了唯一性校验
	RepoName  string `json:"repo_name"  binding:"required"`  //对应repository.name
	RepoURL   string `json:"repo_url"  binding:"required"`   //对应repository.url
	ChartName string `json:"chart_name"  binding:"required"` //对应name
	Version   string `json:"version"  binding:"required"`    //对应version
	Value     string `json:"value"  binding:"required"`      //用户在json编辑器输入的内容
}

type HelmReleaseUpdateReq struct {
	Name      string `json:"name" binding:"required"` //用户输入的，需要全局唯一
	Value     string `json:"value"`
	RepoName  string `json:"repo_name"`
	RepoURL   string `json:"repo_url"`
	ChartName string `json:"chart_name"`
	Version   string `json:"version"`
}

type HelmReleaseListReq struct {
	UserName string `json:"user_name"`                   //用户名
	Search   string `json:"search"`                      //模糊搜索字段，可根据名称来搜索
	Current  int    `json:"current" binding:"required"`  //当前第几页
	PageSize int    `json:"pageSize" binding:"required"` //每页多少条
}

type AppDeleteReq struct {
	Name string `json:"name"`
}

type HelmReleaseReq struct {
	Name string `json:"name"` //release名称
}

type AppScaleReq struct {
	Name string `json:"name" binding:"required"` //app名称
	CPU  string `json:"cpu" binding:"required"`  //cpu核心数，例如1，2，3，4
	Mem  string `json:"mem" binding:"required"`  //内存大小，例如1，2，3，4
}

type AppRestartReq struct {
	Name string `json:"name" binding:"required"` //app名称，注意不是show_name
}

type AppReplicasReq struct {
	Name     string `json:"name" binding:"required"` //app名称，注意不是show_name
	Replicas int    `json:"replicas" binding:"required"`
}

type SnapshotCreateReq struct {
	ShowName string `json:"show_name" binding:"required"` //名称
	AppName  string `json:"app_name"  binding:"required"` //要创建快照的app名称，注意不是show_name
}
type SnapshotListReq struct {
	Search   string `json:"search"`                      //模糊搜索字段，可不传
	UserName string `json:"user_name"`                   //用户名,可不传
	AppName  string `json:"app_name"`                    //app名称，可不传
	Current  int    `json:"current" binding:"required"`  //当前第几页
	PageSize int    `json:"pageSize" binding:"required"` //每页多少条
}
type SnapshotDeleteReq struct {
	Id int `json:"id" binding:"required"` //快照id
}

type AppRestoreReq struct {
	AppName    string `json:"app_name" binding:"required"`
	SnapshotId int    `json:"snapshot_id" binding:"required"` //快照id
}

type AppListReq struct {
	ShowName string `json:"show_name"`                   //名称
	Category string `json:"category"`                    //类别
	UserName string `json:"user_name"`                   //用户名
	Current  int    `json:"current" binding:"required"`  //当前第几页
	PageSize int    `json:"pageSize" binding:"required"` //每页多少条
}

type AppOpenReq struct {
	Name   string `json:"name" binding:"required"` //app名称
	IsOpen bool   `json:"is_open"`                 //开启时传 true，关闭时传 false
}
type AppVersionReq struct {
	Name    string `json:"name" binding:"required"`    //app名称
	Version string `json:"version" binding:"required"` //版本
}
type AppEnvReq struct {
	Name string            `json:"name" binding:"required"` //app名称
	Env  []*myappv1.EnvVar `json:"env"`                     //env,格式同创建app
}

type AppConfigUpdateReq struct {
	Name string            `json:"name" binding:"required"` //app名称
	Data map[string]string `json:"data" binding:"required"`
}

type AppLittleCategoryTagReq struct {
	LittleCategory string `json:"little_category" binding:"required"`
	Search         string `json:"search"`                      //模糊搜索字段，可不传
	Current        int    `json:"current" binding:"required"`  //当前第几页
	PageSize       int    `json:"pageSize" binding:"required"` //每页多少条
}

//用户部分
type UserRegisterReq struct {
	Name          string         `json:"name" binding:"required"`               //用户名
	Email         string         `json:"email"`                                 //邮箱号
	Phone         string         `json:"phone"`                                 //手机号
	PassWord      string         `json:"password"`                              //密码
	RoleId        int            `json:"role_id" binding:"required,oneof=1 2" ` //角色 1表示管理员 2表示开发者
	ResourceQuota *ResourceQuota `json:"resource_quota"`                        //资源限制相关的,不传表示不限制
}

type UserManagerEditReq struct {
	Name          string         `json:"name" binding:"required"`      //用户名
	Email         string         `json:"email"`                        //邮箱号
	Phone         string         `json:"phone"`                        //手机号
	PassWord      string         `json:"password"`                     //密码
	RoleId        int            `json:"role_id" binding:"oneof=1 2" ` //角色 1表示管理员 2表示开发者
	ResourceQuota *ResourceQuota `json:"resource_quota"`               //资源限制相关的,不传表示不限制
}

type ResourceQuota struct {
	CPU string `json:"cpu" binding:"required"` //可以使用的cpu总核数
	Mem string `json:"mem" binding:"required"` //可以使用的内存总量
}

type UserLoginReq struct {
	Name     string `json:"name"`     //用户名
	Password string `json:"password"` //密码
}

type UserDeleteReq struct {
	Users []*UserDeleteBase `json:"users"`
}

type UserDeleteBase struct {
	Name       string `json:"name" binding:"required"` //要删除的用户名
	IsCleanApp bool   `json:"is_clean_app"`            //是否删除对应用户下的所有app
}

type UserListReq struct {
	RoleId   int    `json:"role_id"`                     //角色,可不传 1表示管理员 2表示开发者
	Search   string `json:"search"`                      //模糊搜索字段，可不传
	Current  int    `json:"current" binding:"required"`  //当前第几页
	PageSize int    `json:"pageSize" binding:"required"` //每页多少条
}

type UserEditReq struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}
