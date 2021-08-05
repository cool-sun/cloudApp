package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/coolsun/cloud-app/utils/log"
	json "github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

//文件复制
func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

//判断一个接口是不是nil
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}

//三元表达式
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

//生成一定范围内的随机数
func RandNumber(max int) (n int) {
	rand.Seed(time.Now().UnixNano())
	n = rand.Intn(max) + 1 //实际随机生成的数字范围[0,99]
	return
}

//生成数字随机码
func GetValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())
	var sb strings.Builder
	for i := 0; i < width; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

//生成随机的由数字和字母组成的字符串
func RandStringRunes(n int) string {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func InIntSlice(array []int, needle int) bool {
	for _, e := range array {
		if e == needle {
			return true
		}
	}
	return false
}

func InStringSlice(array []string, needle string) bool {
	for _, e := range array {
		if e == needle {
			return true
		}
	}
	return false
}

func GetInt32Pointer(i int32) *int32 {
	return &i
}

func Field(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func ObjectToByte(obj interface{}) (rb []byte) {
	rb, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("%+v", err)
		return
	}
	return
}

func ObjectToString(obj interface{}) (s string) {
	rb, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("%+v", err)
		return
	}
	s = string(rb)
	return
}

func StringToInt64(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Errorf("%+v", err)
		return 0
	}
	return i
}

//具体类型的slice装interface slice
func ToSlice(arr interface{}) []interface{} {
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		log.Error("toslice arr not slice")
	}
	l := v.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = v.Index(i).Interface()
	}
	return ret
}

//md5加密
func MD5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//获取指定目录下的所有文件,包含子目录下的文件
func GetAllFiles(dirPth string) (files []string) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		log.Error(err)
		return
	}

	for _, fi := range dir {
		p := path.Join(dirPth, fi.Name())
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, p)
			GetAllFiles(p)
		} else {
			files = append(files, p)
		}
	}

	// 读取子目录下文件
	for _, table := range dirs {
		temp := GetAllFiles(table)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return
}
