package master

import (
	"github.com/coreos/etcd/clientv3"
	"time"
	"cron/common"
	"context"
	"encoding/json"
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
	if client,err=clientv3.New(config);err!=nil{
		return
	}
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
	if putResponse,err=jobMgr.kv.Put(context.TODO(),jobKey,string(content),clientv3.WithPrefix());err!=nil{
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
	if delResponse,err=jobMgr.kv.Delete(context.TODO(),jobKey,clientv3.WithPrefix());err!=nil{
		return
	}
	if delResponse.PrevKvs!=nil{
		if err=json.Unmarshal(delResponse.PrevKvs[0].Value,&oldJobObj);err!=nil{
			err=nil
			return
		}
		oldJob=&oldJobObj
	}
	return
}