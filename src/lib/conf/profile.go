package conf

import (
	"reflect"
)

// 添加配置步骤
/*
1、修改 conf.ini 文件
2、增加结构体和相应的 Field() 方法
3、修改 InitConf() 函数
4、修改 ConfType() 函数
*/

import (
	"github.com/pkg/errors"
	"lib/ltlog"
)

// longtooth
type varLongTooth struct {
	// 空队列循环时间
	delayDuration int
	// 应用 Key
	appKey string

	// 长牙服务器地址
	ltIp string
	// 长牙服务器端口
	ltPort int

	// 设备 id
	devId int
	// 应用 id
	appId int
	// 最大线程数
	ltMaxThread int
}

func (t *varLongTooth) Field(field string) (data interface{}, err error) {
	ref := reflect.ValueOf(t).Elem()
	v := ref.FieldByName(field)
	if !v.IsValid() {
		return "", invalidFieldErr
	}

	return getInterfaceVal(v), nil
}

func getInterfaceVal(v reflect.Value) (data interface{}) {
	switch v.Type().String() {
	case "int":
		return v.Int()

	case "string":
		return v.String()

	case "bool":
		return v.Bool()

	default:
		return v
	}
}

// controller
type varController struct {
	// 回调 HttpCenter 域名
	responseDomain string
	// 回调 HttpCenter 路径
	responsePath string
}

func (t *varController) Field(field string) (data interface{}, err error) {
	ref := reflect.ValueOf(t).Elem()
	v := ref.FieldByName(field)
	if !v.IsValid() {
		return "", invalidFieldErr
	}

	return getInterfaceVal(v), nil
}

// redis
type varRedis struct {
	protocol string
	host     string
	port     string

	poolMaxIdle         int
	poolMaxActive       int
	poolMaxConnLifetime int
	poolIdleTimeout     int
	poolWait            bool
}

func (t *varRedis) Field(field string) (data interface{}, err error) {
	ref := reflect.ValueOf(t).Elem()
	v := ref.FieldByName(field)
	if !v.IsValid() {
		return "", invalidFieldErr
	}

	return getInterfaceVal(v), nil
}

// main
type varMain struct {
	// phase 1 每次循环处理队列长度
	goroutineParseMessageCount    int
	// phase 1 go 协程数
	phase1GoroutineCount  int
	// phase 2 go 协程数
	phase2GoroutineCount  int
	// phase 2 单次发送消息条数
	combineSentMessageCount int
	// 每次处理的用户数
	handleUserCount int

	userMsgExpire int
	LogLevel      int
}

func (t *varMain) Field(field string) (data interface{}, err error) {
	ref := reflect.ValueOf(t).Elem()
	v := ref.FieldByName(field)
	if !v.IsValid() {
		return "", invalidFieldErr
	}

	return getInterfaceVal(v), nil
}

type varData interface {
	Field(field string) (data interface{}, err error)
}

var (
	invalidFieldErr error = errors.New("Invalid filed error")
	confController        = new(varController)
	confLongTooth         = new(varLongTooth)
	confRedis             = new(varRedis)
	confMain              = new(varMain)
)

func InitConf(iniFile *string) {
	iniConf := NewConf(*iniFile)

	// controller
	iniConf.StringVar(&confController.responseDomain, "controller", "response_domain", "")
	iniConf.StringVar(&confController.responsePath, "controller", "response_path", "")

	// longtooth
	iniConf.IntVar(&confLongTooth.delayDuration, "longtooth", "delay_duration", 200)
	iniConf.StringVar(&confLongTooth.appKey, "longtooth", "app_key", "")
	iniConf.StringVar(&confLongTooth.ltIp, "longtooth", "lt_ip", "")
	iniConf.IntVar(&confLongTooth.ltPort, "longtooth", "lt_port", 53199)
	iniConf.IntVar(&confLongTooth.devId, "longtooth", "dev_id", 2)
	iniConf.IntVar(&confLongTooth.appId, "longtooth", "app_id", 1)
	iniConf.IntVar(&confLongTooth.ltMaxThread, "longtooth", "lt_max_thread", 64)

	// main
	iniConf.IntVar(&confMain.goroutineParseMessageCount, "main", "goroutine_parse_message_count", 1000)
	iniConf.IntVar(&confMain.phase1GoroutineCount, "main", "phase1_goroutine_count", 100)
	iniConf.IntVar(&confMain.phase2GoroutineCount, "main", "phase2_goroutine_count", 100)
	iniConf.IntVar(&confMain.userMsgExpire, "main", "user_msg_expire", 1)
	iniConf.IntVar(&confMain.LogLevel, "main", "log_level", 6)
	iniConf.IntVar(&confMain.combineSentMessageCount, "main", "combine_sent_message_count", 10)
	iniConf.IntVar(&confMain.handleUserCount, "main", "handle_user_count", 100)

	// redis
	iniConf.StringVar(&confRedis.host, "redis", "host", "127.0.0.1")
	iniConf.StringVar(&confRedis.port, "redis", "port", "6379")
	iniConf.StringVar(&confRedis.protocol, "redis", "protocol", "tcp")
	iniConf.IntVar(&confRedis.poolMaxIdle, "redis", "pool_max_idle", 500)
	iniConf.IntVar(&confRedis.poolIdleTimeout, "redis", "pool_idle_timeout", 100)
	iniConf.IntVar(&confRedis.poolMaxActive, "redis", "pool_max_active", 2000)
	iniConf.IntVar(&confRedis.poolMaxConnLifetime, "redis", "pool_max_conn_lifetime", 5)
	iniConf.BoolVar(&confRedis.poolWait, "redis", "pool_wait", true)

	iniConf.Parse()

	// special setting
	// must not move
	ltlog.LogLevel, _ = Int(ConfType("main").Field("LogLevel"))

	ltlog.Println(ltlog.Linfo, "Initial ini config ...")
	ltlog.Println(ltlog.Ldebug, "Ini file path:", *iniFile)
}

func ConfType(ctype string) (data varData) {
	confMap := map[string]varData{
		"controller": confController,
		"longtooth":  confLongTooth,
		"main":       confMain,
		"redis":      confRedis,
	}

	var ok bool
	if data, ok = confMap[ctype]; !ok {
		return nil
	}

	return
}
