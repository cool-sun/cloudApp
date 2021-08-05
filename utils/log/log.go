package log

import (
	"github.com/coolsun/cloud-app/utils/github/cihub/seelog"
	json "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"reflect"
)

const LoggerConfig = `<seelog levels="trace,debug,info,warn,error,critical" type="sync">
    <outputs formatid="main">
        <!-- 对控制台输出的Log按级别分别用颜色显示。6种日志级别我仅分了三组颜色，如果想每个级别都用不同颜色则需要简单修改即可 -->
        <filter levels="trace,info">
            <console formatid="colored-default"/>
            <!-- 将日志输出到磁盘文件，按文件大小进行切割日志，单个文件最大10M，最多5个日志文件 -->
            <rollingfile formatid="main" type="size" filename="./logs/info.log" maxsize="10485760"                                     maxrolls="100"/>
        </filter>
        <filter levels="warn">
            <console formatid="colored-warn"/>
            <rollingfile formatid="main" type="size" filename="./logs/warn.log" maxsize="10485760"                                     maxrolls="100"/>
        </filter>
        <filter levels="error,critical">
            <console formatid="colored-error"/>
            <rollingfile formatid="main" type="size" filename="./logs/error.log" maxsize="10485760"                                     maxrolls="100"/>
        </filter>
		<filter levels="debug,critical">
            <console formatid="colored-debug"/>
            <rollingfile formatid="main" type="size" filename="./logs/debug.log" maxsize="10485760"                                     maxrolls="100"/>
        </filter>
    </outputs>
    <formats>
        <format id="colored-default" format="%EscM(38)%Date %Time [%LEV] %RelFile:%Line | %Msg%n%EscM(0)"/>
        <format id="colored-warn" format="%EscM(33)%Date %Time [%LEV] %RelFile:%Line | %Msg%n%EscM(0)"/>
        <format id="colored-debug" format="%EscM(34)%Date %Time [%LEV] %RelFile:%Line | %Msg%n%EscM(0)"/>
        <format id="colored-error" format="%EscM(31)%Date %Time [%LEV] %RelFile:%Line | %Msg%n%EscM(0)"/>
        <format id="main" format="%Date %Time [%LEV] %RelFile:%Line | %Msg%n"/>
    </formats>
</seelog>`

var NULLJSON = []byte("null")

func init() {
	//日志模块初始化
	logger, _ := seelog.LoggerFromConfigAsString(LoggerConfig)
	_ = seelog.ReplaceLogger(logger)
}

func init() {
	defaultLogger = &Log{}
}

func SetLogger(logger Logger) error {
	if logger == nil {
		return errors.New("logger can not be nil")
	}
	defaultLogger = logger
	return nil
}

type Logger interface {
	Infof(format string, params ...interface{})
	Info(v ...interface{})
	Errorf(format string, params ...interface{})
	Error(v ...interface{})
	Warnf(format string, params ...interface{})
	Warn(v ...interface{})
	Debugf(format string, params ...interface{})
	Debug(v ...interface{})
	InfoJson(v interface{})
	ErrorJson(v interface{})
	WarnJson(v interface{})
	DebugJson(v interface{})
}

var defaultLogger Logger

type Log struct {
}

func Infof(format string, params ...interface{}) {
	defaultLogger.Infof(format, params...)
}

func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

func Errorf(format string, params ...interface{}) {
	defaultLogger.Errorf(format, params...)
}

func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}

func Warnf(format string, params ...interface{}) {
	defaultLogger.Warnf(format, params...)
}

func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

func Debugf(format string, params ...interface{}) {
	defaultLogger.Debugf(format, params...)
}

func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

func InfoJson(v interface{}) {
	defaultLogger.InfoJson(v)
}

func ErrorJson(v interface{}) {
	defaultLogger.ErrorJson(v)
}

func WarnJson(v interface{}) {
	defaultLogger.WarnJson(v)
}

func DebugJson(v interface{}) {
	defaultLogger.DebugJson(v)
}

func (l Log) Infof(format string, params ...interface{}) {
	seelog.Infof(format, params...)
}
func (l Log) Errorf(format string, params ...interface{}) {
	_ = seelog.Errorf(format, params...)
}
func (l Log) Warnf(format string, params ...interface{}) {
	_ = seelog.Warnf(format, params...)
}
func (l Log) Debugf(format string, params ...interface{}) {
	seelog.Debugf(format, params...)
}

func (l Log) Info(v ...interface{}) {
	seelog.Info(v...)
}
func (l Log) Error(v ...interface{}) {
	_ = seelog.Error(v...)
}
func (l Log) Debug(v ...interface{}) {
	seelog.Debug(v...)
}
func (l Log) Warn(v ...interface{}) {
	_ = seelog.Warn(v...)
}

func (l Log) InfoJson(v interface{}) {
	seelog.Info(convert2Json(v))
}
func (l Log) ErrorJson(v interface{}) {
	_ = seelog.Error(convert2Json(v))
}
func (l Log) DebugJson(v interface{}) {
	seelog.Info(convert2Json(v))
}
func (l Log) WarnJson(v interface{}) {
	_ = seelog.Error(convert2Json(v))
}

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

func convert2Json(v interface{}) (s string) {
	var err error
	var b []byte
	switch v.(type) {
	case []byte:
		b = v.([]byte)
	default:
		if IsNil(v) {
			b = NULLJSON
		} else {
			b, err = json.Marshal(v)
			if err != nil {
				Error(err)
				println(v)
			}
		}
	}
	if len(b) == 0 {
		b = NULLJSON
		return
	}
	s = string(b)
	return
}
