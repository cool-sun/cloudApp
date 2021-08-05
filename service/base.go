package service

import (
	"github.com/coolsun/cloud-app/database"
	"github.com/coolsun/cloud-app/k8s"
	"github.com/coolsun/cloud-app/model"
	"sync"
)

var (
	cfg    *model.Config
	engine *database.Engine
	plat   k8s.Platform
	mutex  *sync.RWMutex
)

func Init(c *model.Config, e *database.Engine, p k8s.Platform) (err error) {
	cfg = c
	engine = e
	plat = p
	mutex = &sync.RWMutex{}
	go loopCreateRelease()
	createRole()
	if p != nil {
		return createRootUser()
	}
	return
}

//创建系统角色
func createRole() {
	_, _ = engine.Insert([]*model.Role{{
		Id:   1,
		Name: "管理员",
	}, {
		Id:   2,
		Name: "开发者",
	}})
	return
}

//创建root用户
func createRootUser() (err error) {
	var name = "cloud-app"
	exist := User{}.Exist(name)
	if exist {
		return User{}.Update("cloud-app", "", "", cfg.AdminPassword, 1, nil)
	} else {
		return User{}.Create(name, "", "", cfg.AdminPassword, 1, nil)
	}
}
