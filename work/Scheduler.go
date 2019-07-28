package master

import (
	"cron/common"
	"time"
)

type Scheduler struct {
	jobEventChan chan *common.JobEnvent
	jobPlanTable map[string]*common.JobSchedulePlan
}

var (
	G_Scheduler *Scheduler
)

func (scheduler *Scheduler)TrySchedule()(scheduleAfter time.Duration)  {
	var(
		jobPlan *common.JobSchedulePlan
		now time.Time
		nearTime *time.Time
	)
	if scheduler.jobPlanTable==nil{
		scheduleAfter =1*time.Second
		return
	}
	now=time.Now()
	//遍历所有任务
	for _,jobPlan=range scheduler.jobPlanTable{
		if jobPlan.NextTime.Before(now)||jobPlan.NextTime.Equal(now){
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime=jobPlan.Expr.Next(now)
		}
		//统计最近一个要过期的任务
		if nearTime==nil||jobPlan.NextTime.Before(*nearTime){
			*nearTime=jobPlan.NextTime
		}
	}
	//下次调度间隔
	scheduleAfter=(*nearTime).Sub(now)
	return
}

func (scheduler *Scheduler) schedulerLoop() {
	var(
		scheduleAfter time.Duration
	)
	//初始化任务状态
	scheduleAfter=scheduler.TrySchedule()
}

func InitScheduler() (err error) {
	G_Scheduler=&Scheduler{
		jobEventChan:make(chan *common.JobEnvent,1000),
		jobPlanTable:make(map[string]*common.JobSchedulePlan),
	}
	//启动调度协程
	go G_Scheduler.schedulerLoop()
	return
}
