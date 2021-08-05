package controller

import (
	"github.com/coolsun/cloud-app/utils"
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

var filePathArr = make([]string, 0, 0)

const STATICPATH = "./static"

//因为本站的静态文件地址前缀是/static, artifacthub的静态文件地址前缀也是/static, 所以需要区分是访问本站的静态资源还是artifacthub的静态资源
//判断要访问的静态文件在当前站有没有，如果有就直接返回，否则就转发到artifacthub。
//上述判断方法简单，但是有可能存在一些特殊情况会出错，比如要访问artifacthub的某个静态文件资源在本站存在同名的，就会返回本站的资源了
//后续如果有需要可以从http请求头中寻找差异信息来判断是否转发请求
func exist(path string) bool {
	for _, v := range filePathArr {
		if strings.Contains(path, v) {
			return true
		}
	}
	return false
}

//1.先把所有静态文件路径加载到数组中,避免每次收到请求都要进行io操作
//2.这样做能提升性能，但是也存在问题，如果在程序运行中修改了静态文件资源,程序一定要重启才能生效
//3.为了避免重启，可以定时执行1步骤
func init() {
	filePathArr = utils.GetAllFiles(STATICPATH)
	go loadFilePathArr()
}

func loadFilePathArr() {
	for {
		time.Sleep(time.Second * 5)
		filePathArr = utils.GetAllFiles(STATICPATH)
	}
}

func Static(c *gin.Context) {
	requestURI := c.Request.RequestURI
	if exist(requestURI) {
		c.File("./" + requestURI)
	} else {
		Proxy(c)
	}
}

func ApiProxy(c *gin.Context) {
	reqBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Error(err)
		return
	}
	resp, err := req.Do(c.Request.Method, "https://artifacthub.io"+c.Request.RequestURI, reqBody)
	if err != nil {
		log.Error(err)
		return
	}
	c.Writer.Header().Set("Content-Type", resp.Response().Header.Get("Content-Type"))
	c.Writer.Write(resp.Bytes())
}

func Proxy(c *gin.Context) {
	proxy := httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.Header = c.Request.Header
			request.URL.Scheme = "https"
			request.URL.Host = "artifacthub.io"
			request.Host = "artifacthub.io"
		},
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}
