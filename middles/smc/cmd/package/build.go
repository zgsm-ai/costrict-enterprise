package pkg

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/internal/utils"
)

/**
 *	Build package descriptor file for executable
 */
func makePackage() error {
	size, md5str, err := utils.CalcFileMd5(optFrom)
	if err != nil {
		return err
	}
	priKey, err := os.ReadFile(optKeyFile)
	if err != nil {
		return err
	}
	data, err := utils.Sign(priKey, []byte(md5str))
	if err != nil {
		return err
	}
	dir, fname := filepath.Split(optFrom)

	pkgData := &utils.PackageVersion{}
	pkgData.Arch = optArch
	pkgData.Os = optOs
	pkgData.PackageName = optPackage
	pkgData.PackageType = utils.PackageType(optType)
	pkgData.FileName = fname
	pkgData.Size = size
	pkgData.Checksum = md5str
	pkgData.ChecksumAlgo = "md5"
	pkgData.Sign = hex.EncodeToString(data)
	pkgData.Description = optDescription
	pkgData.ReleaseTime = time.Now().Local().Format("2006-01-02 15:04:05")
	err = pkgData.VersionId.Parse(optVersion)
	if err != nil {
		return fmt.Errorf("parse version error: %v", err)
	}
	if optFileName != "" {
		pkgData.FileName = optFileName
	}
	bytes, err := json.MarshalIndent(pkgData, "", "  ")
	if err != nil {
		return err
	}
	outputFname := optOutput
	if outputFname == "" {
		outputFname = filepath.Join(dir, "package.json")
	}
	return os.WriteFile(outputFname, bytes, 0666)
}

// packageBuildCmd represents the 'smc package build' command
var packageBuildCmd = &cobra.Command{
	Use:   "build {package | --package package}",
	Short: "Generate package descriptor file",
	Long:  `Signs a file and generates package.json`,
	Args:  cobra.MaximumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		// 如果有位置参数，使用第一个参数作为包名
		if len(args) > 0 {
			optPackage = args[0]
		}
		// 检查是否提供了包名（通过位置参数或选项）
		if optPackage == "" {
			fmt.Println("Error: package name is required (either as positional argument or via -p/--package option)")
			return
		}
		if err := makePackage(); err != nil {
			fmt.Println(err)
		}
	},
}

var optOs string
var optArch string
var optVersion string
var optOutput string
var optFrom string
var optKeyFile string
var optPackage string
var optType string
var optFileName string
var optDescription string

func init() {
	packageCmd.AddCommand(packageBuildCmd)

	packageBuildCmd.Example = `  # Sign shenma.exe with private key costrict-private.pem and generate package descriptor package-windows-amd64-1.0.1120.json (using package option)
  smc package build -p shenma -f ./shenma.exe -k costrict-private.pem -s windows -a amd64 -v 1.0.1120
  # Same command but using positional argument for package name
  smc package build shenma -f ./shenma.exe -k costrict-private.pem -s windows -a amd64 -v 1.0.1120`
	packageBuildCmd.Flags().SortFlags = false
	packageBuildCmd.Flags().StringVarP(&optPackage, "package", "p", "", "Package name")
	packageBuildCmd.Flags().StringVarP(&optFrom, "from", "f", "", "Package file to sign")
	packageBuildCmd.Flags().StringVarP(&optKeyFile, "key", "k", "", "Private key file")
	packageBuildCmd.Flags().StringVarP(&optOs, "os", "s", runtime.GOOS, "Target operating system")
	packageBuildCmd.Flags().StringVarP(&optArch, "arch", "a", runtime.GOARCH, "Target hardware architecture")
	packageBuildCmd.Flags().StringVarP(&optVersion, "version", "v", "1.0.0", "Package version number(semver)")
	packageBuildCmd.Flags().StringVarP(&optType, "type", "t", "exec", "Package type: exec/conf")
	packageBuildCmd.Flags().StringVarP(&optFileName, "filename", "", "", "File installation name/path")
	packageBuildCmd.Flags().StringVarP(&optDescription, "description", "d", "", "Package description")
	packageBuildCmd.Flags().StringVarP(&optOutput, "output", "o", "", "Output .json file")
	packageBuildCmd.MarkFlagRequired("from")
	packageBuildCmd.MarkFlagRequired("key")
}
