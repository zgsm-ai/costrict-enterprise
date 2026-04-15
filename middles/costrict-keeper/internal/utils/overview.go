package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

type VersionOverview struct {
	VersionId   VersionNumber `json:"versionId"`   //版本号，采用SemVer标准
	PackageType PackageType   `json:"packageType"` //包类型: exec/conf
	FileName    string        `json:"fileName"`    //被打包的文件的名字
	Size        uint64        `json:"size"`        //包文件大小
	Build       string        `json:"build"`       //构建信息：Tag/Branch信息 CommitID BuildTime
	Description string        `json:"description"` //版本描述，含有更丰富的可读信息
}

type PlatformOverview struct {
	Os       string            `json:"os"`
	Arch     string            `json:"arch"`
	Newest   VersionOverview   `json:"newest"`
	Versions []VersionOverview `json:"versions"`
}

/**
 *	平台标识
 */
type PlatformId struct {
	Os   string `json:"os"`
	Arch string `json:"arch"`
}

/**
 *	包目录（软件包的系统，平台，版本目录）
 */
type PackageOverview struct {
	PackageName string                      `json:"packageName"` //包名称
	Platforms   []PlatformId                `json:"platforms"`   //包支持的平台列表
	Overviews   map[string]PlatformOverview `json:"overviews"`   //包总览
}

/**
 *	云端可供下载的包列表
 */
type PackageList struct {
	Packages []string `json:"packages"`
}

func (u *Upgrader) GetRemotePlatforms() (PackageOverview, error) {
	//	<base-url>/<package>/platforms.json
	urlStr := fmt.Sprintf("%s/%s/platforms.json", u.BaseUrl, u.packageName)

	bytes, err := u.GetBytes(urlStr, nil)
	if err != nil {
		return PackageOverview{}, err
	}
	plats := &PackageOverview{}
	if err = json.Unmarshal(bytes, plats); err != nil {
		return *plats, fmt.Errorf("GetRemotePlatforms('%s') unmarshal error: %v", urlStr, err)
	}
	return *plats, nil
}

func (u *Upgrader) GetRemotePackages() (PackageList, error) {
	//	<base-url>/packages.json
	urlStr := fmt.Sprintf("%s/packages.json", u.BaseUrl)

	bytes, err := u.GetBytes(urlStr, nil)
	if err != nil {
		return PackageList{}, err
	}
	pkgs := &PackageList{}
	if err = json.Unmarshal(bytes, pkgs); err != nil {
		return *pkgs, fmt.Errorf("GetRemotePackages('%s') unmarshal error: %v", urlStr, err)
	}
	return *pkgs, nil
}

func (u *Upgrader) checkExistPackage(cacheFname string, pkg *PackageVersion) error {
	if _, err := os.Stat(cacheFname); err != nil {
		return err
	}

	if err := u.verifyIntegrity(*pkg, cacheFname); err != nil {
		return err
	}
	return nil
}

/**
 *	SyncPackage 将远程包目录树以镜像的方式同步到本地目录
 *	下载层次结构：
 *	- dstDir/platforms.json ← <base-url>/<package>/platforms.json
 *	- 对于每个平台组合 (os/arch)：
 *	  - dstDir/<os>/<arch>/platform.json ← <base-url>/<package>/<os>/<arch>/platform.json
 *	  - 对于每个版本：
 *	    - dstDir/<os>/<arch>/<version>/package.json ← <base-url>/<package>/<os>/<arch>/<version>/package.json
 *	    - dstDir/<os>/<arch>/<version>/ ← <base-url>/<package>/<os>/<arch>/<version>/<filename>
 */
func (u *Upgrader) SyncPackage(dstDir string) error {
	// 1. 下载 platforms.json
	platformsUrl := fmt.Sprintf("%s/%s/platforms.json", u.BaseUrl, u.packageName)
	platformsPath := filepath.Join(dstDir, "platforms.json")
	if err := u.GetFile(platformsUrl, nil, platformsPath); err != nil {
		return fmt.Errorf("下载 platforms.json 失败: %w", err)
	}

	// 2. 读取 platforms.json 获取平台列表
	bytes, err := os.ReadFile(platformsPath)
	if err != nil {
		return fmt.Errorf("读取 platforms.json 失败: %w", err)
	}

	var packageOverview PackageOverview
	if err := json.Unmarshal(bytes, &packageOverview); err != nil {
		return fmt.Errorf("解析 platforms.json 失败: %w", err)
	}

	// 3. 遍历每个平台组合
	var lastErr error
	for _, platformId := range packageOverview.Platforms {
		if err := u.syncPlatform(dstDir, platformId); err != nil {
			log.Printf("Sync %s-%s/%s to %s error: %v", u.packageName, platformId.Os, platformId.Arch, dstDir, err)
			lastErr = err
			continue
		}
	}

	return lastErr
}

