package longtooth

/*
#include "longtooth.h"

int cb_lt_request(lt_tunnel ltt,
				const char* ltid_str,
				const char* service_str,
				int lt_data_type,
				const char* args,
				size_t argslen,
				char* attachment);
*/
import "C"
import (
	"unsafe"
	"time"
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"encoding/json"

	"lib/cache"
	"lib/tool"
	"strings"
	"lib/ltlog"
)

//export LtStartCallback
func LtStartCallback(event C.int, cLtid *C.char, cSrvStr *C.char, cMsg *C.char,
	cMsgLen C.size_t, cMsgCacheKey *C.char, handler C.lt_attachment_handler) {
	ltlog.Printf(ltlog.Ldebug, "----> Longtooth event: 0x%x\n", int(event))

	defer func() {
		if err := recover(); err != nil {
			ltlog.Println(ltlog.Lerror, err)
		}
	}()

	if (int(event) == int(C.EVENT_LONGTOOTH_ACTIVATED)) {
		// 长牙已完全启动
		<-requestHandleHasRunFlag
		ltlog.Println(ltlog.Linfo, "Longtooth ready go ...")
		go LtRequestHandler()
	} else if int(event) == int(C.EVENT_LONGTOOTH_TIMEOUT) { // 请求或响应超时,30s
		ltlog.Println(ltlog.Ldebug, "Event tag EVENT_LONGTOOTH_TIMEOUT")

		go RequestErrorHandler(cMsgCacheKey, C.EVENT_LONGTOOTH_TIMEOUT)
	} else if int(event) == int(C.EVENT_SERVICE_NOT_EXIST) {
		ltlog.Println(ltlog.Ldebug, "Event tag EVENT_SERVICE_NOT_EXIST")

		// 目标长牙应用的服务不存在
		// 服务名不同
		go RequestErrorHandler(cMsgCacheKey, C.EVENT_SERVICE_NOT_EXIST)
	}
}

func RequestErrorHandler(cMsgCacheKey *C.char, eventCode C.int) {
	var (
		msgCacheKey string
		event       int = int(eventCode)
	)

	if cMsgCacheKey != nil {
		msgCacheKey = C.GoString(cMsgCacheKey)
		defer C.free(unsafe.Pointer(cMsgCacheKey))
	}

	if msgCacheKey == "" {
		ltlog.Println(ltlog.Ldebug, "Request attachment is empty.")
		return
	}

	com, ok := msgHashCache.get(msgCacheKey)
	if !ok {
		ltlog.Printf(ltlog.Lerror, "Cache key[%s] with msgHashCache not exists", msgCacheKey)
		return
	}
	defer msgHashCache.del(msgCacheKey)

	// 防止死锁
	user := buildUser(com.groupId, com.ltId)
	RunProc.sendUnlock(user)

	com.failFunc(&respStatus{300, "Request fail, event: " + ltEventTag[event]})
}

func LtRequestHandler() {
	for {
		// phase 1
		// 解析消息
		// 生成用户列表和用户队列
		reqWg.Add(phase1GoroutineCount)

		for i, c := 0, phase1GoroutineCount; i < c; i++ {
			go phase1Handle()
		}

		reqWg.Wait()

		// phase 2
		// 发送消息
		// 每个用户一个协程，顺序发送用户队列的消息
		for {
			if RunProc.getCounter(STAGE_SEND) > phase2GoroutineCount-1 {
				break
			}

			RunProc.incrCounter(STAGE_SEND)
			go phase2Handle()
		}

		time.Sleep(delayDuration)
	}

	requestHandleHasRunFlag <- true
}

// 第一阶段
func phase1Handle() {
	defer reqWg.Done()

	msgList, err := cache.RangeMsgQueue(goroutineParseMessageCount)
	if err != nil {
		return
	}
	ltlog.Println(ltlog.Ldebug, "Phase1 cache data:", msgList)

	var (
		cmp  Message
		user string
	)
	for _, msgJson := range msgList {
		err = json.Unmarshal([]byte(msgJson), &cmp)
		if err != nil {
			ltlog.Println(ltlog.Lerror,
				tool.JoinString("Json unmarshal fail, msg:", msgJson, ", err:", err.Error()))
			continue
		}

		// user index
		user = buildUser(cmp.GroupId, cmp.LtId)
		ret, er := cache.SetUser(user)
		if er != nil {
			ltlog.Printf(ltlog.Lerror, "Create user index fail, user[%s]", user)
		}
		if ret > 0 {
			ltlog.Printf(ltlog.Linfo, "Create user queue in phase1, user[%s]", user)
		}

		// user message set
		_, er = cache.PushGroupUserMsg(user, msgJson, cmp.SendId)
		if er != nil {
			ltlog.Printf(ltlog.Lerror, "Create user queue fail, user[%s]", user)
		}
		ltlog.Printf(ltlog.Ldebug, "Push user[%s] msg, send_id[%s], data:%s", user, cmp.SendId, msgJson)
	}

	return
}

