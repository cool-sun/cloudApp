package service

import (
	"context"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/builder"
	"github.com/coolsun/cloud-app/utils/github/didi/gendry/scanner"
	"github.com/pkg/errors"
)

type User struct {
}

//删除用户
func (u User) Delete(users []*model.UserDeleteBase) (err error) {
	for _, v := range users {
		var exist bool
		user := &model.User{
			Name: v.Name,
		}
		exist, err = engine.RowGet(user)
		if err != nil {
			return
		}
		if !exist {
			err = errors.New(model.E10201)
			return
		}
		session := engine.NewSession()
		defer session.Close()
		session.Begin()
		_, err = session.ID(user.Id).Update(&model.User{IsDelete: 1})
		if err != nil {
			err = errors.WithStack(err)
			return
		}
		if v.IsCleanApp {
			_, err = session.Where("user_name = ?", user.Name).Update(&model.App{IsDelete: 1})
			if err != nil {
				err = errors.WithStack(err)
				return
			}
			err = plat.App().DeleteUser(v.Name)
			if err != nil {
				return
			}
		}
		session.Commit()
	}

	return
}

func (u User) Login(name, password string) (user *model.User, err error) {
	user = &model.User{Name: name, Password: utils.MD5String(password)}
	b, err := engine.RowGet(user)
	if err != nil {
		return
	}
	if !b {
		err = errors.New(model.E10001)
		return
	}
	return
}

//创建用户，分两步
//1.往数据库插入用户信息
//2.在k8s集群中创建对应的命名空间
func (u User) Create(name, phone, email, password string, role int, resourceQuota *model.ResourceQuota) (err error) {
	user := &model.User{
		Name: name,
	}
	exist, err := engine.Get(user)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if exist {
		err = errors.New(model.E10002)
		return
	}
	newUser := &model.User{
		Name:          name,
		AvatarNum:     utils.RandNumber(210),
		Password:      utils.MD5String(password),
		RoleId:        role,
		ResourceQuota: resourceQuota,
		Email:         email,
		Phone:         phone,
	}
	session := engine.NewSession()
	defer session.Close()
	session.Begin()
	_, err = session.Insert(newUser)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = plat.App().CreateUser(name, resourceQuota)
	if err != nil {
		return
	}
	session.Commit()
	return
}

func (u User) Update(name, phone, email, password string, role int, resourceQuota *model.ResourceQuota) (err error) {
	session := engine.NewSession()
	defer session.Close()
	session.Begin()
	effect, err := session.Where("name = ? ", name).Update(&model.User{
		Password:      utils.MD5String(password),
		RoleId:        role,
		ResourceQuota: resourceQuota,
		Email:         email,
		Phone:         phone,
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if effect == 0 {
		err = errors.New(model.E10201)
		return
	}
	err = plat.App().CreateUser(name, resourceQuota)
	if err != nil {
		return
	}
	session.Commit()
	return
}

func (u User) Edit(id int, phone, email string) (err error) {
	_, err = engine.ID(id).Update(&model.User{
		Phone:    phone,
		Email:    email,
		IsDelete: 0,
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}

func (u User) List(pageNum, pageSize, roleId int, search string) (list []*model.User, count int64, err error) {
	skipLines := (pageNum - 1) * pageSize
	likeValue := "%" + search + "%"
	where := map[string]interface{}{
		"_or": []map[string]interface{}{
			{
				"name like": likeValue,
			},
			{
				"phone like": likeValue,
			},
			{
				"email like": likeValue,
			},
		},
		"role_id":   roleId,
		"is_delete": 0,
		"_orderby":  "create_time desc",
		"_limit":    []uint{uint(skipLines), uint(pageSize)},
	}
	finalWhere := builder.OmitEmpty(where, []string{"role_id", "_or"})
	selectFields := []string{"id", "name", "avatar_num", "phone", "email", "role_id", "resource_quota", "create_time", "modify_time"}
	cond, values, err := builder.BuildSelect("user", finalWhere, selectFields)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	rows, err := engine.DB().Query(cond, values...)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	list = make([]*model.User, 0, 0)
	err = scanner.Scan(rows, &list)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	delete(finalWhere, "_limit") //如果不删除这个key的话,当查询第2页的时候，count会变成0
	rlt, err := builder.AggregateQuery(context.TODO(), engine.DB().DB, "user", finalWhere, builder.AggregateCount("*"))
	count = rlt.Int64()
	return
}

//判断用户是否存在
func (u User) Exist(name string) (b bool) {
	b, _ = engine.Exist(&model.User{Name: name})
	return
}
