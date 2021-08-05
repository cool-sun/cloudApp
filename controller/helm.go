package controller

import (
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/service"
	"github.com/gin-gonic/gin"
)

type Helm struct {
}

// @Tags helm release
// @Summary helm release创建
// @Accept json
// @Produce json
// @Param "" body model.HelmReleaseCreateReq true "helm release创建"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /helm/release/create [POST]
func (h Helm) Create(c *gin.Context) {
	var para model.HelmReleaseCreateReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.Helm{}.ReleaseCreate(user.Name, &para)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags helm release
// @Summary helm release更新
// @Accept json
// @Produce json
// @Param "" body model.HelmReleaseUpdateReq true "helm release更新"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /helm/release/Update [POST]
func (h Helm) Update(c *gin.Context) {
	//var para model.HelmReleaseUpdateReq
	//if err := c.ShouldBind(&para); err != nil {
	//	ErrorResponse(c, err)
	//	return
	//}
	//user := GetUser(c)
	//err := service.Helm{}.ReleaseUpdate(user.Name, &para)
	//if err != nil {
	//	ErrorResponse(c, err)
	//	return
	//}
	SuccessResponse(c, &model.Result{})
}

// @Tags helm release
// @Summary helm release列表信息
// @Accept json
// @Produce json
// @Param "" body model.HelmReleaseListReq true "helm release列表信息"
// @Success 200 {object} model.HelmReleaseListRes ""
// @Failure 400 {object} model.Result ""
// @Router /helm/release/list [POST]
func (h Helm) List(c *gin.Context) {
	var para model.HelmReleaseListReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	//开发者只能查看自己名下的app
	if user.RoleId == 2 || para.UserName == "" {
		para.UserName = user.Name
	}
	list, count, err := service.Helm{}.ReleaseList(para.UserName, &para)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.HelmReleaseListRes{
		Code: 0,
		Msg:  "",
		Data: &model.HelmReleaseList{
			Count: count,
			List:  list,
		},
	})
}

// @Tags helm release
// @Summary helm release删除
// @Accept json
// @Produce json
// @Param "" body model.HelmReleaseReq true "helm release删除"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /helm/release/delete [POST]
func (h Helm) Delete(c *gin.Context) {
	var para model.HelmReleaseReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.Helm{}.ReleaseDelete(user.Name, para.Name)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{})
}

// @Tags helm release
// @Summary 获取helm release的values
// @Accept json
// @Produce json
// @Param "" body model.HelmReleaseReq true "获取helm release的values"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /helm/release/values [POST]
func (h Helm) Values(c *gin.Context) {
	var para model.HelmReleaseReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	data, err := service.Helm{}.GetReleaseValues(user.Name, para.Name)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.Result{Data: data})
}
