package master

import (
	"net"
	"strconv"
	"net/http"
	"time"
)

type ApiServer struct {
	HttpServer *http.Server
}

var (
	G_ApiServer *ApiServer
)

func HandlerSave(w http.ResponseWriter,r *http.Request)  {

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
	mux.HandleFunc("/save",HandlerSave)
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

