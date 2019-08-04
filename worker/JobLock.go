package worker

import (
	"context"
	"cron/common"
	"github.com/coreos/etcd/clientv3"
)

type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string
	cancelFunc context.CancelFunc
	leaseId    clientv3.LeaseID
	isLocked   bool
}

//初始化一把锁
func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp     *clientv3.LeaseGrantResponse
		ctx                context.Context
		cancelFunc         context.CancelFunc
		leaseId            clientv3.LeaseID
		leaseKeepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
		leaseKeepAliveResp *clientv3.LeaseKeepAliveResponse
		txn                clientv3.Txn
		jobKey             string
		txnResp            *clientv3.TxnResponse
	)

	//1创建5秒的租约
	if leaseGrantResp, err = jobLock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	//ctx用于取消续租
	ctx, cancelFunc = context.WithCancel(context.TODO())
	leaseId = leaseGrantResp.ID
	//2进行续租
	if leaseKeepAliveChan, err = jobLock.lease.KeepAlive(ctx, leaseId); err != nil {
		goto FAIL
	}
	//3处理续租应答的协程
	go func() {
		for {
			select {
			case leaseKeepAliveResp = <-leaseKeepAliveChan:
				if leaseKeepAliveResp == nil {
					goto END
				}
			}
		}
	END:
	}()
	//4创建txn事务
	txn = jobLock.kv.Txn(context.TODO())
	//锁路径
	jobKey = common.JOB_LOCK_DIR + jobLock.jobName
	txn.If(clientv3.Compare(clientv3.CreateRevision(jobKey), "=", 0)).
		Then(clientv3.OpPut(jobKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(jobKey))
	//提交事务
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}
	if !txnResp.Succeeded {
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}
	//抢锁成功
	jobLock.leaseId = leaseId
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true
	return
FAIL:
	cancelFunc()//自动取消续租
	jobLock.lease.Revoke(context.TODO(),leaseId)//释放租约
	return
}

func (jobLock *JobLock)UnLock()  {
	if jobLock.isLocked{
		jobLock.cancelFunc()
		jobLock.lease.Revoke(context.TODO(),jobLock.leaseId)
	}

}
