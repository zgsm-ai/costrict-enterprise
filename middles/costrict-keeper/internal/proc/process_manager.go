package proc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"costrict-keeper/internal/logger"
	"costrict-keeper/internal/models"
	"costrict-keeper/internal/utils"
)

type processWatcher struct {
	maxRestartCount int                    //最大重启次数(监测程序通过重启解决临时故障)
	onChanged       func(*ProcessInstance) //监测到进程重启/停止的回调函数
}

/**
 * ProcessInstance 进程实例信息
 * @property {string} title - 进程标题，用于显示
 * @property {string} procName - 进程列表显示的进程名，processName+pid可以确定一个进程身份，放误杀
 * @property {string} command - 执行命令
 * @property {[]string} args - 命令参数
 * @property {string} workDir - 工作目录
 * @property {int} pid - 进程ID
 * @property {string} status - 进程状态: running/exited/stopped/error
 * @property {int} restartCount - 重启次数
 * @property {time.Time} startTime - 启动时间
 * @property {time.Time} lastExitTime - 最后退出时间
 * @property {string} lastExitReason - 最后退出原因
 * @property {processWatcher} watcher - 监控协程设置
 */
type ProcessInstance struct {
	Title          string           //显示用的名字
	ProcessName    string           //进程名，用于查找进程
	Command        string           //进程启动命令
	Args           []string         //进程参数
	WorkDir        string           //工作目录
	Status         models.RunStatus //状态
	RestartCount   int              //重启次数
	StartTime      time.Time        //启动时间
	LastExitTime   time.Time        //最后一次退出的时间
	LastExitReason string           //最后一次退出的原因
	watcher        processWatcher   //监测协程的设置
	process        *os.Process      //统一的进程对象，用于Wait()
	mutex          sync.Mutex       //保护实例数据一致性的读写锁
}

/**
 * NewProcessInstance 创建新的进程实例
 * @param {string} title - 进程标题，可以唯一确定一个进程，即使它重启过
 * @param {string} procName - 进程名
 * @param {string} command - 执行命令
 * @param {[]string} args - 命令参数
 * @returns {ProcessInstance} 返回创建的进程实例
 * @description
 * - 创建并初始化一个新的进程实例
 * - 设置默认的进程状态和属性
 */
func NewProcessInstance(title, procName, command string, args []string) *ProcessInstance {
	return &ProcessInstance{
		Title:        title,
		ProcessName:  procName,
		Command:      command,
		Args:         args,
		WorkDir:      "",
		RestartCount: 0,
		Status:       models.StatusExited,
	}
}

func (pi *ProcessInstance) SetWatcher(maxRestart int, onChanged func(*ProcessInstance)) {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	pi.watcher.onChanged = onChanged
	pi.watcher.maxRestartCount = maxRestart
}

func (pi *ProcessInstance) Pid() int {
	if pi.process == nil {
		return 0
	}
	return pi.process.Pid
}

func (pi *ProcessInstance) GetDetail() models.ProcessDetail {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	return models.ProcessDetail{
		Title:           pi.Title,
		ProcessName:     pi.ProcessName,
		Command:         pi.Command,
		Args:            pi.Args,
		WorkDir:         pi.WorkDir,
		MaxRestartCount: pi.watcher.maxRestartCount,
		Status:          pi.Status,
		Pid:             pi.Pid(),
		RestartCount:    pi.RestartCount,
		StartTime:       pi.StartTime,
		LastExitTime:    pi.LastExitTime,
		LastExitReason:  pi.LastExitReason,
	}
}

/**
 * StartProcess 启动进程
 * @param {ProcessInstance} pi - 进程实例
 * @returns {error} 返回错误信息
 * @description
 * - 启动指定进程
 * - 自动将进程添加到管理器中
 * - 使用协程监控进程状态
 * - 如果进程配置了自动重启，会在进程退出时自动重启
 * - 更新进程状态
 */
func (pi *ProcessInstance) StartProcess(ctx context.Context) error {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()
	return pi.startProcess(ctx)
}

func (pi *ProcessInstance) startProcess(ctx context.Context) error {
	if pi.Status == models.StatusRunning {
		return nil
	}
	fullCommand := pi.Command
	for _, arg := range pi.Args {
		fullCommand += " " + arg
	}
	logger.Infof("Executing command: %s", fullCommand)

	// 创建命令
	cmd := exec.CommandContext(ctx, pi.Command, pi.Args...)

	// 设置工作目录
	if pi.WorkDir != "" {
		cmd.Dir = pi.WorkDir
	}

	if pi.watcher.onChanged == nil {
		// 设置进程属性，使子进程在父进程退出后继续运行
		utils.SetNewPG(cmd)
	}

	if err := cmd.Start(); err != nil {
		pi.Status = models.StatusError
		pi.LastExitReason = fmt.Sprintf("start failed: %v", err)
		logger.Errorf("Failed to start process '%s', error: %v", pi.Title, err)
		return err
	}

	pi.process = cmd.Process // 保存进程对象，用于统一Wait()
	pi.Status = models.StatusRunning
	pi.StartTime = time.Now()

	logger.Infof("Process '%s' started (PID: %d)", pi.Title, pi.Pid())

	if pi.watcher.onChanged != nil { // costrict.exe作为服务器运行时，启动协程监控子进程
		go pi.watchProcess()
	}
	return nil
}

