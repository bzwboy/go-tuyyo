package longtooth

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
	"fmt"
	"strconv"
	"net/url"
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"lib/conf"
	"strings"
	"lib/ltlog"
)

// json 公开访问
type Message struct {
	GroupId string `json:"group_id"`
	SendId  string `json:"send_id"`
	LtId    string `json:"lt_id"`
	Data    string `json:"data"`
}

// longtooth communication struct
// not goroutine safety
type communication struct {
	groupId string
	ltId    string

	// msgSet[sendId] = 1
	crops   map[string]int
	message string

	// auto created
	// send data hash
	sign string
	// auto created
	// send data length
	len int
	// user message number per sending
	counter int
}

func NewCommunication(groupId, ltId string) (com *communication) {
	com = new(communication)
	com.groupId = groupId
	com.ltId = ltId
	com.crops = make(map[string]int, combineSentMessageCount)
	return
}

func (com *communication) pushMsg(msg *Message) {
	if _, ok := com.crops[msg.SendId]; !ok {
		com.crops[msg.SendId] = 1
		com.message += msg.Data + string(msgSeperator)
		com.counter++
	}
}

func (com *communication) pushMsgByString(str string) {
	msg := new(Message)
	err := json.Unmarshal([]byte(str), msg)
	if err != nil {
		ltlog.Println(ltlog.Lerror, "Json decode fail, err:", err)
	}

	if _, ok := com.crops[msg.SendId]; !ok {
		com.crops[msg.SendId] = 1
		com.message += msg.Data + string(msgSeperator)
		com.counter++
	}
}

func (com *communication) assembleField() {
	// sign
	ctx := md5.New()
	ctx.Write([]byte(com.message + "\n"))
	com.sign = hex.EncodeToString(ctx.Sum(nil))

	// len
	com.len = len(com.message)

	// message
	com.message = strings.Trim(com.message, string(msgSeperator))
}

func (com *communication) succFunc(status *respStatus) {
	ltlog.Printf(ltlog.Ldebug, "Send successful data to center, msg:%s\n", com.message)

	var wg sync.WaitGroup
	wg.Add(com.counter)
	for sid, _ := range com.crops {
		go func(sid string) {
			defer wg.Done()

			ret, err := RequestHttpCenterWithGet(sid, status)
			if err != nil {
				com.notifyHttpCenterFail(err)
			} else {
				ltlog.Printf(ltlog.Ldebug, "Inform send_id[%s] to center, reply:%s\n", sid, ret)
			}
		}(sid)
	}
	wg.Wait()
}

func (com *communication) failFunc(status *respStatus) {
	ltlog.Printf(ltlog.Ldebug, "Send failure data to center, msg:%s\n", com.message)

	var wg sync.WaitGroup
	wg.Add(com.counter)
	for sid, _ := range com.crops {
		go func(sid string) {
			defer wg.Done()

			ret, err := RequestHttpCenterWithGet(sid, status)
			if err != nil {
				com.notifyHttpCenterFail(err)
			} else {
				ltlog.Printf(ltlog.Ldebug, "Inform send_id[%s] to center, reply:%s\n", sid, ret)
			}
		}(sid)
	}
	wg.Wait()
}

// 扩展功能
// 以后通知 httpCenter 错误
// 需要执行的逻辑
func (com *communication) notifyHttpCenterFail(err error) {
	ltlog.Printf(ltlog.Lerror, "Notify httpCenter fail, err: %+v", err)
}

func RequestHttpCenterWithGet(sendId string, st *respStatus) (ret string, err error) {
	httpDomain, _ := conf.String(conf.ConfType("controller").Field("responseDomain"))
	httpPath, _ := conf.String(conf.ConfType("controller").Field("responsePath"))

	httpUrl := httpDomain + httpPath +
		"&code=" + strconv.Itoa(st.code) + "&send_id=" + sendId +
		"&code_msg=" + url.PathEscape(st.msg)
	//ltlog.Printf(ltlog.Ldebug, "Call HttpCenter Url: %s\n", httpUrl)

	// 错误重试 3 次
	resp, err := http.Get(httpUrl)
	if err != nil {
		for i, c := 0, 3; i < c; i++ {
			resp, err = http.Get(httpUrl)
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if err != nil {
			return "", err
		}
	}
	defer resp.Body.Close()

	retByte, _ := ioutil.ReadAll(resp.Body)
	return string(retByte), nil
}

// lt response json struct
type respStatus struct {
	code int
	msg  string
}

type response struct {
	status respStatus
}

// 统计正在运行的线程
// 正在发送消息的协程
type runningProcess struct {
	runner  map[string]int       // 正在发送的用户数据
	counter map[string]int       // 计数器，统计 phase2 启动的协程数
	locking map[string]chan bool // 锁，保证顺序发送用户信息
	mu      sync.Mutex
}

func NewRunningProcess() *runningProcess {
	return &runningProcess{
		runner:  make(map[string]int, 100),
		counter: make(map[string]int, 2),
		locking: make(map[string]chan bool, 100),
	}
}

// 启动
func (r *runningProcess) run(user string) {
	defer r.mu.Unlock()
	r.mu.Lock()

	r.runner[user] = PROC_RUN
}

// 暂停
func (r *runningProcess) pending(user string) {
	defer r.mu.Unlock()
	r.mu.Lock()

	r.runner[user] = PROC_STOP
}

// 停止
func (r *runningProcess) stop(user string) {
	defer r.mu.Unlock()
	r.mu.Lock()

	delete(r.runner, user)
}

func (r *runningProcess) sendLock(user string) {
	r.locking[user] <- true
}

// 不能加入 mu lock
// 有可能产生阻塞
// @see RequestErrorHandler()
func (r *runningProcess) sendUnlock(user string) {
	if len(r.locking[user]) > 0 {
		<-r.locking[user]
	}
}

func (r *runningProcess) Status() {
	fmt.Println("** Statistic running process **")
	if len(r.runner) != 0 {
		for ltid, st := range r.runner {
			fmt.Printf("ltid[%s]\tstat[%s]\n", ltid, procTag[st])
		}
	} else {
		fmt.Printf("ltid[%s]\tstat[%s]\n", "None", "None")
	}
	fmt.Println("** End **")
}

// -1 not exist
// 0 stopped
// 1 running
func (r *runningProcess) check(user string) int {
	defer r.mu.Unlock()
	r.mu.Lock()

	if st, ok := r.runner[user]; ok {
		return st
	}
	return PROC_NIL
}

func (r *runningProcess) incrCounter(stage string) {
	defer r.mu.Unlock()
	r.mu.Lock()

	r.counter[stage]++
}

func (r *runningProcess) decrCounter(stage string) {
	defer r.mu.Unlock()
	r.mu.Lock()

	r.counter[stage]--
}

func (r *runningProcess) getCounter(stage string) (count int) {
	defer r.mu.Unlock()
	r.mu.Lock()

	return r.counter[stage]
}

// 缓存发送的消息
// 事件处理使用
type ltHashCache struct {
	cache map[string]*communication
	lock  sync.Mutex
}

func (m *ltHashCache) set(k string, v *communication) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.cache[k] = v
}

func (m *ltHashCache) get(k string) (v *communication, ok bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	v, ok = m.cache[k]
	return
}

func (m *ltHashCache) del(k string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.cache, k)
}
