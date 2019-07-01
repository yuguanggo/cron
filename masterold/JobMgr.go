package masterold

import (
	"github.com/coreos/etcd/clientv3"
	"time"
	"cron/common"
	"encoding/json"
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

//单例
var (
	G_jobMgr *JobMgr
)

func InitJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)
	//初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}
	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	//得到kv和lease的api子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	//赋值单例
	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

func (jobMgr *JobMgr) SaveJob(job *commonold.Job) (oldJob *commonold.Job, err error) {
	var (
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj commonold.Job
	)
	//etcd 保存的key
	jobKey = commonold.JOB_SAVE_DIR + job.Name

	//任务信息json
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	//将任务存到etcd
	if putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}
	//如果是更新返回旧值
	if putResp.PrevKv != nil {
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			oldJob = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

func (jobMgr *JobMgr) DeleteJob(name string) (oldJob *commonold.Job, err error) {
	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj commonold.Job
	)
	jobKey = commonold.JOB_SAVE_DIR + name
	if delResp, err = jobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	if len(delResp.PrevKvs) != 0 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			oldJob = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

func (jobMgr *JobMgr) ListJob() (jobList []*commonold.Job, err error) {
	var (
		dirKey  string
		getResp *clientv3.GetResponse
		kvPair *mvccpb.KeyValue
		job *commonold.Job
	)
	dirKey = commonold.JOB_SAVE_DIR
	if getResp, err = jobMgr.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return
	}
	jobList = make([]*commonold.Job,0)
	for _,kvPair = range getResp.Kvs{
		job = &commonold.Job{}
		if err = json.Unmarshal([]byte(kvPair.Value),job);err!=nil{
			err = nil
			continue
		}
		jobList = append(jobList,job)
	}
	return
}
