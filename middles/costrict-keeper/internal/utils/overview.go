package utils

import (
	"encoding/json"
	"fmt"
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
