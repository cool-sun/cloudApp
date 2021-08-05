package controller

import (
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/service"
	"github.com/gin-gonic/gin"
)

type Snapshot struct {
}

// @Tags app快照
// @Summary app创建存储快照
// @Accept json
// @Produce json
// @Param "" body model.SnapshotCreateReq true "app创建存储快照"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/snapshot/create [POST]
func (a Snapshot) Create(c *gin.Context) {
	var para model.SnapshotCreateReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.Snapshot{}.Create(user.Name, para.ShowName, para.AppName)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags app快照
// @Summary 获取快照列表
// @Accept json
// @Produce json
// @Param "" body model.SnapshotListReq true "获取快照列表"
// @Success 200 {object} model.SnapshotListRes ""
// @Failure 400 {object} model.Result ""
// @Router /app/snapshot/list [POST]
func (a Snapshot) List(c *gin.Context) {
	var para model.SnapshotListReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	list, count, err := service.Snapshot{}.ListInfo(para.Search, para.UserName, para.AppName, para.Current, para.PageSize)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.SnapshotListRes{
		Code: 0,
		Msg:  "",
		Data: &model.SnapshotListBase{
			Count: int(count),
			List:  list,
		},
	})
}

// @Tags app快照
// @Summary 删除快照
// @Accept json
// @Produce json
// @Param "" body model.SnapshotDeleteReq true "删除快照"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /app/snapshot/delete [POST]
func (a Snapshot) Delete(c *gin.Context) {
	var para model.SnapshotDeleteReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.Snapshot{}.Delete(user.Name, para.Id)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}