// 第二阶段
func phase2Handle() {
	defer func() {
		if err := recover(); err != nil {
			ltlog.Println(ltlog.Lerror, "phase2 error, msg:", err)
		}
		RunProc.decrCounter(STAGE_SEND)
	}()

	for {
		users, err := cache.GetUsers(handleUserCount)
		if err != nil {
			break
		}

		for _, user := range users {
			if RunProc.check(user) != PROC_RUN {
				RunProc.run(user)
				initSendLocker(user)

				go sendUserMsg(user)
			}
		}

		time.Sleep(delayDuration)
	}
}

// 发送用户消息
func sendUserMsg(user string) {
	defer func() {
		if err := recover(); err != nil {
			ltlog.Println(ltlog.Lerror, "Send user msg fail, err:", err)

			// 异常恢复用户列表数据
			cache.SetUser(user)
		}

		RunProc.stop(user)
		RunProc.sendUnlock(user)
	}()

	msg := new(Message)
	retryNum := 1

	for {
		com := NewCommunication(parseUser(user))
		msgJson, err := cache.PopGroupUserMsg(user, combineSentMessageCount)
		if err != nil {
			ltlog.Println(ltlog.Lnotice, "Cache data empty or fail, err:", err, tool.JoinString(", delay 1s ..."))
			<-time.After(time.Second)

			// 用户消息队列为空重试 3 次, 则停止用户发送进程, 删除用户发送集合
			//
			// 判断用户队列是否为空，因为 phase1 和 sendUserMsg 是两个不同的协程
			// 当 sendUserMsg sleep 时候，phase1 可能会往 redis 添加数据,
			// 为了避免误删出新添加的数据，所以在清理环境的时候要检查队列是否为空
			if retryNum > 3 {
				err := cache.CheckUserMsg(user)
				if err != nil {
					// clear message queue
					ltlog.Println(ltlog.Linfo, "Del user message queue, user["+user+"]")
					err = cache.DelGroupUserMsg(user)
					if err != nil {
						ltlog.Println(ltlog.Lerror, "Del user message queue fail, user["+user+"]")
					}

					// clear user list
					var num int
					num, err = cache.DelUser(user)
					ltlog.Println(ltlog.Linfo, "Del user list, user["+user+"]")
					if num == 0 {
						ltlog.Println(ltlog.Lnotice, "Del user in list, "+
							"user["+ user+ "] delnum["+ strconv.Itoa(num)+ "]")
					}

					// update RunProc to stopping
					ltlog.Println(ltlog.Linfo, "Update process status, user["+user+"]")
					RunProc.stop(user)
					break
				}
				retryNum = 1
			}

			ltlog.Println(ltlog.Lnotice, "Retry times ["+strconv.Itoa(retryNum)+"/3]")
			retryNum++
			continue
		}

		// lock
		RunProc.sendLock(user)

		for _, v := range msgJson {
			err = json.Unmarshal([]byte(v), &msg)
			if err != nil {
				ltlog.Println(ltlog.Lerror, "Json unmarshal fail, err:", err)
			}

			com.pushMsg(msg)
		}
		com.assembleField()
		ltlog.Printf(ltlog.Ldebug, "Send user[%s] msg, data struct:%#v", user, com)

		go sendUserMsgHelper(com)

		time.Sleep(delayDuration)
	}
}

