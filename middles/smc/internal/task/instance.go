package task

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/zgsm-ai/smc/internal/utils"
	"gopkg.in/yaml.v3"
)

const (
	REQ_TASKS       = "/taskd/api/v1/tasks"
	REQ_TASKS_ONCE  = "/taskd/api/v1/tasks/{0}" //GET
	REQ_TASK_TAGS   = "/taskd/api/v1/tasks/{0}/tags"
	REQ_TASK_LOGS   = "/taskd/api/v1/tasks/{0}/logs"
	REQ_TASK_STATUS = "/taskd/api/v1/tasks/{0}/status"
	REQ_TASK_STOP   = "/taskd/api/v1/tasks/{0}"
)

type TaskSummary struct {
	UUID        string     `json:"uuid"`                   //Task unique ID
	Name        string     `json:"name"`                   //Task name
	Status      string     `json:"status"`                 //Task status
	CreatedBy   string     `json:"created_by"`             //User mark
	Pool        string     `json:"pool,omitempty"`         //Resource pool name
	Warning     string     `json:"warning,omitempty"`      //Warning message
	Error       string     `json:"error,omitempty"`        //Error message
	Tags        string     `json:"tags,omitempty"`         //Tag information
	CreateTime  *time.Time `json:"create_time,omitempty"`  //Enqueue time
	StartTime   *time.Time `json:"start_time,omitempty"`   //Start time
	RunningTime *time.Time `json:"running_time,omitempty"` //Actual running start time
	EndTime     *time.Time `json:"end_time,omitempty"`     //End time
}

type TaskMetadata struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Template    string `json:"template"`
	Project     string `json:"project"`
	Pool        string `json:"pool"`
	Namespace   string `json:"namespace"`
	Timeout     string `json:"timeout,omitempty"`
	Quotas      string `json:"quotas,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Args        string `json:"args,omitempty"`
	Extra       string `json:"extra,omitempty"`
	Callback    string `json:"callback,omitempty"`
	YamlContent string `json:"yaml_content,omitempty"`
	CreatedBy   string `json:"created_by"`
	CreateTime  string `json:"create_time"`
	StartTime   string `json:"start_time,omitempty"`
	UpdateTime  string `json:"update_time,omitempty"`
	EndTime     string `json:"end_time,omitempty"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	EndLog      string `json:"end_log,omitempty"`
}

type ListTasksResult struct {
	Total int            `json:"total"`
	List  []TaskMetadata `json:"list"`
}

type TaskLogsArgs struct {
	Entity     string `form:"entity,omitempty"`
	Tail       int64  `form:"tail,omitempty"`
	Follow     bool   `form:"follow,omitempty"`
	Timestamps bool   `form:"timestamps,omitempty"`
}

type EntityLogs struct {
	Entity    string `json:"entity"`
	Completed bool   `json:"completed"`
	Logs      string `json:"logs"`
}

type TaskLogsResult struct {
	Uuid     string       `json:"uuid,omitempty"`
	Status   string       `json:"status,omitempty"`
	Entities []EntityLogs `json:"entities,omitempty"`
}

type TaskStatusResult struct {
	Uuid   string `json:"uuid"`
	Status string `json:"status"`
}

func GetTask(ss *utils.Session, uuid string, verbose bool) (TaskMetadata, error) {
	var tm TaskMetadata
	rspData, err := ss.GetData(utils.ApiPath(REQ_TASKS_ONCE, uuid), nil)
	if err != nil {
		return tm, err
	}
	if err := json.Unmarshal(rspData, &tm); err != nil {
		return tm, err
	}

	return tm, nil
}

/**
 *	Continuously get logs using long connection
 */
func followLogs(ss *utils.Session, uuid, podName string, timestamps bool) error {
	jsonData := utils.Json{
		"pod":        podName,
		"follow":     true,
		"timestamps": timestamps,
	}
	body, err := ss.GetBody(utils.ApiPath(REQ_TASK_LOGS, uuid), jsonData)
	if err != nil {
		return err
	}
	defer body.Close()

	// Read and process log output line by line
	buf := make([]byte, 1024)
	for {
		n, err := body.Read(buf)
		if err != nil {
			if err == io.EOF {
				// Reached end of stream
				return nil
			}
			return err
		}
		// Process log output
		fmt.Print(string(buf[:n]))
	}
}

