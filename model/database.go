package model

import (
	"time"
)

// App 应用表
type App struct {
	Id                int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	UserName          string    `json:"user_name" xorm:"default '' VARCHAR(20)"`
	Name              string    `json:"name" xorm:"default '' VARCHAR(100)"` //对应k8s资源名
	ShowName          string    `json:"show_name" xorm:"default '' VARCHAR(100)"`
	Controller        string    `json:"controller" xorm:"default '' VARCHAR(100)"` //对应k8s的控制器类别名称
	AppBigCategory    string    `json:"app_big_category" xorm:"default '' VARCHAR(50)"`
	AppLittleCategory string    `json:"app_little_category" xorm:"default '' VARCHAR(100)"`
	CreateTime        time.Time `json:"create_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created"`
	ModifyTime        time.Time `json:"modify_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP updated"`
	IsDelete          int       `json:"is_delete" xorm:"default 0  TINYINT(1)"`
}

// Role 角色表
type Role struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Name       string    `json:"name" xorm:"default '' VARCHAR(20)"`
	CreateTime time.Time `json:"create_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created"`
	ModifyTime time.Time `json:"modify_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP updated"`
	IsDelete   int       `json:"is_delete" xorm:"default 0  TINYINT(1)"`
}

// User 用户表
type User struct {
	Id            int            `json:"id" xorm:"not null pk autoincr INT(10)"`
	Name          string         `json:"name" xorm:"default '' VARCHAR(20) unique"` //name要是唯一的，对应k8s的命名空间
	AvatarNum     int            `json:"avatar_num" xorm:"default 1 VARCHAR(100)"`
	Phone         string         `json:"phone" xorm:"default ''  CHAR(11)"`
	Email         string         `json:"email" xorm:"default '' VARCHAR(50)"`
	Password      string         `json:"password" xorm:"default '' comment('密码,md5加密存储')  CHAR(32)"`
	RoleId        int            `json:"role_id" xorm:"default 2 "` //1表示管理员，2表示开发者
	ResourceQuota *ResourceQuota `json:"resource_quota"`
	CreateTime    time.Time      `json:"create_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created"`
	ModifyTime    time.Time      `json:"modify_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP updated"`
	IsDelete      int            `json:"is_delete" xorm:"default 0  TINYINT(1)"`
}

//应用版本表
type Tag struct {
	Id                int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Library           string    `json:"library" xorm:"default '' VARCHAR(100) unique(tag) "`
	AppLittleCategory string    `json:"app_little_category" xorm:"default '' VARCHAR(100) unique(tag) "`
	Version           string    `json:"version" xorm:"default '' VARCHAR(100) unique(tag) "` //版本号
	CreateTime        time.Time `json:"create_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created"`
	ModifyTime        time.Time `json:"modify_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP updated"`
}

type HelmApp struct {
	Id         int         `json:"id" xorm:"not null pk autoincr INT(10)"`
	Name       string      `json:"name" xorm:"default '' VARCHAR(100) unique(n) "`
	UserName   string      `json:"user_name" xorm:"default '' VARCHAR(20)"`
	RepoName   string      `json:"repo_name" xorm:"name repo_name" `
	RepoURL    string      `json:"repo_url"  xorm:"name repo_url"`
	ChartName  string      `json:"chart_name"  xorm:"name chart_name"`
	Version    string      `json:"version"  `
	Value      string      `json:"value" xorm:"text"`
	Status     int         `json:"status"` //值为0的时候显示创建中(灰色转圈图标),值为-1表示创建失败(红色失败图标)，值为1表示创建成功(绿色成功图标)
	CreateTime time.Time   `json:"create_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created"`
	ModifyTime time.Time   `json:"modify_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP updated"`
	EndPoint   interface{} `json:"end_point" xorm:"-"`
}

type Snapshot struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	UserName   string    `json:"user_name" xorm:"default '' VARCHAR(20)"`
	Name       string    `json:"name" xorm:"default '' VARCHAR(100)"`      //快照名称,对应k8s snapshot的资源名
	ShowName   string    `json:"show_name" xorm:"default '' VARCHAR(100)"` //给用户展示的名称
	AppName    string    `json:"app_name" xorm:"default '' VARCHAR(100)"`
	CreateTime time.Time `json:"create_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created"`
	ModifyTime time.Time `json:"modify_time" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP updated"`
}
