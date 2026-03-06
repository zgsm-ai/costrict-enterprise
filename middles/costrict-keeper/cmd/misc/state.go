package client

import (
	"encoding/json"
	"fmt"
	"time"

	"costrict-keeper/cmd/root"
	"costrict-keeper/internal/config"
	"costrict-keeper/internal/models"
	"costrict-keeper/internal/rpc"

	"github.com/spf13/cobra"
)

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Displays costrict server configuration and states",
	Long:  `Displays costrict server configuration and states`,
	Run: func(cmd *cobra.Command, args []string) {
		showServerState()
	},
}

const stateExample = `  # Display server configuration and states
  costrict state`

func showServerState() {
	rpcClient := rpc.NewHTTPClient(nil)
	resp, err := rpcClient.Get("/costrict/api/v1/state", nil)
	if err != nil {
		fmt.Printf("Failed to call costrict API: %v\n", err)
		return
	}
	if resp.Error != "" {
		fmt.Printf("Costrict API returned error(%d): %s\n", resp.StatusCode, resp.Error)
		return
	}

	var respState models.ServerState
	if err := json.Unmarshal(resp.Body, &respState); err != nil {
		fmt.Printf("Failed to unmarshal state response: %v\n", err)
		return
	}

	// 成功反序列化，显示检查结果
	displayStates(respState)
}

func displayStates(results models.ServerState) {
	fmt.Println("=== Costrict Server States ===")
	fmt.Println()

	// Display timestamp
	fmt.Printf("启动时间: %s\n", results.StartTime.Format(time.RFC3339))
	fmt.Println()

	fmt.Println("=== 环境信息 ===")
	fmt.Printf("云端地址: %s\n", config.GetBaseURL())
	fmt.Printf("安装目录: %s\n", results.Env.CostrictDir)
	fmt.Printf("侦听端口: %v\n", results.Env.ListenPort)
	fmt.Printf("软件版本: %v\n", results.Env.Version)
	fmt.Println()

	// Display midnight rooster status
	fmt.Println("=== 半夜鸡叫信息 ===")
	fmt.Printf("状态: %s\n", results.MidnightRooster.Status)
	fmt.Printf("下次检查时间: %s\n", results.MidnightRooster.NextCheckTime.Format(time.RFC3339))
	fmt.Println()

	fmt.Println("=== 端口分配信息 ===")
	fmt.Printf("可分配范围: [%d, %d]\n", results.PortAlloc.Min, results.PortAlloc.Max)
	fmt.Printf("已分配端口(%d): %v\n", len(results.PortAlloc.Allocates), results.PortAlloc.Allocates)
	fmt.Println()

	fmt.Println("=== 配置 ===")
	fmt.Printf("SystemSpec:\n%s\n", results.Config.SystemSpec)
	fmt.Printf("Software:\n%s\n", results.Config.Software)
	fmt.Printf("Auth:\n%s\n", results.Config.Auth)
	fmt.Printf("Cloud:\n%s\n", results.Config.Cloud)
	fmt.Println()
}

func init() {
	stateCmd.Flags().SortFlags = false
	stateCmd.Example = stateExample
	root.RootCmd.AddCommand(stateCmd)
}
