package master

import (
	"net"
	"strconv"
	"net/http"
	"time"
	"encoding/json"
	"cron/common"
)

type ApiServer struct {
	HttpServer *http.Server
}

var (
	G_ApiServer *ApiServer
)

func HandlerJobSave(w http.ResponseWriter,r *http.Request)  {
	var (
		postJob string
		err error
		job common.Job
		oldJob *common.Job
		bytes []byte
	)
	//获取参数
	if err = r.ParseForm();err!=nil{
		goto ERR
	}
	postJob = r.PostForm.Get("job")

	if err=json.Unmarshal([]byte(postJob),&job);err!=nil{
		goto ERR
	}
	if oldJob,err=G_jobMgr.SaveJob(&job);err!=nil{
		goto ERR
	}
	//正常应答
	if bytes,err=common.BuildResponse(0,"",oldJob);err==nil{
		w.Write(bytes)
	}
	return
ERR:
	if bytes,err=common.BuildResponse(-1,err.Error(),nil);err==nil{
		w.Write(bytes)
	}

}
//删除任务
func HandleJobDelete(w http.ResponseWriter,r *http.Request)  {
	var (
		err error
		name string
		oldJob *common.Job
		bytes []byte
	)
	if err=r.ParseForm();err!=nil{
		goto ERR
	}
	name=r.PostForm.Get("name")
	if oldJob,err=G_jobMgr.DeleteJob(name);err!=nil{
		goto ERR
	}
	if bytes,err=common.BuildResponse(0,"success",oldJob);err==nil{
		w.Write(bytes)
	}
	return
ERR:
	if bytes,err =common.BuildResponse(-1,err.Error(),nil);err==nil{
		w.Write(bytes)
	}
}


//任务列表
func HandleJobList(w http.ResponseWriter,r *http.Request)  {
	var (
		err error
		jobList []*common.Job
		bytes []byte
	)
	if jobList,err=G_jobMgr.ListJobs();err!=nil{
		goto ERR
	}
	if bytes,err=common.BuildResponse(0,"success",jobList);err==nil{
		w.Write(bytes)
	}
	return
ERR:
	if bytes,err =common.BuildResponse(-1,err.Error(),nil);err==nil{
		w.Write(bytes)
	}
}

func HandlerJobKill(w http.ResponseWriter,r *http.Request)  {
	var (
		postJob string
		err error
		bytes []byte
	)
	//获取参数
	if err = r.ParseForm();err!=nil{
		goto ERR
	}
	postJob = r.PostForm.Get("name")

	if err=G_jobMgr.KillJob(postJob);err!=nil{
		goto ERR
	}
	//正常应答
	if bytes,err=common.BuildResponse(0,"",nil);err==nil{
		w.Write(bytes)
	}
	return
ERR:
	if bytes,err=common.BuildResponse(-1,err.Error(),nil);err==nil{
		w.Write(bytes)
	}

}

func InitApiServer() (err error) {
	var (
		listener net.Listener
		mux *http.ServeMux
		server *http.Server

	)
	if listener,err=net.Listen("tcp",":"+strconv.Itoa(G_config.ApiPort));err!=nil{
		return
	}
	//路由
	mux =http.NewServeMux()
	mux.HandleFunc("/job/save",HandlerJobSave)
	mux.HandleFunc("/job/delete",HandleJobDelete)
	mux.HandleFunc("/job/kill",HandlerJobKill)
	mux.HandleFunc("/job/list",HandleJobList)
	//创建了一个http服务
	server = &http.Server{
		ReadTimeout:time.Duration(G_config.ApiReadTimeOut)*time.Millisecond,
		WriteTimeout:time.Duration(G_config.ApiWriteTimeOut)*time.Millisecond,
		Handler:mux,
	}
	G_ApiServer = &ApiServer{
		HttpServer:server,
	}
	go server.Serve(listener)
	return
}

