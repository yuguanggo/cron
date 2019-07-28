package common

import (
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

func BuildJobEvent(eventType int,job *Job) *JobEnvent  {
	return &JobEnvent{
		EventType:eventType,
		Job:job,
	}
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

