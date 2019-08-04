package worker

import (
	"cron/common"
	"fmt"
	"time"
)

type Scheduler struct {
	jobEventChan chan *common.JobEnvent
	jobPlanTable map[string]*common.JobSchedulePlan //任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo
	jobResultChan chan *common.JobExecuteResult
}

var (
	G_Scheduler *Scheduler
)

//尝试执行任务
func (scheduler *Scheduler)TryStartJob(jobPlan *common.JobSchedulePlan)  {
	var(
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting bool
	)
	if jobExecuteInfo,jobExecuting=scheduler.jobExecutingTable[jobPlan.Job.Name];jobExecuting{
		fmt.Println("正在执行任务..",jobPlan.Job.Name)
		return
	}
	jobExecuteInfo=common.BuildJobExecuteInfo(jobPlan)

	//保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name]=jobExecuteInfo

	//执行任务
	//fmt.Println("执行任务",jobExecuteInfo.Job.Name,jobExecuteInfo.PlanTime,jobExecuteInfo.RealTime)
	G_executor.ExecuteJob(jobExecuteInfo)
}

func (scheduler *Scheduler)TrySchedule()(scheduleAfter time.Duration)  {
	var(
		jobPlan *common.JobSchedulePlan
		now time.Time
		nearTime *time.Time
	)
	if len(scheduler.jobPlanTable)==0{
		scheduleAfter =1*time.Second
		return
	}
	fmt.Println("jobplan",scheduler.jobPlanTable)
	now=time.Now()
	//遍历所有任务
	for _,jobPlan=range scheduler.jobPlanTable{
		if jobPlan.NextTime.Before(now)||jobPlan.NextTime.Equal(now){
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime=jobPlan.Expr.Next(now)
		}
		//统计最近一个要过期的任务
		if nearTime==nil||jobPlan.NextTime.Before(*nearTime){
			nearTime=&jobPlan.NextTime
		}
	}
	//下次调度间隔
	scheduleAfter=(*nearTime).Sub(now)
	return
}

func (scheduler *Scheduler)handleJobEvent(jobEvent *common.JobEnvent)  {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting bool
		jobExisted bool
		err error
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		if jobSchedulePlan,err=common.BuildJobSchedulePlan(jobEvent.Job);err!=nil{
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name]=jobSchedulePlan
	case common.JOB_EVENT_DELETE:
		if jobSchedulePlan,jobExisted=scheduler.jobPlanTable[jobEvent.Job.Name];jobExisted{
			delete(scheduler.jobPlanTable,jobEvent.Job.Name)
		}
		fmt.Println("删除任务",jobSchedulePlan.Job.Name)
	case common.JOB_EVENT_KILL:
		if jobExecuteInfo,jobExecuting=scheduler.jobExecutingTable[jobEvent.Job.Name];jobExecuting{
			jobExecuteInfo.CancelFunc()//触发command杀死shell子进程
		}
	}
}
//处理任务执行结果
func (scheduler *Scheduler)handleJobResult(jobResult *common.JobExecuteResult)  {
	delete(scheduler.jobExecutingTable,jobResult.ExecuteInfo.Job.Name)
	fmt.Println("任务执行完成",jobResult.ExecuteInfo.Job.Name,jobResult.OutPut,jobResult.Err,jobResult.StartTime,jobResult.EndTime)
}

//推送任务变化状态
func (scheduler *Scheduler)PushJobEvent(jobEvent *common.JobEnvent)  {
	scheduler.jobEventChan<-jobEvent
}
//推送任务执行结果
func (scheduler *Scheduler)PushJobResult(jobResult *common.JobExecuteResult)  {
	scheduler.jobResultChan<-jobResult
}

func (scheduler *Scheduler) schedulerLoop() {
	var(
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobEvent *common.JobEnvent
		jobResult *common.JobExecuteResult
	)
	//初始化任务状态
	scheduleAfter=scheduler.TrySchedule()

	//调度的延时定时器
	scheduleTimer=time.NewTimer(scheduleAfter)

	for{
		select {
		case jobEvent=<-scheduler.jobEventChan:
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C:
		case jobResult=<-scheduler.jobResultChan:
			scheduler.handleJobResult(jobResult)
		}
		scheduleAfter=scheduler.TrySchedule()
		scheduleTimer.Reset(scheduleAfter)
	}
}

func InitScheduler() (err error) {
	G_Scheduler=&Scheduler{
		jobEventChan:make(chan *common.JobEnvent,1000),
		jobPlanTable:make(map[string]*common.JobSchedulePlan),
		jobExecutingTable:make(map[string]*common.JobExecuteInfo,1000),
		jobResultChan:make(chan *common.JobExecuteResult,1000),
	}
	//启动调度协程
	go G_Scheduler.schedulerLoop()
	return
}
