package router

import (
	"bytes"
	"github.com/coolsun/cloud-app/controller"
	"github.com/coolsun/cloud-app/docs"
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"github.com/thinkerou/favicon"
	"net/http"
	"net/http/httputil"
	"path"
	"strings"
	"time"
)

var cacheSctore = cache.New(10*time.Minute, 60*time.Minute)

type GinPanicWriter struct {
}

func (w GinPanicWriter) Write(p []byte) (n int, err error) {
	log.Error("gin panic recovered :", string(p))
	return 0, err
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func Register() *gin.Engine {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Infof("%-6s %-25s --> %s (%d handlers)", httpMethod, absolutePath, handlerName, nuHandlers)
	}
	r := gin.New()

	//用自己的日志库收集记录panic日志
	r.Use(gin.RecoveryWithWriter(&GinPanicWriter{}))

	//路由这块代码的顺序不能变！
	//路由这块代码的顺序不能变！
	//路由这块代码的顺序不能变！

	//session组件
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("session", store))

	r.Use(controller.Logger())
	r.Use(cors.Default())
	//r.Use(controller.TimeoutMiddleware(time.Minute * 5))
	r.Use(favicon.New("./" + path.Join(controller.STATICPATH, "favicon.ico")))
	//swagger文档托管
	var v2BasePath = "/api/v2"
	docs.SwaggerInfo.Title = "CloudApp System API Document"
	docs.SwaggerInfo.Description = "This is CloudApp System API Document"
	docs.SwaggerInfo.Version = "2.0"
	docs.SwaggerInfo.BasePath = v2BasePath
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiV2 := r.Group(v2BasePath,
		controller.CheckContentType,
		controller.CheckLogin,
	)

	//首页
	view := controller.View{}
	r.GET("/index", view.Index)

	//下面这个路由需要区分一下是访问本站的静态资源还是artifacthub的静态资源
	r.Any("/static/*action", controller.Static)

	//app类别
	appCategory := controller.AppCategory{}
	appCategoryGroup := apiV2.Group("/app-category")
	appCategoryGroup.GET("/template", appCategory.HandleAppCategoryTmpl)
	appCategoryGroup.POST("/tag/list", appCategory.HandleAppLittleCategoryTag)
	//app
	app := controller.App{}
	appGroup := apiV2.Group("/app")
	appGroup.POST("/create", app.Create)
	appGroup.POST("/delete", app.Delete)
	appGroup.POST("/scale", app.Scale)
	appGroup.POST("/restart", app.Restart)
	appGroup.POST("/stop", app.Stop)
	appGroup.POST("/start", app.Start)
	appGroup.POST("/replicas", app.Replicas)
	appGroup.POST("/list", app.List)
	appGroup.POST("/open", app.Open)
	appGroup.POST("/version", app.Version)
	appGroup.POST("/env", app.Env)
	appGroup.POST("/config/update", app.ConfigUpdate)
	appGroup.GET("/config/:app_name", app.Config)
	appGroup.GET("/detail/:app_name", app.Detail)
	appGroup.POST("/restore", app.Restore)
	appGroup.GET("/sc", app.SC)

	//快照
	snapshot := controller.Snapshot{}
	snapshotGroup := apiV2.Group("/app/snapshot")
	snapshotGroup.POST("/create", snapshot.Create)
	snapshotGroup.POST("/list", snapshot.List)
	snapshotGroup.POST("/delete", snapshot.Delete)

	//helm安装的app
	helm := controller.Helm{}
	helmGroup := apiV2.Group("/helm/release")
	helmGroup.POST("/create", helm.Create)
	helmGroup.POST("/update", helm.Update)
	helmGroup.POST("/list", helm.List)
	helmGroup.POST("/delete", helm.Delete)
	helmGroup.POST("/values", helm.Values)

	//用户
	user := controller.User{}
	userGroup := apiV2.Group("/user")
	userGroup.POST("/register", user.Register)
	userGroup.POST("/manager-edit", user.ManagerEdit)
	r.POST("/api/v2/user/login", user.Login)
	userGroup.POST("/logout", user.Logout)
	userGroup.POST("/delete", user.Delete)
	userGroup.POST("/list", user.List)
	userGroup.POST("/edit", user.Edit)

	r.NoRoute(func(c *gin.Context) {
		proxy := httputil.ReverseProxy{
			Director: func(request *http.Request) {
				request.Header = c.Request.Header
				request.URL.Scheme = "https"
				request.URL.Host = "artifacthub.io"
				request.Host = "artifacthub.io"
			},
		}
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		proxy.ServeHTTP(blw, c.Request)
		resDataString := string(blw.body.Bytes())
		resDataString = strings.ReplaceAll(resDataString, `<meta charset="utf-8"/>`, `<meta charset="utf-8"/><link href="/static/style/bootstrap.min.css" rel="stylesheet">`)
		resDataString = strings.ReplaceAll(resDataString, "</body>", `
<link rel="stylesheet" href="/static/style.css">
<link rel="stylesheet" href="/static/style/bootoast.css">
<link href="/static/style/jsoneditor.min.css" rel="stylesheet" type="text/css">
<script src="/static/jquery.min.js"></script>
<script src="/static/jsoneditor.min.js"></script>
<script src="/static/bootstrap.min.js"></script>
<script src="/static/bootoast.js"></script>
<script src="/static/custom.js"></script>
</body>`)
		c.Writer.Write([]byte(resDataString))
		return
	})
	return r
}
