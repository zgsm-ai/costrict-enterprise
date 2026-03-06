package component

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/utils"
)

func upgradePackage() error {
	var err error
	var publicKey string
	if optPublicKey != "" {
		// 读取公钥文件内容
		keyData, err := os.ReadFile(optPublicKey)
		if err != nil {
			return fmt.Errorf("failed to read public key file: %v", err)
		}
		publicKey = string(keyData)
	}
	u := utils.NewUpgrader(optUpgradePackageName, utils.UpgradeConfig{
		BaseUrl:   env.BaseUrl + "/costrict",
		PublicKey: publicKey,
	}, nil)

	var newVer *utils.VersionNumber
	if optUpgradeVersion != "" {
		var ver utils.VersionNumber
		err = ver.Parse(optUpgradeVersion)
		if err != nil {
			return err
		}
		newVer = &ver
	} else {
		newVer = nil
	}

	pkg, upgraded, err := u.GetPackage(newVer)
	if err != nil {
		fmt.Printf("The '%s' upgrade failed: %v", optUpgradePackageName, err)
		return err
	}
	if !upgraded {
		fmt.Printf("The '%s' version '%s' is up to date\n",
			optUpgradePackageName, pkg.VersionId.String())
		return nil
	}
	if err := u.ActivatePackage(pkg); err != nil {
		if optUpgradePackageName == "smc" {
			// 当package选项未设置时，默认升级smc自身
			return activateSelf(u, pkg.VersionId)
		}
		fmt.Printf("The '%s' activate '%s' failed: %v",
			optUpgradePackageName, pkg.VersionId.String(), err)
		return err
	}
	fmt.Printf("The '%s' is upgraded to version %s\n", optUpgradePackageName, pkg.VersionId.String())
	return nil
}

// SetNewPG 设置进程属性，使子进程在父进程退出后继续运行
// Windows系统实现
func SetNewPG(cmd *exec.Cmd) {
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	// }
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade {package-name | -p package-name}",
	Short: "Upgrade package",
	Long:  `Upgrade package`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optUpgradePackageName = args[0]
		}
		if err := common.InitCommonEnv(); err != nil {
			return err
		}
		if optUpgradePackageName == "" {
			optUpgradePackageName = "smc"
		}
		return upgradePackage()
	},
}

const upgradeExample = `  # upgrade package
  smc component upgrade -p codebase-syncer -v 1.0.0
  smc component upgrade -p codebase-syncer
  smc component upgrade -p codebase-syncer --public /path/to/public.key
  # upgrade smc itself
  smc component upgrade
  smc component upgrade --self`

var optUpgradeVersion string
var optUpgradePackageName string
var optPublicKey string

func init() {
	componentCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().SortFlags = false
	upgradeCmd.Example = upgradeExample

	upgradeCmd.Flags().StringVarP(&optUpgradePackageName, "package", "p", "", "package name")
	upgradeCmd.Flags().StringVarP(&optUpgradeVersion, "version", "v", "", "package version")
	upgradeCmd.Flags().StringVar(&optPublicKey, "public", "", "public key file for package verification")
}
