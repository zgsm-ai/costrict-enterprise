package component

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/internal/utils"
)

/**
 *	格式化文件大小
 */
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

/**
 *	获取包详细元数据信息
 */
func getPackageDetailInfo(u *utils.Upgrader, infoUrl string) (*utils.PackageVersion, error) {
	data, err := u.GetBytes(infoUrl, nil)
	if err != nil {
		return nil, err
	}
	pkg := &utils.PackageVersion{}
	if err = json.Unmarshal(data, pkg); err != nil {
		return nil, fmt.Errorf("unmarshal package info error: %v", err)
	}
	return pkg, nil
}

/**
 *	Fields displayed in list format (non-verbose mode)
 */
type RemotePackage_Columns struct {
	PackageName string `json:"packageName"`
	Version     string `json:"version"`
	Os          string `json:"os"`
	Arch        string `json:"arch"`
	Size        string `json:"size"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

/**
 *	Fields displayed in list format (verbose mode)
 */
type RemotePackage_Columns_Verbose struct {
	PackageName string `json:"packageName"`
	Version     string `json:"version"`
	Os          string `json:"os"`
	Arch        string `json:"arch"`
	Size        string `json:"size"`
	Date        string `json:"date"`
	Checksum    string `json:"checksum"`
	Algo        string `json:"checksumAlgo"`
	Build       string `json:"build"`
	Description string `json:"description"`
}

func listPackages(verbose bool) error {
	// 格式化输出版本列表
	var dataList []*orderedmap.OrderedMap
	// 获取包列表以检查Details中是否存在该包
	u := utils.NewUpgrader("", utils.UpgradeConfig{
		BaseUrl: getUpgradeUrl(),
	}, nil)

	packages, err := u.GetRemotePackages()
	if err != nil {
		return err
	}
	if optRemotePackageName != "" {
		ret, err := listPackage(optRemotePackageName, verbose)
		if err != nil {
			return err
		}
		dataList = append(dataList, ret...)
	} else {
		for _, pkg := range packages.Packages {
			ret, err := listPackage(pkg, verbose)
			if err != nil {
				fmt.Printf("error: %v\n", err.Error())
			} else {
				dataList = append(dataList, ret...)
			}
		}
	}
	utils.PrintFormat(dataList)
	return nil
}

func getSimplify(d RemotePackage_Columns_Verbose) RemotePackage_Columns {
	return RemotePackage_Columns{
		PackageName: d.PackageName,
		Version:     d.Version,
		Os:          d.Os,
		Arch:        d.Arch,
		Size:        d.Size,
		Date:        d.Date,
		Description: d.Description,
	}
}

/**
 *	List remote package information
 */
func listPackage(packageName string, verbose bool) ([]*orderedmap.OrderedMap, error) {
	u := utils.NewUpgrader(packageName, utils.UpgradeConfig{
		BaseUrl: getUpgradeUrl(),
	}, nil)
	// 获取该软件包支持的所有平台
	pkg, err := u.GetRemotePlatforms()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote platforms: %v", err)
	}
	// 格式化输出版本列表
	var dataList []*orderedmap.OrderedMap
	// 遍历所有支持的平台，根据 os 和 arch 参数进行过滤
	for _, platform := range pkg.Platforms {
		if optRemoteOs != "" && platform.Os != optRemoteOs { // 如果只指定了 os，只显示匹配 os 的平台
			continue
		}
		if optRemoteArch != "" && platform.Arch != optRemoteArch { // 如果只指定了 arch，只显示匹配 arch 的平台
			continue
		}
		var pov *utils.PlatformOverview
		if pkg.Overviews != nil {
			ov, exists := pkg.Overviews[fmt.Sprintf("%s-%s", platform.Os, platform.Arch)]
			if exists {
				pov = &ov
			}
		}
		var platData []RemotePackage_Columns_Verbose
		if pov != nil {
			platData, err = getPlatform(packageName, pov, verbose)
		} else {
			// 调用 listPlatform 函数搜集单个平台信息
			platData, err = listPlatform(packageName, platform.Os, platform.Arch, verbose)
		}
		if err != nil {
			fmt.Printf("Warning: failed to get platform data for %s/%s: %v\n", platform.Os, platform.Arch, err)
			continue
		}
		for _, d := range platData {
			var row *orderedmap.OrderedMap
			if verbose {
				row, _ = utils.StructToOrderedMap(d)
			} else {
				row, _ = utils.StructToOrderedMap(getSimplify(d))
			}
			dataList = append(dataList, row)
		}
	}
	return dataList, nil
}

/**
 *	从包详细信息构建OrderedMap数据
 */
func getPlatform(packageName string, pov *utils.PlatformOverview, verbose bool) ([]RemotePackage_Columns_Verbose, error) {
	var dataList []RemotePackage_Columns_Verbose

	// 遍历该平台的所有版本
	for _, version := range pov.Versions {
		row := RemotePackage_Columns_Verbose{}
		row.PackageName = packageName
		row.Os = pov.Os
		row.Arch = pov.Arch
		row.Version = version.VersionId.String()
		row.Size = formatSize(version.Size)
		row.Checksum = "*"
		row.Algo = "*"
		row.Build = version.Build
		row.Date = version.ReleaseTime
		row.Description = version.Description
		dataList = append(dataList, row)
	}
	return dataList, nil
}

/**
 *	搜集单个平台信息
 */
func listPlatform(packageName, os, arch string, verbose bool) ([]RemotePackage_Columns_Verbose, error) {
	// 为平台创建特定的配置
	u := utils.NewUpgrader(packageName, utils.UpgradeConfig{
		Os:      os,
		Arch:    arch,
		BaseUrl: getUpgradeUrl(),
	}, nil)

	// 获取该平台的远程版本列表
	versList, err := u.GetRemoteVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote versions for platform %s/%s: %v", os, arch, err)
	}

	// 格式化输出版本列表
	var dataList []RemotePackage_Columns_Verbose

	// 遍历该平台的所有版本
	for _, ver := range versList.Versions {
		// verbose模式：显示所有字段
		row := RemotePackage_Columns_Verbose{}
		row.PackageName = versList.PackageName
		row.Os = versList.Os
		row.Arch = versList.Arch
		row.Version = ver.VersionId.String()
		row.Size = "*"
		row.Checksum = "*"
		row.Algo = "*"
		row.Build = "*"
		row.Description = "*"
		// 获取版本的详细元数据
		if ver.InfoUrl != "" {
			pkgInfo, err := getPackageDetailInfo(u, u.BaseUrl+ver.InfoUrl)
			if err == nil {
				row.Size = formatSize(pkgInfo.Size)
				row.Checksum = pkgInfo.Checksum
				row.Algo = pkgInfo.ChecksumAlgo
				row.Build = pkgInfo.Build
				row.Date = pkgInfo.ReleaseTime
				row.Description = pkgInfo.Description
			}
		}
		dataList = append(dataList, row)
	}
	return dataList, nil
}

// remoteCmd represents the 'smc package remote' command
var remoteCmd = &cobra.Command{
	Use:   "remote {package | -p package} [--os os] [--arch arch]",
	Short: "List remote packages",
	Long:  `Lists remote packages available for download`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optRemotePackageName = args[0]
		}
		return listPackages(optRemoteVerbose)
	},
}

const packageRemoteExample = `  # List all remote packages
  smc package remote
  # List specific remote package
  smc package remote aip
  # Show remote package details
  smc package remote -p aip -v
  # List packages for specific OS and architecture
  smc package remote -p aip --os linux --arch amd64`

var optRemotePackageName string
var optRemoteVerbose bool
var optRemoteOs string
var optRemoteArch string

func init() {
	componentCmd.AddCommand(remoteCmd)
	remoteCmd.Flags().SortFlags = false
	remoteCmd.Example = packageRemoteExample
	remoteCmd.Flags().StringVarP(&optRemotePackageName, "package", "p", "", "Package name")
	remoteCmd.Flags().BoolVarP(&optRemoteVerbose, "verbose", "v", false, "Show details")
	remoteCmd.Flags().StringVar(&optRemoteOs, "os", "", "Target operating system")
	remoteCmd.Flags().StringVar(&optRemoteArch, "arch", "", "Target architecture")
}