/**
 *	Print task logs
 *  If follow parameter is specified, implement long polling by looping API calls
 */
func GetTaskLogs(ss *utils.Session, uuid string, arg *TaskLogsArgs) error {
	if arg.Follow {
		return followLogs(ss, uuid, arg.Entity, arg.Timestamps)
	}
	jsonData := utils.Json{
		"entity":     arg.Entity,
		"tail":       arg.Tail,
		"timestamps": arg.Timestamps,
	}
	rspData, err := ss.GetData(utils.ApiPath(REQ_TASK_LOGS, uuid), jsonData)
	if err != nil {
		return err
	}
	var taskLogs TaskLogsResult
	if err := json.Unmarshal(rspData, &taskLogs); err != nil {
		return err
	}
	for i, podLog := range taskLogs.Entities {
		if i != len(taskLogs.Entities)-1 {
			fmt.Println(color.YellowString("POD %s:", podLog.Entity))
		} else {
			fmt.Println(color.YellowString("%s:", podLog.Entity))
		}
		fmt.Println(podLog.Logs)
	}
	return nil
}

/**
 *	Query task status
 */
func GetTaskStatus(ss *utils.Session, uuid string) ([]byte, error) {
	rspData, err := ss.GetData(utils.ApiPath(REQ_TASK_STATUS, uuid), nil)
	if err != nil {
		return nil, err
	}
	var status TaskStatusResult
	if err := json.Unmarshal(rspData, &status); err != nil {
		return nil, err
	}
	return yaml.Marshal(status)
}

func StartTask(ss *utils.Session, ti *TaskMetadata) ([]byte, error) {
	body, err := json.Marshal(ti)
	if err != nil {
		log.Printf("StartTask(...) Marshal error: %v\n", err)
		return []byte{}, err
	}
	data, err := ss.Post(utils.ApiPath(REQ_TASKS), body)
	if err != nil {
		log.Printf("StartTask(...) Post(%s) error: %v\n", utils.ApiPath(REQ_TASKS), err)
		return data, err
	}
	return data, nil
}

/**
 *	Stop task
 */
func StopTask(ss *utils.Session, uuid string) error {
	_, err := ss.Delete(utils.ApiPath(REQ_TASKS_ONCE, uuid))
	return err
}

/**
 *	Add tag
 */
func SetTaskTags(ss *utils.Session, uuid, key, value string) (map[string]string, error) {
	jsonData := utils.Json{
		key: value,
	}
	tags := make(map[string]string)
	data, err := ss.PostJson(utils.ApiPath(REQ_TASK_TAGS, uuid), jsonData)
	if err != nil {
		return tags, err
	}
	err = json.Unmarshal(data, &tags)
	return tags, err
}

func GetTaskTags(ss *utils.Session, uuid string) (map[string]string, error) {
	tags := make(map[string]string)
	data, err := ss.PostJson(utils.ApiPath(REQ_TASK_TAGS, uuid), nil)
	if err != nil {
		return tags, err
	}
	err = json.Unmarshal(data, &tags)
	return tags, err
}

type ListTasksArgs struct {
	Uuid      string `json:"uuid"`
	Project   string `json:"project"`
	Template  string `json:"template"`
	Engine    string `json:"engine"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Pool      string `json:"pool"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	Sort      string `json:"sort"`
	Verbose   bool   `json:"verbose"`
}

func ListTasks(ss *utils.Session, arg *ListTasksArgs) (*ListTasksResult, error) {
	jsonData := utils.Json{
		"uuid":      arg.Uuid,
		"template":  arg.Template,
		"project":   arg.Project,
		"namespace": arg.Namespace,
		"name":      arg.Name,
		"pool":      arg.Pool,
		"status":    arg.Status,
		"page":      arg.Page,
		"pageSize":  arg.PageSize,
		"sort":      "id DESC",
		"verbose":   arg.Verbose,
	}
	rspData, err := ss.GetData(utils.ApiPath(REQ_TASKS), jsonData)
	if err != nil {
		return nil, err
	}
	var result ListTasksResult
	if err := json.Unmarshal(rspData, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
