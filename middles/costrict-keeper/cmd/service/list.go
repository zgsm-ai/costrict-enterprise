package service

import (
	"context"
	"encoding/json"
	"fmt"

	"costrict-keeper/internal/models"
	"costrict-keeper/internal/rpc"
	"costrict-keeper/internal/utils"

	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [service name]",
	Short: "View service status",
	Long:  "View running status of all services. If service name is specified, only show detailed information of that service.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showServiceStatus(context.Background(), args)
	},
}

/**
 * Show service status information
 * @param {context.Context} ctx - Context for request cancellation and timeout
 * @param {[]string} args - Command line arguments, optionally containing service name
 * @returns {error} Returns error if showing status fails, nil on success
 * @description
 * - Loads system configuration from system-spec.json
 * - Shows all services status if no arguments provided
 * - Shows specific service details if name provided
 * @throws
 * - Configuration loading errors
 * - Service status checking errors
 */
func showServiceStatus(ctx context.Context, args []string) {
	rpcClient := rpc.NewHTTPClient(nil)

	if len(args) == 0 {
		// Display all services status via HTTP request
		showAllServices(rpcClient)
	} else {
		// Display detailed information of specified service via HTTP request
		showSpecificService(rpcClient, args[0])
	}
}

type Service_Columns struct {
	Name      string
	Port      int
	Startup   string
	Status    string
	Pid       int
	Healthy   string
	TunPid    string
	TunPort   string
	TunStatus string
	StartTime string
}

/**
 * Show all services status via HTTP request
 * @param {rpc.HTTPClient} client - HTTP client for API requests
 * @returns {error} Returns error if request fails, nil on success
 * @description
 * - Sends GET request to /costrict/api/v1/services endpoint
 * - Parses and displays service information in tabular format
 * - Handles connection errors and API response errors
 * @throws
 * - HTTP request errors
 * - JSON parsing errors
 * - Response processing errors
 */
func showAllServices(client rpc.HTTPClient) error {
	resp, err := client.Get("/costrict/api/v1/services", nil)
	if err != nil {
		fmt.Printf("Failed to call costrict API: %v\n", err)
		return err
	}
	if resp.Error != "" {
		fmt.Printf("Costrict API returned error(%d): %s\n", resp.StatusCode, resp.Error)
		return fmt.Errorf("API error: %s", resp.Error)
	}

	var services []models.ServiceDetail
	if err := json.Unmarshal(resp.Body, &services); err != nil {
		fmt.Printf("Failed to unmarshal service list: %v\n", err)
		return err
	}

	if len(services) == 0 {
		fmt.Println("No services found")
		return nil
	}

	var dataList []*orderedmap.OrderedMap
	for _, svc := range services {
		row := Service_Columns{}
		row.Name = svc.Name
		row.Status = string(svc.Status)
		row.Pid = svc.Pid
		row.Port = svc.Port
		row.StartTime = svc.StartTime
		row.Startup = svc.Spec.Startup
		if svc.Healthy == models.Healthy {
			row.Healthy = "Y"
		} else {
			row.Healthy = "N"
		}
		if svc.Tunnel == nil {
			if svc.Spec.Accessible == "remote" {
				row.TunPid = "0"
				row.TunPort = "0"
				row.TunStatus = "Closed"
			} else {
				row.TunPid = "-"
				row.TunPort = "-"
				row.TunStatus = "-"
			}
		} else {
			row.TunPid = fmt.Sprint(svc.Tunnel.Pid)
			row.TunPort = fmt.Sprint(svc.Tunnel.Pairs[0].MappingPort)
			if svc.Tunnel.Status == models.StatusRunning {
				if svc.Tunnel.Healthy == models.Healthy {
					row.TunStatus = "Opened"
				} else {
					row.TunStatus = "Unhealthy"
				}
			} else {
				row.TunStatus = "Closed"
			}
		}

		recordMap, _ := utils.StructToOrderedMap(row)
		dataList = append(dataList, recordMap)
	}
	utils.PrintFormat(dataList)
	return nil
}

/**
 * Show specific service details with status
 * @param {spec *models.SystemSpecification} System configuration
 * @param {string} name - Name of service
 * @returns {error} Returns error if showing status fails, nil on success
 * @description
 * - Searches for service by name
 * - Displays detailed information with status
 * - Shows version information
 * @throws
 * - Service not found errors
 */
/**
 * Show specific service details via HTTP request
 * @param {rpc.HTTPClient} client - HTTP client for API requests
 * @param {string} name - Name of the service to get details for
 * @returns {error} Returns error if request fails, nil on success
 * @description
 * - Sends GET request to /costrict/api/v1/services/{name} endpoint
 * - Parses and displays detailed service information
 * - Handles connection errors and API response errors
 * @throws
 * - HTTP request errors
 * - JSON parsing errors
 * - Response processing errors
 */
