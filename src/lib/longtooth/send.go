package longtooth

/*
#cgo CFLAGS: -I../../include
#cgo LDFLAGS: -L../../clib -lltc5-unix
#include "longtooth.h"
#include "tuyyo.h"

int cb_lt_start(int64_t devid, int appid, const char* appkey, int64_t machineid) {
	return lt_start(devid, appid, appkey, machineid, LtStartCallback);
}

int cb_lt_request(lt_tunnel ltt,
				const char* ltid_str,
				const char* service_str,
				int lt_data_type,
				const char* args,
				size_t argslen,
				char* attachment) {
	return lt_request(ltt,
		ltid_str,
		service_str,
		lt_data_type,
		args,
		argslen,
		attachment,
		NULL,
		(lt_service_response_handler)LtRequestCallback);
}
*/
import "C"
import (
	"unsafe"
	"lib/conf"
	"time"
	"lib/ltlog"
)

// 启动 longtooth 模块
func Start() {
	var (
		devId, _         = conf.Int(conf.ConfType("longtooth").Field("devId"))
		appId, _         = conf.Int(conf.ConfType("longtooth").Field("appId"))
		ltMaxThread, _   = conf.Int(conf.ConfType("longtooth").Field("ltMaxThread"))
		appKey, _        = conf.String(conf.ConfType("longtooth").Field("appKey"))
		ltIp, _          = conf.String(conf.ConfType("longtooth").Field("ltIp"))
		ltPort, _        = conf.Int(conf.ConfType("longtooth").Field("ltPort"))
		intDelayDuration, _ = conf.Int(conf.ConfType("longtooth").Field("delayDuration"))

		cDevId     = C.int64_t(devId)
		cAppId     = C.int(appId)
		cMachineId = C.int64_t(machineId)
		cAppKey    = C.CString(appKey)
		cLtPort    = C.int(ltPort)
	)

	// 控制标志
	ltServiceHasRunFlag = make(chan bool, 1)
	requestHandleHasRunFlag = make(chan bool, 1)
	RunProc = NewRunningProcess()

	delayDuration = time.Duration(intDelayDuration) * time.Millisecond
	phase1GoroutineCount, _ = conf.Int(conf.ConfType("main").Field("phase1GoroutineCount"))
	phase2GoroutineCount, _ = conf.Int(conf.ConfType("main").Field("phase2GoroutineCount"))
	goroutineParseMessageCount, _ = conf.Int(conf.ConfType("main").Field("goroutineParseMessageCount"))
	combineSentMessageCount, _ = conf.Int(conf.ConfType("main").Field("combineSentMessageCount"))
	handleUserCount, _ = conf.Int(conf.ConfType("main").Field("handleUserCount"))

	cLtIp := C.CString(ltIp)

	defer func() {
		C.free(unsafe.Pointer(cAppKey))
		C.free(unsafe.Pointer(cLtIp))
	}()

	C.lt_thread_max(C.int(ltMaxThread))
	C.lt_lan_mode_set(false)
	C.lt_register_host_set(cLtIp, cLtPort)
	ret := int(C.cb_lt_start(cDevId, cAppId, cAppKey, cMachineId))
	if -1 == ret {
		ltlog.Fatalf("Longtooth start fail, lt_start return: %d", ret)
	} else {
		ltlog.Printf(ltlog.Ldebug, "dev_id[%d] app_id[%d] machine_id[%d] lt_ip[%s] lt_port[%d]",
			devId, appId, machineId, ltIp, ltPort)
	}

	// 启动 request handler
	requestHandleHasRunFlag <- true

	// 监听服务状态
	<-ltServiceHasRunFlag
}

func Close() {
	ltServiceHasRunFlag <- true

	close(ltServiceHasRunFlag)
	close(requestHandleHasRunFlag)
}
