package component

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/utils"
)

/**
 *	Fields displayed in list format
 */
type Package_Columns struct {
	A           string `json:"A"`
	PackageName string `json:"packageName"`
	Size        string `json:"size"`
	Checksum    string `json:"checksum"`
	Algo        string `json:"algo"`
	Version     string `json:"version"`
	Os          string `json:"os"`
	Arch        string `json:"arch"`
	Description string `json:"description"`
}

type PackageInfo struct {
	Ver       *utils.PackageVersion
	Activated bool
}

func addPackage(pkg *utils.PackageVersion, pkgs map[string]*PackageInfo) error {
	key := pkg.PackageName + pkg.VersionId.String()

	if k, exists := pkgs[key]; exists {
		if !k.Activated {
			return fmt.Errorf("package %s already exists", pkg.PackageName)
		}
	} else {
		pkgs[key] = &PackageInfo{Ver: pkg, Activated: false}
	}
	return nil
}

func scanAllPackages(packageDir string, pkgs map[string]*PackageInfo) error {
	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		return err
	}
	scanVersions(packageDir, pkgs)
	for _, v := range pkgs {
		v.Activated = true
	}

	cachesDir := filepath.Join(packageDir, "caches")
	err := filepath.WalkDir(cachesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path == cachesDir {
			return nil
		}
		versDir := filepath.Join(cachesDir, d.Name())
		scanVersions(versDir, pkgs)
		return filepath.SkipDir
	})

	return err
}

func scanVersions(packageDir string, pkgs map[string]*PackageInfo) error {
	// 检查目录是否存在
	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		return err
	}

	// 遍历目录中的 *.json 文件
	filepath.WalkDir(packageDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != packageDir {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		// 只处理 *.json 文件
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file '%s': %v", path, err)
		}
		var pkgInfo utils.PackageVersion
		if err := json.Unmarshal(data, &pkgInfo); err != nil {
			return fmt.Errorf("failed to unmarshal package info from '%s': %v", path, err)
		}
		if pkgInfo.PackageName == "" {
			return nil
		}
		addPackage(&pkgInfo, pkgs)
		return nil
	})
	return nil
}

/**
 *	List package information
 */
func packageList(packageName string, verbose bool) error {
	// 获取 .costrict/package 目录路径
	u := utils.NewUpgrader(packageName, utils.UpgradeConfig{
		BaseUrl: env.BaseUrl + "/costrict",
	}, nil)
	packageDir := filepath.Join(u.BaseDir, "package")

	// 扫描目录并收集包信息
	pkgs := make(map[string]*PackageInfo)
	err := scanAllPackages(packageDir, pkgs)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	packageInfos := make([]PackageInfo, 0, len(pkgs))
	for _, pkg := range pkgs {
		if packageName != "" && pkg.Ver.PackageName != packageName {
			continue
		}
		packageInfos = append(packageInfos, *pkg)
	}

	// 排序
	sort.Slice(packageInfos, func(i, j int) bool {
		if packageInfos[i].Ver.PackageName == packageInfos[j].Ver.PackageName {
			return utils.CompareVersion(packageInfos[i].Ver.VersionId, packageInfos[j].Ver.VersionId) < 0
		}
		return packageInfos[i].Ver.PackageName < packageInfos[j].Ver.PackageName
	})

	// 如果指定了包名且只有一个包，显示详细信息
	if len(pkgs) == 1 && verbose {
		utils.PrintYaml(packageInfos[0].Ver)
		return nil
	}

	// 格式化输出包列表
	var dataList []*orderedmap.OrderedMap
	for _, p := range packageInfos {
		pkg := p.Ver
		row := Package_Columns{}
		row.PackageName = pkg.PackageName
		row.Os = pkg.Os
		row.Arch = pkg.Arch
		row.Size = fmt.Sprintf("%d", pkg.Size)
		row.Checksum = pkg.Checksum
		row.Algo = pkg.ChecksumAlgo
		row.Version = fmt.Sprintf("%d.%d.%d", pkg.VersionId.Major, pkg.VersionId.Minor, pkg.VersionId.Micro)
		row.Description = pkg.Description
		if p.Activated {
			row.A = "*"
		} else {
			row.A = " "
		}

		recordMap, _ := utils.StructToOrderedMap(row)
		dataList = append(dataList, recordMap)
	}

	utils.PrintFormat(dataList)
	return nil
}

// packageListCmd represents the 'smc package list' command
var packageListCmd = &cobra.Command{
	Use:   "list {package | -p package}",
	Short: "List available packages",
	Long:  `Lists all available packages in the .costrict/package directory`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optListPackageName = args[0]
		}
		return packageList(optListPackageName, optListVerbose)
	},
}

const packageListExample = `  # List all packages
  smc package list
  # List specific package
  smc package list aip
  # Show package details
  smc package list -p aip -v
  # List remote versions
  smc package remote aip`

var optListPackageName string
var optListVerbose bool

func init() {
	componentCmd.AddCommand(packageListCmd)
	packageListCmd.Flags().SortFlags = false
	packageListCmd.Example = packageListExample
	packageListCmd.Flags().StringVarP(&optListPackageName, "package", "p", "", "Package name")
	packageListCmd.Flags().BoolVarP(&optListVerbose, "verbose", "v", false, "Show details")
}