// 发送用户消息
func sendUserMsgHelper(d *communication) {
	var (
		cDataType    C.int   = C.LT_STREAM
		cLtt         *C.char = C.CString("0000000000")
		cLtid        *C.char = C.CString(d.ltId)
		cServiceName *C.char = C.CString(ServiceName)
	)

	req := d.sign + "," + strconv.Itoa(d.len)
	cReq := C.CString(req)
	cLen := C.size_t(len(req))
	msgCacheKey := msgHashSum(d)
	cMsgCacheKey := C.CString(msgCacheKey)
	defer func() {
		C.free(unsafe.Pointer(cLtid))
		C.free(unsafe.Pointer(cServiceName))
		C.free(unsafe.Pointer(cReq))
		//C.free(unsafe.Pointer(cLtt))

		// 不能添加，在 LtStartCallback() 函数中需要通过指针
		// 调用对应的内存值
		//C.free(unsafe.Pointer(cMsgCacheKey))

		if err := recover(); err != nil {
			ltlog.Println(ltlog.Lerror, "Send user msg by longtooth fail, err:", err)

			// 防止死锁
			user := buildUser(d.groupId, d.ltId)
			RunProc.sendUnlock(user)

			d.failFunc(&respStatus{102, "Send user msg fail"})
		}
	}()

	ret := int(C.cb_lt_request(cLtt, cLtid, cServiceName, cDataType, cReq, cLen, cMsgCacheKey))
	if ret != 0 {
		d.failFunc(&respStatus{101, "longtooth service not run"})
		return
	}
	ltlog.Printf(ltlog.Ldebug, "Request data succ, service_name[%s] lt_id[%s] sign[%s] count[%d]\n",
		ServiceName, d.ltId, d.sign, d.counter)

	// save message hash to cache
	msgHashCache.set(msgCacheKey, d)
}

//export LtRequestCallback
func LtRequestCallback(cLtt *C.char, cLtid *C.char, cServiceStr *C.char,
	cDataType C.int, cArgs *C.char, cArgsLen C.int, cMsgCacheKey *C.char,
	cHandler C.lt_attachment_handler) {
	// 接收 lt_response 响应
	goArgs := C.GoStringN(cArgs, cArgsLen)
	C.ltt_receive(cLtt, nil, -1)

	var (
		qd   *communication
		ok   bool
		resp response
	)
	msgCacheKey := C.GoString(cMsgCacheKey)

	if qd, ok = msgHashCache.get(msgCacheKey); !ok {
		ltlog.Printf(ltlog.Lerror, "Hash msgHashCache[%s] not exists", msgCacheKey)
		return
	}
	defer func() {
		msgHashCache.del(msgCacheKey)

		// 防止死锁
		user := buildUser(qd.groupId, qd.ltId)
		RunProc.sendUnlock(user)

		if err := recover(); err != nil {
			ltlog.Println(ltlog.Lerror, "Send user msg by longtooth fail, err:", err)
			qd.failFunc(&respStatus{204, "Send user msg fail"})
		}
	}()

	err := tool.DecodeJson([]byte(goArgs), &resp)
	if err != nil {
		ltlog.Println(ltlog.Lerror, "Parse json fail, jsonStr: "+goArgs)
		qd.failFunc(&respStatus{201, "Decode response json fail"})
		return
	}
	if resp.status.code != 0 {
		ltlog.Println(ltlog.Lerror, "Response fail, jsonStr: "+goArgs)
		qd.failFunc(&respStatus{202, "Response wrong, response:" + goArgs})
		return
	}

	// 发送数据体
	data := qd.message
	cData := C.CString(data)
	defer C.free(unsafe.Pointer(cData))

	sendLen := int(C.ltt_send(cLtt, cData, C.int(len(data))))
	if sendLen < 0 {
		ltlog.Printf(ltlog.Lerror, "ltt_send Send data fail, transfer data length: %s\n"+strconv.Itoa(sendLen))
		qd.failFunc(&respStatus{203, "Send data fail, transfer data length:" + strconv.Itoa(sendLen)})
		return
	}
	C.ltt_send(cLtt, nil, -1)

	ltlog.Printf(ltlog.Linfo, "Send data succ, data stuct:%#v\n", qd)
	qd.succFunc(&respStatus{0, "Send Data Succ"})
}

func msgHashSum(com *communication) string {
	ctx := md5.New()
	ctx.Write([]byte(com.sign + tool.GetCurrTimeStamp() + "\n"))
	return hex.EncodeToString(ctx.Sum(nil))
}

func buildUser(groupId, ltid string) string {
	return tool.JoinString(groupId, ":", ltid)
}

func parseUser(user string) (groupId, ltId string) {
	u := strings.Split(user, ":")

	groupId = u[0]
	ltId = u[1]
	return
}

func initSendLocker(user string) {
	RunProc.locking[user] = make(chan bool, 1)
}
