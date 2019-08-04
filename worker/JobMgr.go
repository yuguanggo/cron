package worker

import (
	"context"
	"cron/common"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

type JobMgr struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
	watcher clientv3.Watcher
}
var (
	G_jobMgr *JobMgr
)

func (jobMgr *JobMgr)watchJobs()(err error)  {
	var(
		getResponse *clientv3.GetResponse
		kvpair *mvccpb.KeyValue
		job *common.Job
		revision int64
		watchChan clientv3.WatchChan
		watchResponse clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent *common.JobEnvent
		jobName string
	)
	//1获取当前的任务列表
	if getResponse,err=jobMgr.kv.Get(context.TODO(),common.JOB_SAVE_DIR,clientv3.WithPrefix());err!=nil{
		return
	}
	for _,kvpair=range getResponse.Kvs{
		//反序列化得到job
		if job,err=common.UnpackJob(kvpair.Value);err==nil{
			jobEvent=common.BuildJobEvent(common.JOB_EVENT_SAVE,job)
			//同步给调度协成
			G_Scheduler.PushJobEvent(jobEvent)
		}
	}
	//从当前版本开始监听任务的变化
	go func() {
		fmt.Println("启动watchJobs:")
		revision = getResponse.Header.Revision+1
		watchChan = jobMgr.watcher.Watch(context.TODO(),common.JOB_SAVE_DIR,clientv3.WithRev(revision),clientv3.WithPrefix())
		for watchResponse=range watchChan{
			for _,watchEvent=range watchResponse.Events{
				switch watchEvent.Type {
				case mvccpb.PUT:
					if job,err=common.UnpackJob(watchEvent.Kv.Value);err!=nil{
						continue
					}
					//构建一个更新事件
					jobEvent=common.BuildJobEvent(common.JOB_EVENT_SAVE,job)
				case mvccpb.DELETE:
					jobName=common.ExtractJobName(string(watchEvent.Kv.Key))
					job=&common.Job{Name:jobName}
					jobEvent=common.BuildJobEvent(common.JOB_EVENT_DELETE,job)
				}
				fmt.Println("更新事件:",jobEvent.Job.Name)
				G_Scheduler.PushJobEvent(jobEvent)
			}


		}
	}()
	return
}

//监听强杀通知
func (jobMgr *JobMgr)watchKiller()  {
	var(
		watchChan clientv3.WatchChan
		watchResponse clientv3.WatchResponse
		event *clientv3.Event
		jobName string
		job *common.Job
		jobEvent *common.JobEnvent
	)
	go func() {
		watchChan=jobMgr.watcher.Watch(context.TODO(),common.JOB_KILL_DIR,clientv3.WithPrefix())
		for watchResponse=range watchChan{
			for _,event=range watchResponse.Events{
				switch event.Type {
				case mvccpb.PUT:
					//取出杀死的job name
					jobName=common.ExtractKillerName(common.JOB_KILL_DIR)
					job=&common.Job{
						Name:jobName,
					}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE,job)
					G_Scheduler.handleJobEvent(jobEvent)
				case mvccpb.DELETE://kill 标记自动过期被删除

				}
			}
		}
	}()
}

func InitJobMgr()(err error)  {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
		watcher clientv3.Watcher
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
	watcher=clientv3.NewWatcher(client)
	G_jobMgr = &JobMgr{
		client:client,
		kv:kv,
		lease:lease,
		watcher:watcher,
	}
	//启动任务监听
	G_jobMgr.watchJobs()
	//启动监听杀死任务
	G_jobMgr.watchKiller()
	return
}

func (jobMgr *JobMgr)CreateJobLock(jobName string) (jobLock *JobLock)  {
	jobLock=InitJobLock(jobName,jobMgr.kv,jobMgr.lease)
	return
}

