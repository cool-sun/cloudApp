package controller

import (
	"errors"
	"github.com/coolsun/cloud-app/utils"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	trans ut.Translator
)

//gin验证错误信息翻译成中文

func init() {
	//注册翻译器
	zh2 := zh.New()
	uni := ut.New(zh2)
	trans, _ = uni.GetTranslator("zh")
	//获取gin的校验器
	validate := binding.Validator.Engine().(*validator.Validate)
	// 注册一个获取json tag的自定义方法
	validate.RegisterTagNameFunc(utils.Field)
	//注册翻译器
	_ = zh_translations.RegisterDefaultTranslations(validate, trans)
}

//Translate 翻译错误信息
func Translate(err error) error {
	var result = ""
	validatorErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}
	for _, err := range validatorErrors {
		result += err.Translate(trans) + ","
	}
	result = result[0 : len(result)-1]
	return errors.New(result)
}
