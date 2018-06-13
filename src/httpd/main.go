package main

// http module 没有修改
// 是 v1.0 版本旧系统
// 2018-05-17 15:56:35
import (
	"flag"
	"log"
	"lib/conf"
	"os"
	"os/signal"
	"syscall"
	"context"
	"time"
	"fmt"
	"net/http"
	"lib/ltlog"
	"router"
	"github.com/gin-gonic/gin"
)

var (
	HttpPort = flag.String("p", "8081", "http port")
	ModeFlag = flag.String("m", "dev", "run mode [dev|product]")
	WorkDir  = flag.String("d", "/tmp", "work dir")
	IniFile  = flag.String("f", "/home/ubuntu/tuyyo/etc/http.ini", "ini file")
)
var Serv *http.Server

func init() {
	flag.Parse()

	log.SetPrefix("[info] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	initLogWriter()
	initPidFile()
	conf.InitConf(IniFile)
}

func main() {
	go httpdService()

	for {
		OperSignal := make(chan os.Signal, 1)
		signal.Notify(OperSignal, syscall.SIGQUIT, syscall.SIGUSR1)
		sig := <-OperSignal

		if sig == syscall.SIGQUIT {
			stopHandler()
			break
		} else if sig == syscall.SIGUSR1 {
			reloadHandler()
		}
	}
}

func httpdService() {
	if *ModeFlag == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.SetRouter()
	Serv = &http.Server{
		Addr:         ":" + (*HttpPort),
		Handler:      r,
		ReadTimeout:  50 * time.Second,
		WriteTimeout: 50 * time.Second,
	}

	log.Println("Start Http Service")
	if err := Serv.ListenAndServe(); err != nil {
		log.Printf("listen: %s\n", err)
	} else {
		log.Printf("Http Port: %s\n", *HttpPort)
	}
}

func initLogWriter() {
	logFile := fmt.Sprintf("%s/httpd_%s_%s.log", *WorkDir, *ModeFlag, *HttpPort)
	lf, err := ltlog.NewLogFile(logFile, nil)
	if err != nil {
		log.Fatal("Unable to create log file: ", err)
	}
	log.SetOutput(lf)

	// rotate log every 15 days
	rotateLogSignal := time.Tick(15 * 24 * time.Hour)
	go func() {
		for {
			<-rotateLogSignal
			if err := lf.Rotate(); err != nil {
				log.Fatal("Unable to rotate log: ", err)
			}
		}
	}()
}

func stopHandler() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Serv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown: %+v", err)
	}
	log.Println("Close Http Service")
	log.Println("Server Stopped.")
}

func reloadHandler() {
	log.Println("Configuration Reloaded")
	conf.InitConf(IniFile)
}

func initPidFile() {
	pidFile := fmt.Sprintf("%s/httpd_%s_%s.pid", *WorkDir, *ModeFlag, *HttpPort)

	fd, err := os.Create(pidFile)
	if err != nil {
		log.Fatalln("init httpd fail, ", err)
	}
	fd.Write([]byte(fmt.Sprint(os.Getpid())))
	defer fd.Close()
}
