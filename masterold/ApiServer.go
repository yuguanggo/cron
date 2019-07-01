package masterold

import (
	"net/http"
	"net"
	"strconv"
	"time"
	"encoding/json"
	"cron/common"
)

type ApiServer struct {
	httpServer *http.Server
}

//单例

var (
	G_apiServer *ApiServer
)

func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		postJob string
		job     commonold.Job
		oldJob  *commonold.Job
		bytes   []byte
	)
	//解析表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//取出表单的job字段
	postJob = req.PostForm.Get("job")
	//反序列化job
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	//存入etcd
	if oldJob, err = G_jobMgr.SaveJob(&job); err != nil {
		goto ERR
	}
	//返回正常应答
	if bytes, err = commonold.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
//返回异常应答
	if bytes, err = commonold.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		jobName string
		oldJob  *commonold.Job
		bytes   []byte
	)
	//解析表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//获取要删除的Job名
	jobName = req.PostForm.Get("name")
	//删除job
	if oldJob, err = G_jobMgr.DeleteJob(jobName); err != nil {
		goto ERR
	}
	//正常返回
	if bytes, err = commonold.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
//异常返回
	if bytes, err = commonold.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		jobList []*commonold.Job
		err     error
		bytes   []byte
	)
	if jobList, err = G_jobMgr.ListJob(); err != nil {
		goto ERR
	}
	if bytes, err = commonold.BuildResponse(0, "success", jobList); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
//异常返回
	if bytes, err = commonold.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

//初始化服务
func InitApiServer() (err error) {
	var (
		mux        *http.ServeMux
		listener   net.Listener
		httpServer *http.Server
	)
	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)

	//启动tcp监听
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		return
	}
	//创建一个http服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Microsecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}
	//赋值单例
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}
	//启动了服务端
	go httpServer.Serve(listener)
	return
}
