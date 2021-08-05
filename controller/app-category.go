package controller

import (
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/service"
	"github.com/gin-gonic/gin"
)

type AppCategory struct {
}

// @Tags app类别
// @Summary 获取所有支持的App类别模板
// @Accept json
// @Produce json
// @Success 200 {object} model.AppCategoryTmplRes "获取所有支持的App类别模板"
// @Failure 400 {object} model.Result ""
// @Router /app-category/template [GET]
func (ac AppCategory) HandleAppCategoryTmpl(c *gin.Context) {
	tmpl := service.AppCategory{}.GetTmpl()
	SuccessResponse(c, &model.AppCategoryTmplRes{
		Data: tmpl,
	})
}

// @Tags app类别
// @Summary 获取某个类型的app的所有支持的tag号，即版本号
// @Accept json
// @Produce json
// @Param "" body model.AppLittleCategoryTagReq true "app创建"
// @Success 200 {object} model.AppLittleCategoryTagRes ""
// @Failure 400 {object} model.Result ""
// @Router /app-category/tag/list [POST]
func (ac AppCategory) HandleAppLittleCategoryTag(c *gin.Context) {
	var para model.AppLittleCategoryTagReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	tags, count, err := service.AppCategory{}.GetTagsList(para.Current, para.PageSize, para.LittleCategory, para.Search)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.AppLittleCategoryTagRes{
		Data: &model.AppLittleCategoryTagBase{
			Count: count,
			Tags:  tags,
		},
	})
}

//todo env有重复的

//todo 监听和收集k8s事件，用途如下:
//并发送给相应的用户，方便用户根据相关信息采取相应的操作
//事件历史数据查询
//事件按类别分类统计，可视化展示，如 https://cdn.nlark.com/yuque/0/2020/png/347081/1579242500156-ff8cc069-e7d9-443a-82b5-0a6c2b371d07.png#align=left&display=inline&height=2033&name=image.png&originHeight=2033&originWidth=1826&size=554101&status=done&style=none&width=1826
//实现逻辑如下：
//k8s包中监听event变更事件，发送给该包中的一个 EventChan 对象；
//monitor包中读取 EventChan 的数据，并写入到数据库；
//EventChan对象为无缓冲的channel，之所以不用有缓冲的，主要是防止程序突然奔溃导致缓冲中的数据丢失
