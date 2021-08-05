package controller

import (
	"bytes"
	"context"
	"encoding/gob"
	"github.com/coolsun/cloud-app/model"
	"github.com/coolsun/cloud-app/utils"
	"github.com/coolsun/cloud-app/utils/log"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	RESERROR       = "RESERROR"
	USERSESSIONKEY = "user"
)

func TimeoutMiddleware(timeout time.Duration) func(c *gin.Context) {
	return func(c *gin.Context) {

		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)

		defer func() {
			// check if context timeout was reached
			if ctx.Err() == context.DeadlineExceeded {

				// write response and abort the request
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				c.Abort()
			}

			//cancel to clear resources after finished
			cancel()
		}()

		// replace request with context wrapped request
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func init() {
	gob.Register(model.User{})
}

func UnauthorizedResponse(c *gin.Context, err error) {
	err = Translate(err)
	c.Set(RESERROR, err)
	obj := &model.Result{
		Msg: err.Error(),
	}
	c.JSON(http.StatusUnauthorized, obj)
	return
}

func ErrorResponse(c *gin.Context, err error, code ...int) {
	err = Translate(err)
	c.Set(RESERROR, err)
	obj := &model.Result{
		Msg: err.Error(),
	}
	if len(code) >= 1 {
		obj.Code = code[0]
	}
	c.JSON(http.StatusBadRequest, obj)
	return
}

func SuccessResponse(c *gin.Context, obj interface{}) {
	if utils.IsNil(obj) {
		obj = &model.Result{}
	}
	c.JSON(http.StatusOK, obj)
	return
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

//自定义gin路由日志
//注意，这里每个handler应该只做一次日志打印，在并发情况下会打串了
type HandlerInfo struct {
	Id          string    `json:"id"`
	LatencyTime float64   `json:"latency_time"`
	Request     *Request  `json:"request"`
	Response    *Response `json:"response"`
}
type Request struct {
	Header *RequestHeader `json:"header"`
	Body   JsonRawString  `json:"body"`
}
type Response struct {
	Header *ResponseHeader `json:"header"`
	Body   JsonRawString   `json:"body"`
}

type JsonRawString string

func (m JsonRawString) MarshalJSON() ([]byte, error) {
	if len(m) == 0 {
		return []byte("null"), nil
	}
	return []byte(m), nil
}

type RequestHeader struct {
	Method      string `json:"method"`
	RequestURI  string `json:"request_uri"`
	ContentType string `json:"content_type"`
	ClientIP    string `json:"client_ip"`
	UserAgent   string `json:"user_agent"`
}
type ResponseHeader struct {
	Status int `json:"status"`
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 打印路由日志
		startTime := time.Now()
		reqMethod := c.Request.Method
		reqURI := c.Request.RequestURI
		reqBody, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Errorf("%+v", err)
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()
		statusCode := c.Writer.Status()
		latencyTime := time.Since(startTime)

		var id = utils.GetValidateCode(8)
		handlerInfo := &HandlerInfo{
			Id:          id,
			LatencyTime: float64(latencyTime) / 1000 / 1000,
			Request: &Request{
				Header: &RequestHeader{
					Method:      reqMethod,
					RequestURI:  reqURI,
					ContentType: c.ContentType(),
					ClientIP:    c.ClientIP(),
					UserAgent:   c.Request.UserAgent(),
				},
				Body: JsonRawString(reqBody),
			},
			Response: &Response{
				Header: &ResponseHeader{
					Status: statusCode,
				},
				Body: JsonRawString(blw.body.Bytes()),
			},
		}
		if strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
			if c.Writer.Status() < 400 {
				log.InfoJson(handlerInfo)
			} else if c.Writer.Status() >= 400 {
				err, _ := c.Get(RESERROR)
				log.ErrorJson(handlerInfo)
				log.Errorf("handler id is %s %+v ,", id, err)
			}
		} else {
			if c.Writer.Status() < 400 {
				log.InfoJson(handlerInfo.Request.Header)
			} else if c.Writer.Status() >= 400 {
				log.ErrorJson(handlerInfo)
			}
		}
	}
}

func CheckContentType(c *gin.Context) {
	contentType := c.ContentType()
	if contentType == "" || strings.Contains(contentType, "application/json") {
		c.Next()
		return
	}
	ErrorResponse(c, errors.New(model.E10000))
	c.Abort()
}

func CheckLogin(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get(USERSESSIONKEY)
	if info == nil && !(os.Getenv("mode") == "debug") {
		UnauthorizedResponse(c, errors.New(model.E10200))
		c.Abort()
		return
	}
	c.Next()
}

func GetUser(c *gin.Context) (user model.User) {
	session := sessions.Default(c)
	info := session.Get(USERSESSIONKEY)
	if os.Getenv("mode") == "debug" && utils.IsNil(info) {
		user = model.User{
			Name: "cloud-app",
		}
		return
	}
	user = info.(model.User)
	return
}
