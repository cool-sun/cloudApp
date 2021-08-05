package spider

import (
	"github.com/coolsun/cloud-app/database"
	"github.com/coolsun/cloud-app/k8s"
	"github.com/coolsun/cloud-app/model"
	"sync"
	"time"
)

var (
	cfg    *model.Config
	engine *database.Engine
	plat   k8s.Platform
	mutex  *sync.RWMutex
)

func Init(c *model.Config, e *database.Engine, p k8s.Platform) {
	cfg = c
	engine = e
	plat = p
	mutex = &sync.RWMutex{}
	go startLoadAppTags()
}

//定时每天更新一次tag
func startLoadAppTags() {
	for {
		loadAppTags()
		time.Sleep(time.Hour * 24)
	}
}