/**
 * Sync platform versions and update platform info to keep only last 3 versions
 * @param {string} dstDir - Destination directory for platform packages
 * @param {PlatformId} pi - Platform identifier (OS and Architecture)
 * @returns {error} Returns error if sync fails, nil on success
 * @description
 * - Downloads platform.json from remote server
 * - Sorts version list by version number (ascending)
 * - Keeps only the last 3 (newest) versions in platformInfo
 * - Saves updated platformInfo back to file
 * - Downloads package files for the last 3 versions only
 * - Skips versions that already exist locally
 * @throws
 * - Directory creation error (os.MkdirAll)
 * - JSON file download error (GetFile)
 * - JSON unmarshal error (json.Unmarshal)
 * - Version sync error (syncVersion)
 * - JSON marshal error (json.Marshal)
 * - File write error (os.WriteFile)
 * @example
 * err := upgrader.syncPlatform("/cache/packages", PlatformId{Os: "linux", Arch: "amd64"})
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (u *Upgrader) syncPlatform(dstDir string, pi PlatformId) error {
	platformDir := filepath.Join(dstDir, pi.Os, pi.Arch)
	if err := os.MkdirAll(platformDir, 0755); err != nil {
		return err
	}

	// 下载 platform.json
	platformUrl := fmt.Sprintf("%s/%s/%s/%s/platform.json", u.BaseUrl, u.packageName, pi.Os, pi.Arch)
	platformJsonPath := filepath.Join(platformDir, "platform.json")
	platformBytes, err := u.GetBytes(platformUrl, nil)
	if err != nil {
		return err
	}

	var platformInfo PlatformInfo
	if err := json.Unmarshal(platformBytes, &platformInfo); err != nil {
		return err
	}

	// 对版本列表排序，只保留最后三个版本
	sort.Slice(platformInfo.Versions, func(i, j int) bool {
		return CompareVersion(platformInfo.Versions[i].VersionId, platformInfo.Versions[j].VersionId) < 0
	})
	versionsCount := len(platformInfo.Versions)
	if versionsCount > 3 {
		platformInfo.Versions = platformInfo.Versions[versionsCount-3:]
	}

	// 更新 platform.json 文件，只保留最后三个版本版本信息
	updatedBytes, err := json.MarshalIndent(platformInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal platform info error: %v", err)
	}
	if err := os.WriteFile(platformJsonPath, updatedBytes, 0644); err != nil {
		return fmt.Errorf("write platform info error: %v", err)
	}

	var lastErr error
	for _, versionAddr := range platformInfo.Versions {
		verDir := filepath.Join(platformDir, versionAddr.VersionId.String())
		if err := u.syncVersion(verDir, versionAddr); err != nil {
			log.Printf("Sync %s-%s to %s error: %v", u.packageName, versionAddr.VersionId.String(), verDir, err)
			lastErr = err
			continue
		}
	}
	u.cleanupOlderRepos(&platformInfo, platformDir)
	return lastErr
}

func (u *Upgrader) syncVersion(verDir string, verAddr VersionAddr) error {
	if err := os.MkdirAll(verDir, 0755); err != nil {
		return err
	}
	pkgJsonUrl := u.BaseUrl + verAddr.InfoUrl
	pkgJsonPath := filepath.Join(verDir, "package.json")
	if err := u.GetFile(pkgJsonUrl, nil, pkgJsonPath); err != nil {
		return err
	}

	var pkg PackageVersion
	if err := pkg.Load(pkgJsonPath); err != nil {
		return err
	}
	_, fname := filepath.Split(pkg.FileName)
	cacheFname := filepath.Join(verDir, fname)
	if err := u.checkExistPackage(cacheFname, &pkg); err == nil {
		return nil
	}
	if err := u.GetFile(u.BaseUrl+verAddr.AppUrl, nil, cacheFname); err != nil {
		return err
	}
	return nil
}

/**
 * Clean up older version directories that are not in platform info
 * @param {*PlatformInfo} plat - Platform information containing current versions list
 * @param {string} platDir - Platform directory path to clean up
 * @returns {error} Returns error if directory operations fail, nil on success
 * @description
 * - Lists all subdirectories in platform directory
 * - Parses each subdirectory name as version number
 * - Compares with versions in platform info
 * - Removes directories whose version is not in the active versions list
 * @throws
 * - Directory read error (os.ReadDir)
 * - Version parse error (VersionNumber.Parse)
 * - Directory removal error (os.RemoveAll)
 * @example
 * err := upgrader.cleanupOlderRepos(&platformInfo, "/cache/packages/linux/amd64")
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (u *Upgrader) cleanupOlderRepos(plat *PlatformInfo, platDir string) error {
	// Create a set of valid versions for quick lookup
	validVersions := make(map[string]bool)
	for _, ver := range plat.Versions {
		validVersions[ver.VersionId.String()] = true
	}

	// Read all entries in platform directory
	entries, err := os.ReadDir(platDir)
	if err != nil {
		return fmt.Errorf("read platform directory '%s' error: %v", platDir, err)
	}

	// Check each subdirectory
	for _, entry := range entries {
		// Skip non-directories and platform.json file
		if !entry.IsDir() || entry.Name() == "platform.json" {
			continue
		}

		// Parse directory name as version number
		var verNum VersionNumber
		if err := verNum.Parse(entry.Name()); err != nil {
			// Skip directories that are not valid version numbers
			continue
		}

		// Check if version is in the valid versions list
		if !validVersions[verNum.String()] {
			// Remove the version directory
			verDir := filepath.Join(platDir, entry.Name())
			if err := os.RemoveAll(verDir); err != nil {
				log.Printf("Remove directory '%s' error: %v", verDir, err)
				continue
			}
			log.Printf("Cleaned up older version directory: %s", verDir)
		}
	}

	return nil
}
