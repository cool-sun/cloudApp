package utils

import (
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/pkg/errors"
)

//使用validator进行数据格式校验，并将校验结果翻译成中文
var (
	validate *validator.Validate
	trans    ut.Translator
)

func init() {
	//注册翻译器
	zh2 := zh.New()
	uni := ut.New(zh2)
	trans, _ = uni.GetTranslator("zh")

	//获取校验器
	validate = validator.New()
	// 注册一个获取json tag的自定义方法
	validate.RegisterTagNameFunc(Field)

	RegisterValidations(validate)
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

func Validate() *validator.Validate {
	return validate
}

type rv struct {
	tag string //校验的tag标记
	cn  string //中文说明,当校验失败时，将错误信息翻译成中文后就是该值
	fn  validator.Func
}

func RegisterValidations(validate *validator.Validate) {
	rvs := []*rv{&rv{
		tag: "pure-number",
		cn:  "必须是纯数字",
		fn:  isPureNumbers,
	}}
	var err error
	for _, v := range rvs {
		err = registerValidation(v.tag, v.cn, v.fn)
		if err != nil {
			log.Errorf("%+v", err)
		}
	}
	return
}

func isPureNumbers(fl validator.FieldLevel) (b bool) {
	return
}
func registerValidation(tag, cn string, fn validator.Func) (err error) {
	err = validate.RegisterValidation(tag, fn)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	err = validate.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
		return ut.Add(tag, "{0}"+cn, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field())
		return t
	})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
