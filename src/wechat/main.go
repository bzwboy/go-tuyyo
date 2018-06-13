package main

import (
	"flag"
	"log"
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
	"runtime"

	"lib/cache"
	"lib/ltlog"
	lt "lib/longtooth"
	"lib/conf"
	"path/filepath"
)

var (
	ModeFlag = flag.String("m", "dev", "run mode [dev|product]")
	ServName = flag.String("n", "longtooth", "service name")
	WorkDir  = flag.String("d", "/tmp", "work dir")
	IniFile  = flag.String("f", "/home/ubuntu/tuyyo/etc/conf.ini", "ini file")
)

// ** 谨慎在非 main 包中使用 init 初始化函数 **
// 在 main.init() 启动之前，import 原因，其他包的 init 已经执行
// 如果用到 flag 中的参数，非 main 包中的 init 将不能获得配置的参数
// @date 2018-04-28 17:57
func init() {
	flag.Parse()

	log.SetFlags(log.LstdFlags)

	// dev 模式日志输出到 /tmp/debug
	if "dev" == *ModeFlag {
		*IniFile = filepath.Dir(*IniFile) + "/conf_" + *ModeFlag + ".ini"
	} else {
		initLogWriter()
	}
	initPidFile()

	// init
	conf.InitConf(IniFile)
	cache.InitPool()
}

func main() {
	go func() {
		for {
			initLtcService()

			// clear env
			stopHandler()
			runtime.GC()

			time.Sleep(500 * time.Millisecond)
		}
	}()

	for {
		OperSignal := make(chan os.Signal, 1)
		signal.Notify(OperSignal, syscall.SIGQUIT, os.Interrupt, syscall.SIGUSR1, syscall.SIGUSR2)
		sig := <-OperSignal

		if sig == syscall.SIGQUIT {
			stopHandler()
			break
		} else if sig == syscall.SIGUSR1 {
			reloadHandler()
		} else if sig == syscall.SIGUSR2 {
			statusHandler()
		}
	}
}

func initLtcService() {
	defer func() {
		if err := recover(); err != nil {
			ltlog.Println(ltlog.Lerror, err)
		}
	}()

	lt.ServiceName = *ServName
	lt.ModeFlag = *ModeFlag
	lt.Start()
}

func initLogWriter() {
	logFile := fmt.Sprintf("%s/wechat_%s.log", *WorkDir, *ModeFlag)
	lf, err := ltlog.NewLogFile(logFile, nil)
	if err != nil {
		ltlog.Fatalln("Unable to create log file: ", err)
	}
	log.SetOutput(lf)

	// rotate log every 15 days
	rotateLogSignal := time.Tick(15 * 24 * time.Hour)
	go func() {
		for {
			<-rotateLogSignal
			if err := lf.Rotate(); err != nil {
				ltlog.Fatalln("Unable to rotate log: ", err)
			}
		}
	}()
}

// 中断程序
func stopHandler() {
	cache.Close()
	lt.Close()

	ltlog.Println(ltlog.Lnotice, "Close Longtooth Service")
	ltlog.Println(ltlog.Lnotice, "Server Stopped.")
}

// 重新加载配置
func reloadHandler() {
	ltlog.Println(ltlog.Linfo, "Configuration Reloaded")
	conf.InitConf(IniFile)

	if err := cache.Close(); err != nil {
		ltlog.Println(ltlog.Lerror, "reload error, err:", err)
	}
	cache.InitPool()
}

// 长牙模块统计信息
func statusHandler()  {
	lt.RunProc.Status()
}

func initPidFile() {
	pidFile := fmt.Sprintf("%s/wechat_%s.pid", *WorkDir, *ModeFlag)

	fd, err := os.Create(pidFile)
	if err != nil {
		ltlog.Fatalln("init LtCenter fail, ", err)
	}
	fd.Write([]byte(fmt.Sprint(os.Getpid())))
	defer fd.Close()
}
