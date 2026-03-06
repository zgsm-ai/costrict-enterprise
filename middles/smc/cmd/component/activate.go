package component

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/utils"
)

func activateSelf(u *utils.Upgrader, newVer utils.VersionNumber) error {
	tmpFname := filepath.Join(u.BaseDir, "package", newVer.String(), "smc")
	targetFile := filepath.Join(u.BaseDir, "bin", "smc")
	if runtime.GOOS == "windows" {
		targetFile += ".exe"
		tmpFname += ".exe"
	}

	// 直接执行升级命令，不保存脚本文件
	fmt.Println("启动升级命令...")
	if runtime.GOOS == "windows" {
		// Windows: 使用 start 命令在新窗口中执行升级命令
		// 对路径进行引号包裹，防止空格和特殊字符导致的问题
		quotedTmpFname := fmt.Sprintf(`"%s"`, tmpFname)
		quotedTargetFile := fmt.Sprintf(`"%s"`, targetFile)

		// 构建更简单可靠的命令
		command := fmt.Sprintf("echo 升级smc... && timeout /t 3 /nobreak > nul && copy /Y %s %s || pause",
			quotedTmpFname, quotedTargetFile)
		//
		cmd := exec.Command("cmd", "/C", fmt.Sprintf("start /min cmd /C \"%s\"", command))
		SetNewPG(cmd)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("启动升级命令失败: %v", err)
		}
		fmt.Printf("command line: %s\n", cmd.String())
	} else {
		// Unix/Linux: 使用 nohup 在后台执行升级命令
		cmd := exec.Command("sh", "-c", fmt.Sprintf("nohup sh -c 'echo \"升级smc...\" && sleep 3 && cp -f \"%s\" \"%s\" && echo \"升级完成\"' > /dev/null 2>&1 &",
			tmpFname, targetFile))
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("启动升级命令失败: %v", err)
		}
	}

	fmt.Printf("smc正在后台升级到版本 %s\n", newVer.String())

	// 立即退出当前程序
	os.Exit(0)
	return nil // 这行代码不会执行，只是为了语法完整
}

func activatePackage() error {
	var err error
	if err = common.InitCommonEnv(); err != nil {
		return err
	}

	if optActivatePackageName == "" {
		fmt.Println("Error: package name is required (either as positional argument or via -p/--package option)")
		return fmt.Errorf("miss parameter")
	}
	if optActivatePackageVersion == "" {
		fmt.Println("Error: package version is required")
		return fmt.Errorf("miss parameter")
	}
	var ver utils.VersionNumber
	if err = ver.Parse(optActivatePackageVersion); err != nil {
		fmt.Printf("The version '%s' is invalid", optActivatePackageVersion)
		return err
	}
	u := utils.NewUpgrader(optActivatePackageName, utils.UpgradeConfig{
		BaseUrl: env.BaseUrl + "/costrict",
	}, nil)
	pkg, err := u.GetLocalVersion(&ver)
	if err != nil {
		fmt.Printf("The version '%s' is not exist\n", ver.String())
		return err
	}
	if err = u.ActivatePackage(pkg); err != nil {
		if optActivatePackageName == "smc" {
			// 当package选项未设置时，默认升级smc自身
			return activateSelf(u, ver)
		}
		fmt.Printf("The '%s-%s' activate failed: %v", optActivatePackageName, optActivatePackageVersion, err)
		return err
	}

	fmt.Printf("The '%s-%s' is activated successfully\n", optActivatePackageName, optActivatePackageVersion)
	return nil
}

var activateCmd = &cobra.Command{
	Use:   "activate {package-name | -p package-name}",
	Short: "Activate package",
	Long:  `Activate package`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optActivatePackageName = args[0]
		}
		return activatePackage()
	},
}

const activateExample = `  # activate package
  smc package activate -p codebase-syncer -v 1.2.0
  smc package activate codebase-syncer -v 1.2.0`

var optActivatePackageName string
var optActivatePackageVersion string

func init() {
	componentCmd.AddCommand(activateCmd)
	activateCmd.Flags().SortFlags = false
	activateCmd.Example = activateExample

	activateCmd.Flags().StringVarP(&optActivatePackageName, "package", "p", "", "package name")
	activateCmd.Flags().StringVarP(&optActivatePackageVersion, "version", "v", "", "special package version")
}
