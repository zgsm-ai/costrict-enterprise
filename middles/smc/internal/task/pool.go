package task

import (
	"encoding/json"

	"github.com/zgsm-ai/smc/internal/utils"
)

const (
	REQ_POOLS      = "/taskd/api/v1/pools"
	REQ_POOLS_ONCE = "/taskd/api/v1/pools/{0}" //GET
)

/**
 * Summary of task resource pool
 */
type TaskPoolSummary struct {
	PoolId      string `json:"pool_id"`
	Engine      string `json:"engine"`
	Config      string `json:"config"`
	Description string `json:"description"`
	MaxWaiting  int    `json:"max_waiting"`
	MaxRunning  int    `json:"max_running"`
	Waiting     int    `json:"waiting"`
	Running     int    `json:"running"`
}

/**
 * Resource utilization information
 */
type ResourceItem struct {
	Name     string `json:"name"`     //Resource name
	Capacity string `json:"capacity"` //Configuration amount
	Allocate string `json:"allocate"` //Allocated amount
	Remain   string `json:"remain"`   //Actual remaining
}

/**
 * Detailed task pool information with running tasks
 */
type TaskPoolDetail struct {
	PoolId      string         `json:"pool_id"`
	Engine      string         `json:"engine"`
	Config      string         `json:"config"`
	Description string         `json:"description"`
	MaxWaiting  int            `json:"max_waiting"`
	MaxRunning  int            `json:"max_running"`
	Waiting     int            `json:"waiting"`
	Running     int            `json:"running"`
	Tasks       []TaskSummary  `json:"tasks,omitempty"`
	Resources   []ResourceItem `json:"resources,omitempty"`
}

/**
 * Basic pool information for create/update operations
 */
type PoolBasic struct {
	PoolId      string `json:"pool_id"`
	Engine      string `json:"engine"`
	Config      string `json:"config"`
	Running     int    `json:"running"`
	Waiting     int    `json:"waiting"`
	Description string `json:"description"`
}

/**
 * List all available task pools
 * @param ss active session
 * @return []TaskPoolSummary list of pool summaries
 * @return error if operation fails
 */
func ListPools(ss *utils.Session) ([]TaskPoolSummary, error) {
	rspData, err := ss.GetData(utils.ApiPath(REQ_POOLS), nil)
	if err != nil {
		return nil, err
	}
	var result []TaskPoolSummary
	if err := json.Unmarshal(rspData, &result); err != nil {
		return nil, err
	}
	return result, nil
}

/**
 * Get detailed pool information including tasks
 * @param ss active session
 * @param poolId identifier of the pool
 * @param verbose show verbose details
 * @return *TaskPoolDetail detailed pool information
 * @return error if operation fails
 */
func GetPool(ss *utils.Session, poolId string, verbose bool) (*TaskPoolDetail, error) {
	paras := utils.Json{
		"verbose": verbose,
	}
	rspData, err := ss.GetData(utils.ApiPath(REQ_POOLS_ONCE, poolId), paras)
	if err != nil {
		return nil, err
	}
	var result TaskPoolDetail
	if err := json.Unmarshal(rspData, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

/**
 * Add new task pool
 * @param ss active session
 * @param p pool configuration to add
 * @return []byte raw response data
 * @return error if operation fails
 */
func AddPool(ss *utils.Session, p *PoolBasic) ([]byte, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return []byte{}, err
	}
	return ss.Post(utils.ApiPath(REQ_POOLS), data)
}

/**
 * Remove existing task pool
 * @param ss active session
 * @param poolId identifier of pool to remove
 * @return error if operation fails
 */
func RemovePool(ss *utils.Session, poolId string) error {
	_, err := ss.Delete(utils.ApiPath(REQ_POOLS_ONCE, poolId))
	return err
}
