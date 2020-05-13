package worker

import (
	"context"
	"cron/common"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"net"
	"time"
)
type Register struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
	localIp string
}

var (
	G_register *Register
)

func getLocalIp()(ipv4 string,err error)  {
	var (
		addrs []net.Addr
		addr net.Addr
		ipNet *net.IPNet
		isIpNet bool
	)
	if addrs,err=net.InterfaceAddrs();err!=nil{
		return
	}
	fmt.Println(addrs)
	for _,addr=range addrs{
		if ipNet,isIpNet=addr.(*net.IPNet);isIpNet&&!ipNet.IP.IsLoopback(){
			if ipNet.IP.To4()!=nil{
				ipv4=ipNet.IP.String()
				return
			}
		}
	}
	err=common.ERR_NO_LOCAL_IP_FOUND
	return
}

func (register *Register)keepOnline()  {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId clientv3.LeaseID
		keapAliveRespChan <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp *clientv3.LeaseKeepAliveResponse
		cancelTex context.Context
		cancelFunc context.CancelFunc
		regKey string
		err error
	)
	for{
		//注册路径
		regKey=common.JOB_WORKER_DIR+register.localIp
		//创建租约
		if leaseGrantResp,err=register.lease.Grant(context.TODO(),10);err!=nil{
			goto RETRY
		}
		leaseId=leaseGrantResp.ID
		//续租
		if keapAliveRespChan,err=register.lease.KeepAlive(context.TODO(),leaseId);err!=nil{
			goto RETRY
		}
		cancelTex,cancelFunc=context.WithCancel(context.TODO())
		//注册到etcd
		if _,err=register.kv.Put(cancelTex,regKey,"",clientv3.WithLease(leaseId));err!=nil{
			goto RETRY
		}
		//处理租约应答
		for {
			select {
			case keepAliveResp=<-keapAliveRespChan:
				if keepAliveResp==nil{
					goto RETRY
				}
			}
		}
		RETRY:
			time.Sleep(1*time.Second)
		if cancelFunc!=nil{
			cancelFunc()
		}

	}
}



func InitRegister() (err error)  {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
		localIp string
	)
	config =clientv3.Config{
		Endpoints:G_config.EtcdEndpoints,
		DialTimeout:time.Duration(G_config.EtcdDialTimeout)*time.Millisecond,
	}
	if client,err=clientv3.New(config);err!=nil{
		return
	}
	//获取ip
	if localIp,err=getLocalIp();err!=nil{
		return
	}
	kv=clientv3.NewKV(client)
	lease=clientv3.NewLease(client)
	G_register=&Register{
		client:client,
		kv:kv,
		lease:lease,
		localIp:localIp,
	}
	//服务注册
	G_register.keepOnline()
	return
}
