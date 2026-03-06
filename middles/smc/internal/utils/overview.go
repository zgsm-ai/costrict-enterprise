package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}
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

func (u *Upgrader) syncPlatform(dstDir string, pi PlatformId) error {
	platformDir := filepath.Join(dstDir, pi.Os, pi.Arch)
	if err := os.MkdirAll(platformDir, 0755); err != nil {
		return err
	}

	platformUrl := fmt.Sprintf("%s/%s/%s/%s/platform.json", u.BaseUrl, u.packageName, pi.Os, pi.Arch)
	platformJsonPath := filepath.Join(platformDir, "platform.json")
	if err := u.GetFile(platformUrl, nil, platformJsonPath); err != nil {
		return err
	}

	platformBytes, err := os.ReadFile(platformJsonPath)
	if err != nil {
		return err
	}

	var platformInfo PlatformInfo
	if err := json.Unmarshal(platformBytes, &platformInfo); err != nil {
		return err
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