/**
 * StopProcess 停止进程
 * @param {ProcessInstance} pi - 进程实例
 * @returns {error} 返回错误信息
 * @description
 * - 停止指定进程
 * - 取消进程上下文，终止进程
 * - 自动从管理器中移除进程
 * - 更新进程状态
 */
func (pi *ProcessInstance) StopProcess() error {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	if pi.Status != models.StatusRunning {
		return nil
	}
	pi.Status = models.StatusStopped
	pi.LastExitTime = time.Now()
	pi.LastExitReason = "stopped by user"

	pid := pi.Pid()
	if pi.process != nil {
		if err := pi.process.Kill(); err != nil {
			logger.Errorf("Failed to kill process '%s' (PID: %d, NAME: %s)",
				pi.Title, pid, pi.ProcessName)
			return err
		}
		pi.process.Wait()
		pi.process = nil
	}

	logger.Infof("Process '%s' (PID: %d, NAME: %s) stopped",
		pi.Title, pid, pi.ProcessName)
	return nil
}

func (pi *ProcessInstance) CheckProcess() models.HealthyStatus {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	if pi.Status != models.StatusRunning {
		return models.Unavailable
	}
	if pi.process == nil {
		return models.Unavailable
	}
	running, err := utils.IsProcessRunning(pi.Pid())
	if err != nil || !running {
		logger.Warnf("Process '%s' (PID: %d, NAME: %s) isn't running", pi.Title, pi.Pid(), pi.ProcessName)
		pi.Status = models.StatusError
		pi.process = nil
		return models.Unavailable
	}
	return models.Healthy
}

func getReason(status models.RunStatus) string {
	switch status {
	case models.StatusError:
		return "error"
	case models.StatusStopped:
		return "user"
	default:
		return "unknown"
	}
}

/**
 * watchProcess 监控进程状态的协程
 * @param {ProcessInstance} pi - 进程实例
 * @description
 * - 使用协程监控进程状态
 * - 统一使用process.Wait()等待进程退出
 * - 如果进程配置了自动重启，在进程退出时自动重启
 * - 更新进程状态并记录退出原因
 */
func (pi *ProcessInstance) watchProcess() {
	_, err := pi.process.Wait()

	pi.mutex.Lock()
	defer pi.mutex.Unlock()

	if pi.watcher.onChanged == nil { //只有onChanged!=nil才会进入watchProcess，但存在中途修改的可能性
		return
	}

	if pi.Status == models.StatusStopped || pi.Status == models.StatusError {
		logger.Infof("Process '%s' stopped by %s", pi.Title, getReason(pi.Status))
		pi.watcher.onChanged(pi)
		return
	}
	pi.LastExitTime = time.Now()
	if err != nil {
		logger.Errorf("Process '%s' (PID: %d) exited with error: %v", pi.Title, pi.Pid(), err)
		pi.LastExitReason = fmt.Sprintf("exited with error: %v", err)
	} else {
		logger.Infof("Process '%s' (PID: %d) exited normally", pi.Title, pi.Pid())
		pi.LastExitReason = "exited normally"
	}
	pi.Status = models.StatusExited
	pi.process = nil
	pi.autoRestart()
}

/**
 * autoRestart 自动重启进程
 * @param {ProcessInstance} pi - 进程实例
 * @description
 * - 检查重启次数是否超过限制
 * - 增加重启计数
 * - 延迟重启进程
 * - 对于附加的进程，无法重启，只记录日志
 */
func (pi *ProcessInstance) autoRestart() {
	// 重启次数超过限制也不自动重启
	if pi.RestartCount >= pi.watcher.maxRestartCount {
		logger.Warnf("Process '%s' has reached maximum restart count (%d), not restarting",
			pi.Title, pi.watcher.maxRestartCount)
		pi.watcher.onChanged(pi)
		return
	}

	logger.Infof("Process '%s' will restart in %v (restart: %d/%d)",
		pi.Title, time.Second, pi.RestartCount, pi.watcher.maxRestartCount)
	// 延迟重启，避免死锁
	time.AfterFunc(time.Second, func() {
		pi.mutex.Lock()
		defer pi.mutex.Unlock()

		if pi.watcher.onChanged == nil { //只有onChanged!=nil才会进入watchProcess，但存在中途修改的可能性
			return
		}
		if pi.Status == models.StatusStopped || pi.Status == models.StatusError {
			logger.Infof("Process '%s' stopped by %s", pi.Title, getReason(pi.Status))
			pi.watcher.onChanged(pi)
			return
		}
		pi.RestartCount++
		pi.startProcess(context.Background())
		pi.watcher.onChanged(pi)
	})
}
