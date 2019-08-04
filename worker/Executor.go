package worker

import (
	"cron/common"
	"fmt"
	"os/exec"
	"time"
)

type Executor struct {

}

var (
	G_executor *Executor
)

func (executor *Executor)ExecuteJob(info *common.JobExecuteInfo)  {
	go func() {
		var (
			result *common.JobExecuteResult
			jobLock *JobLock
			cmd *exec.Cmd
			outPut []byte
			err error
		)
		result=&common.JobExecuteResult{
			ExecuteInfo:info,
			OutPut:make([]byte,0),
		}
		//创建一个joblock
		jobLock=G_jobMgr.CreateJobLock(info.Job.Name)
		err=jobLock.TryLock()
		defer jobLock.UnLock()

		//任务开始时间
		result.StartTime=time.Now()

		if err!=nil{
			result.Err=err
			result.EndTime=time.Now()
		}else {
			//抢锁成功
			fmt.Println("抢锁成功",jobLock.jobName)
			result.StartTime=time.Now()

			//执行shell命令
			cmd=exec.CommandContext(info.CancelCtx,"/bin/bash","-c",info.Job.Command)

			//执行并捕获输出
			outPut,err=cmd.CombinedOutput()

			//记录任务结束时间
			result.OutPut=outPut
			result.EndTime=time.Now()
			result.Err=err
		}
		//执行完成后，把jobResult推给schedule
		G_Scheduler.PushJobResult(result)

	}()
}

func InitExecutor()(err error)  {
	G_executor=&Executor{

	}
	return
}