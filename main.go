package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/coolsun/cloud-app/database"
	"github.com/coolsun/cloud-app/k8s"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/router"
	"github.com/coolsun/cloud-app/service"
	"github.com/coolsun/cloud-app/spider"
	"github.com/coolsun/cloud-app/utils/log"
	"os"
)

func main() {
	log.Info("开始启动程序")
	cfg := &model.Config{}
	err := env.Parse(cfg)
	if err != nil {
		log.Errorf("%+v\n", err)
		return
	}
	log.Info("env配置加载成功")
	log.InfoJson(cfg)
	if os.Getenv("mode") == "debug" {
		cfg.IsDebug = true
	}
	db, err := database.NewEngine(cfg)
	if err != nil {
		log.Errorf("%+v\n", err)
		return
	}
	log.Info("数据库连接成功")
	//利用chan阻塞等到k8s中的本地缓存Lister加载完毕再执行后续操作
	k8sNewDone := make(chan int)
	plat, err := k8s.New(cfg, k8sNewDone)
	if err != nil {
		log.Errorf("%+v\n", err)
		return
	}
	<-k8sNewDone

	err = service.Init(cfg, db, plat)
	if err != nil {
		log.Errorf("%+v\n", err)
		return
	}

	spider.Init(cfg, db, plat)

	r := router.Register()
	go func() {
		err = r.Run(":" + cfg.ListenPort)
		if err != nil {
			panic(err)
		}
	}()

	//利用chan读取不到数据就一直阻塞来保证主协程不退出
	done := make(chan int)
	<-done
}
