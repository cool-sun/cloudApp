package controller

import (
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/service"
	"github.com/gin-gonic/gin"
)

type App struct {
}

// @Tags app
// @Summary app创建
// @Accept json
// @Produce json
// @Param "" body model.AppCreateReq true "app创建"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/create [POST]
func (a App) Create(c *gin.Context) {
	//todo 参数校验  ，包括各资源限制的范围，tag是否存在，env的值是否存在,cpu和mem只能是数字,磁盘大小也只能是数字
	var para model.AppCreateReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.App{}.Create(&user, &para)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app
// @Summary app删除
// @Accept json
// @Produce json
// @Param "" body model.AppDeleteReq true "app删除"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/delete [POST]
func (a App) Delete(c *gin.Context) {
	var para model.AppDeleteReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.App{}.Delete(&user, para.Name)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app
// @Summary app扩缩容
// @Accept json
// @Produce json
// @Param "" body model.AppScaleReq true "app扩缩容,主要是内存和cpu的调整"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/scale [POST]
func (a App) Scale(c *gin.Context) {
	var para model.AppScaleReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.App{}.Scale(&user, para.Name, para.CPU, para.Mem)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app
// @Summary app重启
// @Accept json
// @Produce json
// @Param "" body model.AppRestartReq true "app重启"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/restart [POST]
func (a App) Restart(c *gin.Context) {
	var para model.AppRestartReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.App{}.Restart(user.Name, para.Name)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app
// @Summary app停止
// @Accept json
// @Produce json
// @Param "" body model.AppRestartReq true "app停止"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/stop [POST]
func (a App) Stop(c *gin.Context) {
	SuccessResponse(c, &model.Result{})
	return
}

// @Tags app
// @Summary app启动
// @Accept json
// @Produce json
// @Param "" body model.AppRestartReq true "app启动"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/start [POST]
func (a App) Start(c *gin.Context) {
	SuccessResponse(c, &model.Result{})
	return
}

// @Tags app
// @Summary app副本数伸缩
// @Accept json
// @Produce json
// @Param "" body model.AppReplicasReq true "app副本数伸缩"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/replicas [POST]
func (a App) Replicas(c *gin.Context) {
	SuccessResponse(c, &model.Result{})
	return
}

// @Tags app
// @Summary app根据快照恢复数据
// @Accept json
// @Produce json
// @Param "" body model.AppRestoreReq true "app根据快照恢复数据"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/restore [POST]
func (a App) Restore(c *gin.Context) {
	var para model.AppRestoreReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.App{}.Restore(user.Name, para.AppName, para.SnapshotId)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app
// @Summary 获取支持的存储类型
// @Accept json
// @Produce json
// @Success 200 {object} model.SCListRes "获取支持的存储类型"
// @Failure 400 {object} model.Result ""
// @Router /app/sc [GET]
func (a App) SC(c *gin.Context) {
	scList, err := service.App{}.GetSc()
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.SCListRes{
		Code: 0,
		Msg:  "",
		Data: scList,
	})
}

// @Tags app
// @Summary app信息列表
// @Accept json
// @Produce json
// @Param "" body model.AppListReq true "app信息列表"
// @Success 200 {object} model.AppListRes ""
// @Failure 400 {object} model.Result ""
// @Router /app/list [POST]
func (a App) List(c *gin.Context) {
	var para model.AppListReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	//开发者只能查看自己名下的app
	if user.RoleId == 2 || para.UserName == "" {
		para.UserName = user.Name
	}
	list, count, err := service.App{}.List(para.UserName, para.ShowName, para.Category, para.Current, para.PageSize)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.AppListRes{
		Code: 0,
		Msg:  "",
		Data: &model.AppListBase{
			Count: count,
			List:  list,
		},
	})
	//注意app的状态
}

// @Tags app
// @Summary app开启或关闭集群外访问
// @Accept json
// @Produce json
// @Param "" body model.AppOpenReq true "app开启或关闭集群外访问"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/open [POST]
func (a App) Open(c *gin.Context) {
	var para model.AppOpenReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.App{}.Open(&user, para.Name, para.IsOpen)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app
// @Summary app版本更换
// @Accept json
// @Produce json
// @Param "" body model.AppVersionReq true "app版本更换"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/version [POST]
func (a App) Version(c *gin.Context) {
	var para model.AppVersionReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app
// @Summary app更新env
// @Accept json
// @Produce json
// @Param "" body model.AppEnvReq true "app更新env"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/env [POST]
func (a App) Env(c *gin.Context) {
	var para model.AppEnvReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app
// @Summary 更新app的配置文件
// @Accept json
// @Produce json
// @Param "" body model.AppConfigUpdateReq true "更新app的配置文件"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/config/update [POST]
func (a App) ConfigUpdate(c *gin.Context) {
	var para model.AppConfigUpdateReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.App{}.ConfigUpdate(&user, para.Name, para.Data)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app
// @Summary 获取app详细信息
// @Accept json
// @Produce json
// @Param "" app_name path string true "app_name" "获取app详细信息"
// @Success 200 {object} model.AppDetailRes ""
// @Failure 400 {object} model.Result ""
// @Router /app/detail/{app_name} [GET]
func (a App) Detail(c *gin.Context) {
	appName := c.Param("app_name")
	user := GetUser(c)
	detail, err := service.App{}.Detail(&user, appName)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.AppDetailRes{
		Code: 0,
		Msg:  "",
		Data: detail,
	})
}

// @Tags app
// @Summary 获取app配置
// @Accept json
// @Produce json
// @Param "" app_name path string true "app_name" "获取app配置"
// @Success 200 {object} model.AppConfigRes ""
// @Failure 400 {object} model.Result ""
// @Router /app/config/{app_name} [GET]
func (a App) Config(c *gin.Context) {
	appName := c.Param("app_name")
	user := GetUser(c)
	path, config, err := service.App{}.Config(&user, appName)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.AppConfigRes{
		Code: 0,
		Msg:  "",
		Data: &model.AppConfigBase{
			Path: path,
			File: config,
		},
	})
}
