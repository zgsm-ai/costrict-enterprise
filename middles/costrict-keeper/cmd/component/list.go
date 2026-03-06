package component

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"costrict-keeper/internal/config"
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/utils"
	"costrict-keeper/services"

	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cobra"
)

var optServer bool

var listCmd = &cobra.Command{
	Use:   "list [component name]",
	Short: "List information of all components",
	Long: `List information of all components, including local version and latest server version.
If component name is specified, only show detailed information of that component.
When --server flag is set, display all available packages on the server with their version information.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.LoadSpec(); err != nil {
			fmt.Printf("Costrict is uninitialized")
			return
		}

		listInfo(context.Background(), args)
	},
}

/**
 * List component information with version details
 * @param {context.Context} ctx - Context for request cancellation and timeout
 * @param {[]string} args - Command line arguments, optionally containing component name
 * @returns {error} Returns error if listing fails, nil on success
 * @description
 * - Loads system configuration from system-spec.json
 * - Lists all components with version info if no arguments provided
 * - Lists specific component details if name provided
 * - Shows local version and remote version
 * @throws
 * - Configuration loading errors
 * - Version checking errors
 */
func listInfo(ctx context.Context, args []string) {
	fmt.Printf("------------------------------------------\n")
	fmt.Printf("云端地址: %s\n", config.GetBaseURL())
	fmt.Printf("安装目录: %s\n", env.CostrictDir)
	fmt.Printf("------------------------------------------\n")
	if optServer {
		// 显示远程包列表
		if err := listRemotePackages(); err != nil {
			fmt.Printf("Failed to list remote packages: %v\n", err)
			return
		}
		return
	}
	if len(args) == 0 {
		listAllComponents()
	} else {
		listSpecificComponent(args[0])
	}
}

/**
 *	Fields displayed in list format
 */
type Component_Columns struct {
	Name        string `json:"name"`
	Local       string `json:"local"`
	Remote      string `json:"remote"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

/**
 * List all components with detailed information
 * @param {spec *models.SystemSpecification} System configuration
 * @returns {error} Returns error if listing fails, nil on success
 * @description
 * - Lists components with local and remote versions
 * - Uses tabwriter for formatted output
 */
func listAllComponents() {
	manager := services.GetComponentManager()
	manager.Init()
	components := manager.GetComponents(true, true)
	if len(components) == 0 {
		fmt.Println("No components found")
		return
	}
	var dataList []*orderedmap.OrderedMap
	for _, ci := range components {
		cpn := ci.GetDetail()
		row := Component_Columns{}
		row.Name = cpn.Spec.Name
		row.Path = "-"
		row.Local = cpn.Local.Version
		row.Remote = cpn.Remote.Newest
		if cpn.Installed {
			row.Path = cpn.Local.FileName
			row.Description = cpn.Local.Description
		}
		recordMap, _ := utils.StructToOrderedMap(row)
		dataList = append(dataList, recordMap)
	}

	utils.PrintFormat(dataList)
}

/**
 * List specific component details
 * @param {spec *models.SystemSpecification} System configuration
 * @param {string} name - Name of component
 * @returns {error} Returns error if listing fails, nil on success
 * @description
 * - Searches for component by name
 * - Displays detailed information with version comparison
 * @throws
 * - Component not found errors
 */
func listSpecificComponent(name string) {
	manager := services.GetComponentManager()
	manager.Init()

	ci := manager.GetComponent(name)
	if ci == nil {
		fmt.Printf("Component '%s' not found\n", name)
		return
	}
	cpn := ci.GetDetail()
	fmt.Printf("=== Detailed information of component '%s' ===\n", name)
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Need upgrade: %v\n", cpn.NeedUpgrade)
	fmt.Printf("Version range: %s\n", cpn.Spec.Version)

	// Display version information
	if cpn.Local.Version != "" {
		fmt.Printf("Local version: %s\n", cpn.Local.Version)
	} else {
		fmt.Printf("Local version: Not installed\n")
	}
	if cpn.Remote.Newest != "" {
		fmt.Printf("Latest server version: %s\n", cpn.Remote.Newest)
	} else {
		fmt.Printf("Latest server version: Unable to retrieve\n")
	}
}

// formatSize 格式化文件大小
func formatSize(size uint64) string {
	if size < 1024 {
		return strconv.FormatUint(size, 10) + "B"
	} else if size < 1024*1024 {
		return strconv.FormatUint(size/1024, 10) + "KB"
	} else if size < 1024*1024*1024 {
		return strconv.FormatUint(size/(1024*1024), 10) + "MB"
	} else {
		return strconv.FormatUint(size/(1024*1024*1024), 10) + "GB"
	}
}

// 获取包详细元数据信息
func getPackageDetailInfo(u *utils.Upgrader, infoUrl string) (*utils.PackageVersion, error) {
	data, err := u.GetBytes(u.BaseUrl+infoUrl, nil)
	if err != nil {
		return nil, err
	}
	pkg := &utils.PackageVersion{}
	if err = json.Unmarshal(data, pkg); err != nil {
		return nil, fmt.Errorf("unmarshal package info error: %v", err)
	}
	return pkg, nil
}

// RemotePackageColumns 非详细模式显示字段
type RemotePackageColumns struct {
	PackageName string `json:"packageName"`
	Version     string `json:"version"`
	Os          string `json:"os"`
	Arch        string `json:"arch"`
	Description string `json:"description"`
}

// RemotePackageColumnsVerbose 详细模式显示字段
type RemotePackageColumnsVerbose struct {
	PackageName string `json:"packageName"`
	Size        string `json:"size"`
	Checksum    string `json:"checksum"`
	Algo        string `json:"checksumAlgo"`
	Version     string `json:"version"`
	Build       string `json:"build"`
	Os          string `json:"os"`
	Arch        string `json:"arch"`
	Description string `json:"description"`
}

// listRemotePackages 显示远程包列表
func listRemotePackages() error {
	var dataList []*orderedmap.OrderedMap

	// 获取包列表
	u := utils.NewUpgrader("", utils.UpgradeConfig{
		BaseUrl: config.GetBaseURL() + "/costrict",
	}, nil)
	defer u.Close()

	packages, err := u.GetRemotePackages()
	if err != nil {
		return err
	}

	// 遍历所有包
	for _, pkg := range packages.Packages {
		ret, err := listRemotePackage(pkg)
		if err != nil {
			fmt.Printf("error: %v\n", err.Error())
		} else {
			dataList = append(dataList, ret...)
		}
	}

	utils.PrintFormat(dataList)
	return nil
}

// listRemotePackage 列出指定远程包的信息
func listRemotePackage(packageName string) ([]*orderedmap.OrderedMap, error) {
	u := utils.NewUpgrader(packageName, utils.UpgradeConfig{
		BaseUrl: config.GetBaseURL() + "/costrict",
	}, nil)
	defer u.Close()

	// 获取该软件包支持的所有平台
	pkg, err := u.GetRemotePlatforms()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote platforms: %v", err)
	}

	var dataList []*orderedmap.OrderedMap

	// 遍历所有支持的平台
	for _, platform := range pkg.Platforms {
		// 获取该平台的远程版本列表
		versList, err := u.GetRemoteVersions()
		if err != nil {
			fmt.Printf("Warning: failed to get remote versions for platform %s/%s: %v\n",
				platform.Os, platform.Arch, err)
			continue
		}

		// 遍历该平台的所有版本
		for _, ver := range versList.Versions {
			// 非详细模式：仅显示基本字段
			row := RemotePackageColumns{}
			row.PackageName = versList.PackageName
			row.Os = versList.Os
			row.Arch = versList.Arch
			row.Version = ver.VersionId.String()
			row.Description = "*"

			// 获取版本的详细元数据（仅获取description）
			if ver.InfoUrl != "" {
				pkgInfo, err := getPackageDetailInfo(u, ver.InfoUrl)
				if err == nil {
					row.Description = pkgInfo.Description
				}
			}

			recordMap, _ := utils.StructToOrderedMap(row)
			dataList = append(dataList, recordMap)
		}
	}

	return dataList, nil
}

func init() {
	componentCmd.AddCommand(listCmd)
	// 添加 server 标志
	listCmd.Flags().BoolVarP(&optServer, "server", "s", false, "Show all remote packages available for download")
}
