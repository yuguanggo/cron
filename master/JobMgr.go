package master

import (
	"context"
	"cron/common"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}
var (
	G_jobMgr *JobMgr
)
func InitJobMgr()(err error)  {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
	)
	config = clientv3.Config{
		Endpoints:G_config.EtcdEndpoints,
		DialTimeout:time.Duration(G_config.EtcdDialTimeout)*time.Millisecond,
	}
	fmt.Println("8")
	if client,err=clientv3.New(config);err!=nil{
		fmt.Println("9")
		return
	}
	fmt.Println("10")
	kv=clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	G_jobMgr = &JobMgr{
		client:client,
		kv:kv,
		lease:lease,
	}
	return
}

//保存任务
func (jobMgr *JobMgr)SaveJob(job *common.Job)(oldJob *common.Job,err error)  {
	var(
		jobKey string
		content []byte
		putResponse *clientv3.PutResponse
		oldJobObj common.Job
	)
	jobKey=common.JOB_SAVE_DIR+job.Name
	if content,err=json.Marshal(job);err!=nil{
		return
	}
	if putResponse,err=jobMgr.kv.Put(context.TODO(),jobKey,string(content),clientv3.WithPrevKV());err!=nil{

		return
	}
	if putResponse.PrevKv!=nil{
		if err=json.Unmarshal([]byte(putResponse.PrevKv.Value),&oldJobObj);err!=nil{
			err=nil
			return
		}
		oldJob=&oldJobObj
	}
	return
}

//删除任务
func (jobMgr *JobMgr)DeleteJob(name string)(oldJob *common.Job,err error){
	var(
		jobKey string
		delResponse *clientv3.DeleteResponse
		oldJobObj common.Job
	)
	jobKey = common.JOB_SAVE_DIR+name
	fmt.Println("DeleteJob1")
	if delResponse,err=jobMgr.kv.Delete(context.TODO(),jobKey,clientv3.WithPrevKV());err!=nil{
		fmt.Println("DeleteJob2")
		return
	}
	fmt.Println("DeleteJob3")
	if delResponse.PrevKvs!=nil{
		if err=json.Unmarshal(delResponse.PrevKvs[0].Value,&oldJobObj);err!=nil{
			err=nil
			return
		}
		oldJob=&oldJobObj
	}
	return
}

//任务列表
func (jobMgr *JobMgr)ListJobs()(jobList []*common.Job,err error){
	var(
		jobdir string
		getResp *clientv3.GetResponse
		kvPair *mvccpb.KeyValue
		job *common.Job
	)
	jobdir = common.JOB_SAVE_DIR
	fmt.Println("4")
	if getResp,err=jobMgr.kv.Get(context.TODO(),jobdir,clientv3.WithPrefix());err!=nil{
		fmt.Println("5")
		return
	}
	fmt.Println("6")
	jobList=make([]*common.Job,0)
	for _,kvPair=range getResp.Kvs{
		job=&common.Job{}
		if err=json.Unmarshal(kvPair.Value,job);err!=nil{
			err=nil
			continue
		}
		jobList=append(jobList,job)
	}
	return
}

//杀死任务
func (jobMgr *JobMgr)KillJob(job string)(err error){
	var(
		killKey string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId clientv3.LeaseID
	)
	killKey=common.JOB_KILL_DIR+job
	if leaseGrantResp,err=jobMgr.lease.Grant(context.TODO(),1);err!=nil{
		return
	}
	leaseId = leaseGrantResp.ID
	if _,err=jobMgr.kv.Put(context.TODO(),killKey,"",clientv3.WithLease(leaseId));err!=nil{
		return
	}
	return
}