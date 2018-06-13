package longtooth

import (
	"time"
	"lib/tool"
	"sync"
)

var (
	// Global Parameters
	ServiceName string = "longtooth"
	ModeFlag    string = ""
	machineId          = tool.GetIntDate()

	// setting
	ltServiceHasRunFlag        chan bool
	requestHandleHasRunFlag    chan bool
	phase1GoroutineCount       int
	phase2GoroutineCount       int
	goroutineParseMessageCount int
	combineSentMessageCount    int
	handleUserCount            int

	// 消息队列循环时间
	delayDuration time.Duration
	msgHashCache  = &ltHashCache{cache: make(map[string]*communication)}

	// 统计，正在发送消息的协程
	RunProc *runningProcess

	// request singal
	reqWg sync.WaitGroup

	// 消息分隔符
	// escape character
	msgSeperator rune = 27
)

/*
Longtooth Event Desp
*/
var ltEventTag = map[int]string{
	0x20002: "EVENT_LONGTOOTH_ACTIVATED",
	0x28002: "EVENT_LONGTOOTH_TIMEOUT",
	0x28003: "EVENT_LONGTOOTH_UNREACHABLE",
	0x40001: "EVENT_SERVICE_NOT_EXIST",
	0x40004: "EVENT_SERVICE_TIMEOUT",
}

var procTag = map[int]string{
	PROC_RUN:  "PROC_RUN",
	PROC_NIL:  "PROC_NIL",
	PROC_STOP: "PROC_STOP",
}

// 运行状态
const (
	PROC_STOP int = 0
	PROC_RUN  int = 1
	PROC_NIL  int = -1

	STAGE_PARSE string = "phase1"
	STAGE_SEND  string = "phase2"
)
