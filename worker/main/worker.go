package main

import (
	"cron/worker"
	"flag"
	"fmt"
	"runtime"
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
	flag.StringVar(&configFile,"config","./worker.json","worker.json")
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
	if err=worker.InitConfig(configFile);err!=nil{
		goto ERROR
	}
	//服务注册
	if err=worker.InitRegister();err!=nil{
		goto ERROR
	}
	//初始化调度协成
	if err=worker.InitScheduler();err!=nil{
		goto ERROR
	}

	//初始话任务管理器
	if err=worker.InitJobMgr();err!=nil{
		goto ERROR
	}



	for{
		time.Sleep(1*time.Second)
	}

	ERROR:
		fmt.Println(err)

}