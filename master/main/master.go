package main

import (
	"runtime"
	"flag"
	"cron/master"
	"fmt"
	"time"
)

var (
	confFile string //配置文件路径
)

func initArgs() {
	flag.StringVar(&confFile, "config", "./master.json", "指定配置文件")
	flag.Parse()
}

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
func main() {
	var (
		err error
	)
	//初始化命令行参数
	initArgs()
	//初始化线程
	initEnv()
	if err = master.InitConfig(confFile); err != nil {
		goto ERR
	}
	//初始化任务管理器
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}
	//启动api服务器
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}
	// 正常退出
	for {
		time.Sleep(1 * time.Second)
	}
	return
ERR:
	fmt.Println(err)
}
