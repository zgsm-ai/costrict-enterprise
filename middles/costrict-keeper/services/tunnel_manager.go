package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"costrict-keeper/internal/config"
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/httpc"
	"costrict-keeper/internal/logger"
	"costrict-keeper/internal/models"
	"costrict-keeper/internal/proc"
	"costrict-keeper/internal/utils"
)

// 端口分配请求
type PortAllocationRequest struct {
	ClientId   string `json:"clientId"`
	AppName    string `json:"appName"`
	ClientPort int    `json:"clientPort"`
}

// 端口分配响应
type PortAllocationResponse struct {
	ClientId    string `json:"clientId"`
	AppName     string `json:"appName"`
	ClientPort  int    `json:"clientPort"`
	MappingPort int    `json:"mappingPort"`
}

type PortQueryResponse struct {
	MappingPort int `json:"mappingPort"`
}

type TunnelArgs struct {
	AppName     string
	LocalPort   int
	MappingPort int
	Pairs       []models.PortPair
	RemoteAddr  string
	ProcessName string
	ProcessPath string
}

type TunnelCache struct {
	Name        string            `json:"name"`        // service name
	Pairs       []models.PortPair `json:"pairs"`       // Port pairs
	Status      models.RunStatus  `json:"status"`      // tunnel status(running/stopped/error/exited)
	CreatedTime time.Time         `json:"createdTime"` // creation time
	Pid         int               `json:"pid"`         // process ID of the tunnel
}

type TunnelInstance struct {
	name        string                // service name
	pairs       []models.PortPair     // Port pairs
	status      models.RunStatus      // tunnel status(running/stopped/error/exited)
	createdTime time.Time             // creation time
	pi          *proc.ProcessInstance // Process cotun.exe
}

/**
 * Create new tunnel instance with default values
 * @param {string} name - Application name for the tunnel
 * @param {int} port - Local port number for the tunnel
 * @returns {*TunnelInstance} Returns new tunnel instance with initialized values
 * @description
 * - Creates new tunnel with specified name and port
 * - Initializes default values: mapping port 0, HTTP protocol, stopped status
 * - Sets creation time to current time and PID to 0
 * - Tunnel is not started yet, just created with initial configuration
 * @example
 * tun := CreateTunnel("myapp", []int{8080})
 */
func CreateTunnel(appName string, ports []int) *TunnelInstance {
	pairs := []models.PortPair{}
	for _, p := range ports {
		pairs = append(pairs, models.PortPair{LocalPort: p, MappingPort: 0})
	}
	tun := &TunnelInstance{
		name:        appName,
		pairs:       pairs,
		status:      "exited",
		createdTime: time.Now().Local(),
	}
	return tun
}

/**
 * Get title string for tunnel instance
 * @returns {string} Returns formatted title string
 * @description
 * - Creates formatted title with name, local port, and mapping port
 * - Format: {name}:{localPort}->{mappingPort}
 * - Used for logging and display purposes
 * @private
 * @example
 * title := tunnelInstance.getTitle()
 * // Returns: "myapp:8080->9000"
 */
func (ti *TunnelInstance) getTitle() string {
	return fmt.Sprintf("%s:%d->%d", ti.name, ti.pairs[0].LocalPort, ti.pairs[0].MappingPort)
}

