package models

import "time"

type RunStatus string

const (
	// 表示正在运行
	StatusRunning RunStatus = "running"
	//	表示未运行或程序主动退出，正常停止，快速重试流程会立即重启
	StatusExited RunStatus = "exited"
	// 表示出错停止，快速重试已经无法自动恢复，5分钟检测流程会尝试重启
	StatusError RunStatus = "error"
	// 表示被用户手动停止，5分钟检测流程不会尝试重启，用户通过启动命令可以手动启动
	StatusStopped RunStatus = "stopped"
)

type ProcessDetail struct {
	Title           string    `json:"title"`           //显示用的名字
	ProcessName     string    `json:"processName"`     //进程名，用于查找进程
	Command         string    `json:"command"`         //进程启动命令
	Args            []string  `json:"args"`            //进程参数
	WorkDir         string    `json:"workDir"`         //工作目录
	MaxRestartCount int       `json:"maxRestartCount"` //最大重启次数
	Pid             int       `json:"pid"`             //进程PID
	Status          RunStatus `json:"status"`          //状态
	RestartCount    int       `json:"restartCount"`    //重启次数
	StartTime       time.Time `json:"startTime"`       //启动时间
	LastExitTime    time.Time `json:"lastExitTime"`    //最后一次退出的时间
	LastExitReason  string    `json:"lastExitReason"`  //最后一次退出的原因
}
