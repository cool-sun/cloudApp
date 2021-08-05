package controller

import (
	"fmt"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type User struct {
}

// @Tags 用户
// @Summary 管理员添加用户
// @Accept json
// @Produce json
// @Param "" body model.UserRegisterReq true "管理员添加用户"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /user/register [POST]
func (u User) Register(c *gin.Context) {
	//todo 参数校验，会根据用户名创建相应的命名空间,所以用户名不能太长，限定只能由数字和小写组成
	var para model.UserRegisterReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	err := service.User{}.Create(para.Name, para.Phone, para.Email, para.PassWord, para.RoleId, para.ResourceQuota)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, nil)
}

// @Tags 用户
// @Summary 管理员编辑用户
// @Accept json
// @Produce json
// @Param "" body model.UserRegisterReq true "管理员编辑用户"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /user/manager-edit [POST]
func (u User) ManagerEdit(c *gin.Context) {
	//todo 参数校验，会根据用户名创建相应的命名空间,所以用户名不能太长，限定只能由数字和小写组成
	var para model.UserManagerEditReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	err := service.User{}.Update(para.Name, para.Phone, para.Email, para.PassWord, para.RoleId, para.ResourceQuota)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, nil)
}

// @Tags 用户
// @Summary 用户登录
// @Accept json
// @Produce json
// @Param "" body model.UserLoginReq true "用户登录"
// @Success 200 {object} model.UserLoginRes ""
// @Failure 400 {object} model.Result ""
// @Router /user/login [POST]
func (u User) Login(c *gin.Context) {
	var para model.UserLoginReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user, err := service.User{}.Login(para.Name, para.Password)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	session := sessions.Default(c)
	session.Set(USERSESSIONKEY, *user)
	err = session.Save()
	if err != nil {
		ErrorResponse(c, errors.WithStack(err))
		return
	}
	SuccessResponse(c, &model.UserLoginRes{
		Code: 0,
		Msg:  "",
		Data: &model.UserLoginResBase{
			User:      user,
			AvatarUrl: fmt.Sprintf("/static/avatar/%v.png", user.AvatarNum),
		},
	})
}

// @Tags 用户
// @Summary 用户登出
// @Accept json
// @Produce json
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /user/logout [POST]
func (u User) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(USERSESSIONKEY)
	err := session.Save()
	if err != nil {
		ErrorResponse(c, errors.WithStack(err))
		return
	}
	SuccessResponse(c, nil)
}

// @Tags 用户
// @Summary 删除用户
// @Accept json
// @Produce json
// @Param "" body model.UserDeleteReq true "删除用户"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /user/delete [POST]
func (u User) Delete(c *gin.Context) {
	var para model.UserDeleteReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	err := service.User{}.Delete(para.Users)
	if err != nil {
		ErrorResponse(c, err)
		return
	}

	SuccessResponse(c, nil)
}

// @Tags 用户
// @Summary 获取用户列表
// @Accept json
// @Produce json
// @Param "" body model.UserListReq true "获取用户列表"
// @Success 200 {object} model.UserListRes ""
// @Failure 400 {object} model.Result ""
// @Router /user/list [POST]
func (u User) List(c *gin.Context) {
	var para model.UserListReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}

	list, count, err := service.User{}.List(para.Current, para.PageSize, para.RoleId, para.Search)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, &model.UserListRes{
		Code: 0,
		Msg:  "",
		Data: &model.UserListBase{
			Count: count,
			List:  list,
		},
	})
}

// @Tags 用户
// @Summary 编辑用户信息
// @Accept json
// @Produce json
// @Param "" body model.UserEditReq true "编辑用户信息,适用于用户自己编辑自己的信息"
// @Success 200 {object} model.Result ""
// @Failure 400 {object} model.Result ""
// @Router /user/edit [POST]
func (u User) Edit(c *gin.Context) {
	var para model.UserEditReq
	if err := c.ShouldBind(&para); err != nil {
		ErrorResponse(c, err)
		return
	}
	user := GetUser(c)
	err := service.User{}.Edit(user.Id, para.Phone, para.Email)
	if err != nil {
		ErrorResponse(c, err)
		return
	}
	SuccessResponse(c, nil)
}