func (ti *TunnelInstance) toJSON() (string, error) {
	cache := TunnelCache{
		Name:        ti.name,
		Pid:         0,
		Status:      ti.status,
		CreatedTime: ti.createdTime,
		Pairs:       ti.pairs,
	}
	if ti.pi != nil {
		cache.Pid = ti.pi.Pid()
	}
	data, err := json.MarshalIndent(&cache, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

/**
 * Generate cache file name for tunnel instance
 * @param {*TunnelInstance} tun - Tunnel instance to generate cache file name for
 * @returns {string} Returns the full path to the cache file
 * @description
 * - Constructs cache file path using tunnel name and local port
 * - File name format: {name}-{port}.json
 * - Cache files are stored in CostrictDir/cache/tunnels directory
 * @example
 * fname := tunnelMgr.getCacheFname(tunnelInstance)
 * // Returns: /path/to/costrict/cache/tunnels/myapp-8080.json
 */
func (tun *TunnelInstance) getCacheFname() string {
	return filepath.Join(env.CostrictDir, "cache", "tunnels", fmt.Sprintf("%s.json", tun.name))
}

/**
 * Request port mapping from tunnel manager service
 * @param {*TunnelInstance} tun - Tunnel instance to request mapping for
 * @returns {error} Returns error if request fails, nil on success
 * @description
 * - Creates HTTP client and prepares port allocation request
 * - Includes machine ID, app name and client port in request body
 * - Adds authentication headers from config
 * - Sends POST request to tunnel manager service
 * - Handles HTTP response and error statuses
 * - Parses JSON response and updates tunnel mapping port
 * - Logs detailed error information on failures
 * @throws
 * - JSON marshaling errors for request body
 * - HTTP request creation errors
 * - Network request errors
 * - Non-200 HTTP status codes
 * - JSON parsing errors for response
 */
func (tun *TunnelInstance) allocMappingPort() error {
	tun.pairs[0].MappingPort = 0

	// 创建请求 body
	requestBody := PortAllocationRequest{
		ClientId:   config.GetMachineID(),
		AppName:    tun.name,
		ClientPort: tun.pairs[0].LocalPort,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", config.Cloud().TunManagerUrl+"/ports", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	authKey, authValue := config.GetAuthHeader()
	req.Header.Set(authKey, authValue)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpc.GetClient().Do(req)
	if err != nil {
		logger.Errorf("allocMappingPort failed - URL: %s, Body: %s, Error: %v", req.URL.String(), string(jsonBody), err)
		return fmt.Errorf("failed to request manager: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Errorf("Failed to read response body: %v", err)
		} else {
			logger.Errorf("Failed to request URL: %s, Body: %s, Status Code: %d, Response Body: %s", req.URL.String(), string(jsonBody), resp.StatusCode, string(bodyBytes))
		}
		return fmt.Errorf("manager returned error status code: %d", resp.StatusCode)
	}

	var result PortAllocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf("Failed to parse response: %v", err)
		return fmt.Errorf("failed to parse response: %w", err)
	}
	tun.pairs[0].MappingPort = result.MappingPort
	logger.Infof("Successfully applied for port mapping, result: %+v", result)
	return nil
}

func (tun *TunnelInstance) GetPid() int {
	if tun.pi == nil {
		return 0
	}
	return tun.pi.Pid()
}

func (tun *TunnelInstance) GetDetail() models.TunnelDetail {
	detail := models.TunnelDetail{
		Name:        tun.name,
		Status:      tun.status,
		CreatedTime: tun.createdTime,
		Pairs:       tun.pairs,
		Pid:         0,
		Healthy:     models.Healthy,
	}
	if tun.pi != nil {
		detail.Pid = tun.pi.Pid()
		detail.Healthy = tun.GetHealthy()
	}
	return detail
}

/**
 * Start tunnel process and initialize connection
 * @param {*TunnelInstance} tunnel - Tunnel instance to start
 * @returns {error} Returns error if any step fails, nil on success
 * @description
 * - Sets tunnel status to error initially (for safety)
 * - Requests port mapping from tunnel manager service
 * - Creates process instance with tunnel configuration
 * - Sets restart callback to update PID and save tunnel on restart
 * - Starts tunnel process via process manager
 * - Updates tunnel status, PID and creation time on success
 * - Saves tunnel state to cache via defer function
 * - Logs successful tunnel creation with details
 * @throws
 * - Port mapping request errors
 * - Process instance creation errors
 * - Process start errors
 */
func (tun *TunnelInstance) OpenTunnel(ctx context.Context) error {
	if tun.status == models.StatusRunning {
		logger.Infof("Tunnel (%s) has been started, PID: %d", tun.getTitle(), tun.pi.Pid())
		return nil
	}
	var err error

	defer func() {
		tun.saveTunnel()
	}()
	tun.status = models.StatusError

	if err := tun.allocMappingPort(); err != nil {
		logger.Errorf("Allocate mapping port failed: %v", err)
		return err
	}

	tun.pi, err = tun.createProcessInstance()
	if err != nil {
		logger.Errorf("Failed to get command info: %v", err)
		return err
	}
	if env.Daemon {
		tun.pi.SetWatcher(3, func(pi *proc.ProcessInstance) {
			switch pi.Status {
			case models.StatusExited, models.StatusError:
				tun.status = models.StatusError
			default: //models.StatusStopped, models.StatusRunning
				tun.status = pi.Status
			}
			tun.saveTunnel()
		})
	}
	if err := tun.pi.StartProcess(ctx); err != nil {
		return err
	}
	tun.status = models.StatusRunning
	tun.createdTime = tun.pi.StartTime

	logger.Infof("Successfully created tunnel (%s), process: %s (PID: %d)",
		tun.getTitle(), tun.pi.ProcessName, tun.pi.Pid())
	return nil
}

/**
 * Stop tunnel process and clean up resources
 * @description
 * - Stops tunnel process via process manager if it exists
 * - Logs success or failure of tunnel stop operation
 * - Frees the local port used by the tunnel
 * - Cleans up tunnel cache and state
 * - Updates tunnel status to stopped and resets PID
 * - Used for graceful tunnel shutdown
 * @private
 * @example
 * tunnelInstance.closeTunnel()
 */
func (tun *TunnelInstance) CloseTunnel() error {
	if tun.pi == nil {
		return nil
	}
	logger.Infof("Tunnel '%s' (PID: %d) will be closed", tun.getTitle(), tun.pi.Pid())
	tun.status = models.StatusStopped
	tun.pi.StopProcess()
	utils.FreePort(tun.pairs[0].LocalPort)
	tun.removeTunnelFile()
	return nil
}

func (tun *TunnelInstance) CheckTunnel() models.HealthyStatus {
	if tun.status != models.StatusRunning {
		return models.Unavailable
	}
	if tun.pi == nil {
		return models.Unavailable
	}
	if status := tun.pi.CheckProcess(); status != models.Healthy {
		tun.status = models.StatusExited
		tun.removeTunnelFile()
		return status
	}
	return models.Healthy
}

func (tun *TunnelInstance) GetHealthy() models.HealthyStatus {
	if tun.status != models.StatusRunning {
		return models.Unavailable
	}
	if tun.pi == nil {
		return models.Unavailable
	}
	pid := tun.pi.Pid()
	if pid == 0 {
		return models.Unavailable
	}
	running, err := utils.IsProcessRunning(pid)
	if err != nil || !running {
		return models.Unavailable
	}
	return models.Healthy
}

/**
 * Get process instance for tunnel execution
 * @param {*TunnelInstance} tunnel - Tunnel instance to create process for
 * @returns {(*ProcessInstance, error)} Returns process instance and error if any
 * @description
 * - Reads tunnel configuration from config
 * - Adjusts process name for Windows (.exe extension)
 * - Creates TunnelArgs with tunnel-specific parameters
 * - Uses text/template to process command and arguments from config
 * - Generates command line with substituted template variables
 * - Returns new ProcessInstance with generated command and args
 * - Template variables include: RemoteAddr, MappingPort, LocalPort, ProcessName, ProcessPath
 * @throws
 * - Command line generation errors
 */
func (tun *TunnelInstance) createProcessInstance() (*proc.ProcessInstance, error) {
	cfg := config.App()
	name := cfg.Tunnel.ProcessName
	if runtime.GOOS == "windows" {
		name = fmt.Sprintf("%s.exe", cfg.Tunnel.ProcessName)
	}
	args := TunnelArgs{
		AppName:     tun.name,
		LocalPort:   tun.pairs[0].LocalPort,
		MappingPort: tun.pairs[0].MappingPort,
		RemoteAddr:  config.Cloud().TunnelUrl,
		ProcessName: name,
		ProcessPath: filepath.Join(env.CostrictDir, "bin", name),
	}
	command, cmdArgs, err := utils.GetCommandLine(cfg.Tunnel.Command, cfg.Tunnel.Args, args)
	if err != nil {
		logger.Errorf("Tunnel startup settings are incorrect, setting: %+v", cfg.Tunnel)
		return nil, err
	}
	return proc.NewProcessInstance("tunnel "+tun.name, name, command, cmdArgs), nil
}

/**
 * Save tunnel instance to cache file
 * @param {*TunnelInstance} tun - Tunnel instance to save
 * @returns {error} Returns error if save operation fails, nil on success
 * @description
 * - Creates cache directory if it doesn't exist
 * - Serializes tunnel instance to JSON format
 * - Writes JSON data to cache file with 0644 permissions
 * - Logs error if save operation fails
 * - Uses inner function for better error handling
 * - File path is generated using getCacheFname()
 * @throws
 * - Directory creation errors
 * - JSON serialization errors
 * - File write errors
 */
func (tun *TunnelInstance) saveTunnel() error {
	err := func() error {
		tunnelsDir := filepath.Join(env.CostrictDir, "cache", "tunnels")
		if err := os.MkdirAll(tunnelsDir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory: %w", err)
		}

		data, err := tun.toJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize tunnel info: %w", err)
		}
		filePath := tun.getCacheFname()
		if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
			return fmt.Errorf("failed to write tunnel info file: %w", err)
		}
		return nil
	}()
	if err != nil {
		logger.Errorf("Save tunnel failed: %v", err)
	}
	return err
}

/**
 * Remove tunnel cache file
 * @param {*TunnelInstance} tun - Tunnel instance to clean
 * @returns {error} Returns error if file deletion fails, nil on success
 * @description
 * - Generates cache file path using getCacheFname()
 * - Checks if cache file exists using os.Stat()
 * - Removes cache file if it exists
 * - Logs error if deletion fails
 * - Silently returns if file doesn't exist (no error)
 * - Used when closing tunnels to clean up cached data
 * @throws
 * - File deletion errors
 */
func (tun *TunnelInstance) removeTunnelFile() error {
	filePath := tun.getCacheFname()
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			logger.Errorf("Failed to delete cache file: %v", err)
			return err
		}
	}
	return nil
}
