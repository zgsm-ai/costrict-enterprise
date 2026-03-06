package client

import (
	"encoding/json"
	"fmt"
	"time"

	"costrict-keeper/cmd/root"
	"costrict-keeper/internal/config"
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/models"
	"costrict-keeper/internal/rpc"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check server status and health",
	Long:  `Check server status and health by connecting to the costrict server via RPC and calling the check API`,
	Run: func(cmd *cobra.Command, args []string) {
		checkServerStatus()
	},
}

const checkExample = `  # Check server status
  costrict check`

/**
 * Check server status by connecting via RPC and calling check API
 * @returns {void} No return value, outputs results directly or exits on error
 * @description
 * - Creates RPC client to connect to costrict server
 * - Calls /costrict/api/v1/check endpoint via RPC
 * - Handles connection errors and API response errors
 * - Displays check results if successful
 * - Optionally displays configuration if global showConfig flag is true
 * @throws
 * - Connection establishment errors
 * - API request errors
 * - Response parsing errors
 * @example
 * checkServerStatus()
 */
func checkServerStatus() {
	rpcClient := rpc.NewHTTPClient(nil)
	resp, err := rpcClient.Post("/costrict/api/v1/check", nil)
	if err != nil {
		fmt.Printf("Failed to call costrict API: %v\n", err)
		return
	}
	if resp.Error != "" {
		fmt.Printf("Costrict API returned error(%d): %s\n", resp.StatusCode, resp.Error)
		return
	}

	var checkResp models.CheckResponse
	if err := json.Unmarshal(resp.Body, &checkResp); err != nil {
		fmt.Printf("Failed to unmarshal check response: %v\n", err)
		return
	}

	// 成功反序列化，显示检查结果
	displayCheckResults(checkResp)
}

func displayServices(services []models.ServiceDetail) {
	if len(services) == 0 {
		return
	}
	fmt.Printf("=== 服务检查结果 (%d 项) ===\n", len(services))
	for _, svc := range services {
		statusIcon := "✅"
		if svc.Healthy != models.Healthy || svc.Status != "running" {
			statusIcon = "❌"
		}

		fmt.Printf("%s 服务: %s", statusIcon, svc.Name)
		if svc.Pid > 0 {
			fmt.Printf(" (PID: %d)", svc.Pid)
		}
		if svc.Port > 0 {
			fmt.Printf(" 端口: %d", svc.Port)
		}
		if svc.Process.RestartCount > 0 {
			fmt.Printf(" 重启次数: %d", svc.Process.RestartCount)
		}
		fmt.Printf(" 状态: %s", svc.Status)
		if svc.Healthy == models.Healthy {
			fmt.Printf(" 健康")
		} else {
			fmt.Printf(" 不健康")
		}
		fmt.Println()
		displayTunnel(svc.Tunnel)
	}
	fmt.Println()
}

func displayTunnel(tunnel *models.TunnelDetail) {
	if tunnel == nil {
		return
	}
	statusIcon := "✅"
	if tunnel.Healthy != models.Healthy {
		statusIcon = "❌"
	}
	fmt.Printf("  %s 隧道: %s", statusIcon, tunnel.Name)
	if tunnel.Pid > 0 {
		fmt.Printf(" (PID: %d)", tunnel.Pid)
	}
	fmt.Printf(" 隧道数: %d", len(tunnel.Pairs))
	for _, tun := range tunnel.Pairs {
		fmt.Printf(" (本地端口: %d -> 映射端口: %d)", tun.LocalPort, tun.MappingPort)
	}
	fmt.Printf(" 状态: %s", tunnel.Status)
	if tunnel.Healthy == models.Healthy {
		fmt.Printf(" 健康")
	} else {
		fmt.Printf(" 不健康")
	}
	fmt.Println()
}

func displayComponents(components []models.ComponentDetail) {
	if len(components) == 0 {
		return
	}
	fmt.Printf("=== 组件检查结果 (%d 项) ===\n", len(components))
	for _, cpn := range components {
		statusIcon := "✅"
		if !cpn.Installed || cpn.NeedUpgrade {
			statusIcon = "❌"
		}

		fmt.Printf("%s %s", statusIcon, cpn.Name)
		if cpn.Installed {
			fmt.Printf(" (本地版本: %s", cpn.Local.Version)
			if cpn.NeedUpgrade {
				fmt.Printf(" -> 远程版本: %s) 需要升级", cpn.Remote.Newest)
			} else {
				fmt.Printf(") 已安装")
			}
		} else {
			fmt.Printf(" 未安装")
		}
		fmt.Println()
	}
	fmt.Println()
}

/**
 * Display formatted check results to user
 * @param {models.CheckResponse} results - Check results from server
 * @description
 * - Formats and displays overall system status
 * - Shows service, process, tunnel, and component check results
 * - Displays midnight rooster status
 * - Shows summary statistics
 * - Optionally displays configuration if global showConfig flag is true
 */
func displayCheckResults(results models.CheckResponse) {
	fmt.Println("=== Costrict Server Status Check ===")
	fmt.Println()

	// Display timestamp
	fmt.Printf("检查时间: %s\n", results.Timestamp.Format(time.RFC3339))
	fmt.Printf("云端地址: %s\n", config.GetBaseURL())
	fmt.Printf("安装目录: %s\n", env.CostrictDir)
	fmt.Println()

	// Display overall status
	statusIcon := ""
	switch results.OverallStatus {
	case "warning":
		statusIcon = "⚠️"
	case "error":
		statusIcon = "❌"
	case "healthy":
		statusIcon = "✅"
	default:
		statusIcon = "❓"
	}
	fmt.Printf("%s 总体状态: %s\n", statusIcon, results.OverallStatus)
	fmt.Println()

	// Display statistics
	fmt.Printf("总检查项: %d\n", results.TotalChecks)
	fmt.Printf("通过检查: %d\n", results.PassedChecks)
	fmt.Printf("失败检查: %d\n", results.FailedChecks)
	fmt.Println()

	displayServices(results.Services)
	displayComponents(results.Components)

	fmt.Println("=== 检查完成 ===")
}

func init() {
	checkCmd.Flags().SortFlags = false
	checkCmd.Example = checkExample
	root.RootCmd.AddCommand(checkCmd)
}
