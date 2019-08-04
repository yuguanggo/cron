package common

const (
	//任务保存目录
	JOB_SAVE_DIR="/cron/jobs/"

	//杀死任务目录
	JOB_KILL_DIR="/cron/kill/"

	//锁目录
	JOB_LOCK_DIR="/cron/lock/"

	//保存任务事件
	JOB_EVENT_SAVE=0
	//删除任务事件
	JOB_EVENT_DELETE=1
	// 强杀任务事件
	JOB_EVENT_KILL = 3
)