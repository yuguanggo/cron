package main

import (
	"runtime"
	"cron/master"
	"flag"
	"fmt"
	"time"
)

var (
	configFile string
)
//初始化线程
func initProcs()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

//初始化参数
func initArgs()  {
	flag.StringVar(&configFile,"config","./master.json","指定master.json")
	flag.Parse()
}

func main()  {
	var (
		err error
	)
	//初始化线程
	initProcs()

	//初始化命令行参数
	initArgs()

	//初始化配置
	if err=master.InitConfig(configFile);err!=nil{
		goto ERROR
	}
	//初始话etcd
	if err=master.InitJobMgr();err!=nil{
		goto ERROR
	}

	//初始化服务器
	if err=master.InitApiServer();err!=nil{
		goto ERROR
	}

	for{
		time.Sleep(1*time.Second)
	}

	ERROR:
		fmt.Println(err)

}