/**
 * Get service detail from API
 * @param {rpc.HTTPClient} client - HTTP client for API requests
 * @param {string} name - Name of the service to get details for
 * @returns {models.ServiceDetail} Service detail information
 * @returns {error} Returns error if request fails, nil on success
 * @description
 * - Sends GET request to /costrict/api/v1/services/{name} endpoint
 * - Returns parsed service detail information
 * - Handles connection errors and API response errors
 * @throws
 * - HTTP request errors
 * - JSON parsing errors
 * - Response processing errors
 */
func getServiceDetail(client rpc.HTTPClient, name string) (*models.ServiceDetail, error) {
	resp, err := client.Get(fmt.Sprintf("/costrict/api/v1/services/%s", name), nil)
	if err != nil {
		fmt.Printf("Failed to call costrict API: %v\n", err)
		return nil, err
	}
	if resp.Error != "" {
		fmt.Printf("Costrict API returned error(%d): %s\n", resp.StatusCode, resp.Error)
		return nil, fmt.Errorf("API error: %s", resp.Error)
	}

	var detail models.ServiceDetail
	if err := json.Unmarshal(resp.Body, &detail); err != nil {
		fmt.Printf("Failed to unmarshal service detail: %v\n", err)
		return nil, err
	}

	return &detail, nil
}

/**
 * Display service detail information
 * @param {models.ServiceDetail} detail - Service detail information to display
 * @param {string} name - Name of the service
 * @description
 * - Displays detailed service information in formatted output
 * - Shows basic info, process info, version info, and tunnel info
 */
func displayServiceDetail(detail *models.ServiceDetail, name string) {
	fmt.Printf("=== Detailed information of service '%s' ===\n", name)
	fmt.Printf("Name: %s\n", detail.Name)
	fmt.Printf("Running status: %s\n", detail.Status)
	fmt.Printf("Port: %d\n", detail.Port)
	fmt.Printf("PID: %d\n", detail.Pid)
	fmt.Printf("Start time: %s\n", detail.StartTime)
	fmt.Printf("Startup command: %s\n", detail.Process.Command)
	fmt.Printf("Startup args: %+v\n", detail.Process.Args)
	fmt.Printf("Startup mode: %s\n", detail.Spec.Startup)
	fmt.Printf("Protocol: %s\n", detail.Spec.Protocol)
	if detail.Spec.Metrics != "" {
		fmt.Printf("Metrics endpoint: %s\n", detail.Spec.Metrics)
	}
	if detail.Spec.Accessible != "" {
		fmt.Printf("Access permission: %s\n", detail.Spec.Accessible)
	}

	// Display version information
	if detail.Component != nil {
		fmt.Printf("Local version: %s\n", detail.Component.Local.Version)
		if detail.Component.Remote.Newest != "" {
			fmt.Printf("Latest server version: %s\n", detail.Component.Remote.Newest)
		} else {
			fmt.Printf("Latest server version: Unable to retrieve\n")
		}
	} else {
		fmt.Printf("Local version: Not installed\n")
		fmt.Printf("Latest server version: Unable to retrieve\n")
	}

	// Display endpoint URL
	if detail.Spec.Protocol != "" && detail.Port > 0 {
		endpointURL := fmt.Sprintf("%s://localhost:%d", detail.Spec.Protocol, detail.Port)
		fmt.Printf("Access URL: %s\n", endpointURL)
	}

	if len(detail.Tunnel.Pairs) > 0 {
		fmt.Printf("Local Port: %d\n", detail.Tunnel.Pairs[0].LocalPort)
		fmt.Printf("Mapping Port: %d\n", detail.Tunnel.Pairs[0].MappingPort)
	}
	fmt.Printf("Tunnel PID: %d\n", detail.Tunnel.Pid)
	fmt.Printf("Tunnel Status: %s\n", detail.Tunnel.Status)
}

/**
 * Show specific service details via HTTP request
 * @param {rpc.HTTPClient} client - HTTP client for API requests
 * @param {string} name - Name of the service to get details for
 * @returns {error} Returns error if request fails, nil on success
 * @description
 * - Gets service detail via HTTP request
 * - Displays detailed service information
 * - Handles connection errors and API response errors
 * @throws
 * - HTTP request errors
 * - JSON parsing errors
 * - Response processing errors
 */
func showSpecificService(client rpc.HTTPClient, name string) error {
	detail, err := getServiceDetail(client, name)
	if err != nil {
		return err
	}
	displayServiceDetail(detail, name)
	return nil
}

func init() {
	serviceCmd.AddCommand(listCmd)
}
