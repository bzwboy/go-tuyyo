package cache

import (
	"github.com/gomodule/redigo/redis"
	"lib/conf"
	"time"
	"github.com/pkg/errors"
	"lib/tool"
	"lib/ltlog"
)

var (
	pool      *redis.Pool
	msgExpire int

	errEmpty = errors.New("data is empty")
	errFail  = errors.New("operate cache fail")
)

// 从 pool 获取 redis 连接
func InitPool() (err error) {
	host, _ := conf.String(conf.ConfType("redis").Field("host"))
	port, _ := conf.String(conf.ConfType("redis").Field("port"))
	protocol, _ := conf.String(conf.ConfType("redis").Field("protocol"))
	remote := tool.JoinString(host, ":", port)
	// unit
	msgExpire, _ = conf.Int(conf.ConfType("main").Field("userMsgExpire"))
	msgExpire = msgExpire * 3600

	poolMaxIdle, _ := conf.Int(conf.ConfType("redis").Field("poolMaxIdle"))
	poolMaxActive, _ := conf.Int(conf.ConfType("redis").Field("poolMaxActive"))
	poolIdleTimeout, _ := conf.Int(conf.ConfType("redis").Field("poolIdleTimeout"))
	poolMaxConnLifetime, _ := conf.Int(conf.ConfType("redis").Field("poolMaxConnLifetime"))
	poolWait, _ := conf.Bool(conf.ConfType("redis").Field("poolWait"))

	pool = redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial(protocol, remote)
	}, poolMaxIdle)

	pool.Wait = poolWait
	pool.MaxActive = poolMaxActive
	pool.MaxConnLifetime = time.Duration(poolMaxConnLifetime) * time.Second
	pool.IdleTimeout = time.Duration(poolIdleTimeout) * time.Second
	pool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		if time.Since(t) < time.Minute {
			return nil
		}

		if _, err := c.Do("PING"); err != nil {
			ltlog.Println(ltlog.Lfatal, "Redis error, info:", err)
		}
		return err
	}

	return nil
}

// 关闭 redis pool
func Close() error {
	return pool.Close()
}

// 获取队列信息
var lpopScript = redis.NewScript(1, `
		local r = redis.call('LRANGE', KEYS[1], ARGV[1], ARGV[2])
		if r ~= nil then
			redis.call('LTRIM', KEYS[1], ARGV[2] + 1, -1)
		end
		return r
	`)

func RangeMsgQueue(length int) (msg []string, err error) {
	conn := pool.Get()
	defer conn.Close()

	ck, err := keyMsgQueue()
	msg, err = redis.Strings(lpopScript.Do(conn, ck, 0, length-1))
	if err != nil {
		return nil, err
	}
	if len(msg) == 0 {
		return nil, errEmpty
	}

	return
}

// 设置待发送的用户
func SetUser(user string) (num int, err error) {
	conn := pool.Get()
	defer conn.Close()

	ck, _ := keyUserList()
	return redis.Int(conn.Do("SADD", ck, user))
}

// 获取用户
func GetUsers(count int) (users []string, err error) {
	conn := pool.Get()
	defer conn.Close()

	ck, _ := keyUserList()
	users, err = redis.Strings(conn.Do("SRANDMEMBER", ck, count))
	if len(users) == 0 {
		return nil, errEmpty
	}

	return
}

// 删除用户
func DelUser(user string) (num int, err error) {
	conn := pool.Get()
	defer conn.Close()

	ck, _ := keyUserList()
	return redis.Int(conn.Do("SREM", ck, user))
}

// 设置组用户信息
func PushGroupUserMsg(user, val string, sendId string) (int, error) {
	conn := pool.Get()
	defer conn.Close()

	ck, _ := keyUserMsg(user)
	num, err := redis.Int(conn.Do("ZADD", ck, sendId, val))
	if err != nil {
		return 0, err
	}

	reply, err := redis.Int(conn.Do("EXPIRE", ck, msgExpire))
	if err != nil {
		return 0, err
	}
	if reply != 1 {
		return 0, errFail
	}

	return num, nil
}

func CheckUserMsg(user string) (err error) {
	conn := pool.Get()
	defer conn.Close()

	ck, _ := keyUserMsg(user)
	repl, err := redis.Strings(conn.Do("ZRANGE", ck, 0, -1))
	if err != nil {
		return err
	}

	if len(repl) == 0 {
		return errEmpty
	}

	return nil
}

// 获取组用户信息
var zpopScript = redis.NewScript(1, `
		local r = redis.call('ZRANGE', KEYS[1], 0, ARGV[1])
		if r ~= nil then
			redis.call('ZREMRANGEBYRANK', KEYS[1], 0, ARGV[1])
		end
		return r
	`)

func PopGroupUserMsg(user string, popNum int) (msg []string, err error) {
	conn := pool.Get()
	defer conn.Close()

	ck, _ := keyUserMsg(user)
	msg, err = redis.Strings(zpopScript.Do(conn, ck, popNum-1))
	if err != nil {
		return nil, err
	}
	if len(msg) == 0 {
		return nil, errEmpty
	}

	return msg, nil
}

// 删除无用的用户信息集合
func DelGroupUserMsg(user string) (err error) {
	conn := pool.Get()
	defer conn.Close()

	ck, _ := keyUserMsg(user)
	_, err = redis.Int(conn.Do("DEL", ck))
	return
}
