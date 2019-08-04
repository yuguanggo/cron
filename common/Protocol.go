package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

type Job struct {
	Name string `json:"name"`
	Command string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

//http 接口应答
type Response struct {
	Errno int `json:"errno"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

//变化事件
type JobEnvent struct {
	EventType int
	Job *Job
}

//任务调度计划
type JobSchedulePlan struct {
	Job *Job
	Expr *cronexpr.Expression
	NextTime time.Time
}

//任务执行状态
type JobExecuteInfo struct {
	Job *Job
	PlanTime time.Time//理论调度时间
	RealTime time.Time//实际调度时间
	CancelCtx context.Context //任务command的context
	CancelFunc context.CancelFunc //用于取消command执行的cancel函数
}

type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo //执行状态
	OutPut []byte //脚本输出
	Err error
	StartTime time.Time
	EndTime time.Time
}


func BuildJobSchedulePlan(job *Job)(jobSchedulePlan *JobSchedulePlan,err error)  {
	var (
		expr *cronexpr.Expression
	)
	if expr,err=cronexpr.Parse(job.CronExpr);err!=nil{
		return
	}
	jobSchedulePlan=&JobSchedulePlan{
		Job:job,
		Expr:expr,
		NextTime:expr.Next(time.Now()),
	}
	return
}

func BuildJobEvent(eventType int,job *Job) *JobEnvent  {
	return &JobEnvent{
		EventType:eventType,
		Job:job,
	}
}

//构建执行状态
func BuildJobExecuteInfo(jobPlan *JobSchedulePlan)(jobExecuteInfo *JobExecuteInfo)  {
	jobExecuteInfo=&JobExecuteInfo{
		Job:jobPlan.Job,
		PlanTime:jobPlan.NextTime,
		RealTime:time.Now(),
	}
	jobExecuteInfo.CancelCtx,jobExecuteInfo.CancelFunc=context.WithCancel(context.TODO())
	return
}

func BuildResponse(errno int,msg string,data interface{})(resp []byte,err error)  {
	var (
		response Response
	)
	response.Errno=errno
	response.Msg=msg
	response.Data=data

	resp,err=json.Marshal(response)
	return
}

func UnpackJob(value []byte)(ret *Job,err error)  {
	var (
		job *Job
	)
	job=&Job{}
	if err=json.Unmarshal(value,job);err!=nil{
		return
	}
	ret=job
	return
}
//提取job的名字
func ExtractJobName(jobKey string) string  {
	return strings.TrimPrefix(jobKey,JOB_SAVE_DIR)
}

//提取job的名字
func ExtractKillerName(jobKey string) string  {
	return strings.TrimPrefix(jobKey,JOB_KILL_DIR)
}

