package config

import (
	"completion-agent/pkg/utils"
	"fmt"
	"log"
)

/**
 * 获取远程配置
 * @param {string} pkgName - 要获取的配置包名称，如"completion-config"
 * @returns {error} 返回获取过程中的错误，成功返回nil
 * @description
 * - 创建升级器实例用于下载远程配置
 * - 设置基础URL和基础目录
 * - 调用升级器升级包到最新版本
 * - 记录升级结果和版本信息
 * - 用于从远程服务器获取最新的配置文件
 * @throws
 * - 包升级失败时返回错误
 * - 网络连接失败时返回错误
 * @example
 * err := fetchRemoteConfig("completion-config")
 * if err != nil {
 *     log.Printf("获取配置失败: %v", err)
 * }
 */
func fetchRemoteConfig(pkgName string) error {
	u := utils.NewUpgrader(pkgName, utils.UpgradeConfig{
		BaseUrl: fmt.Sprintf("%s/costrict", GetBaseURL()),
		BaseDir: getCostrictDir(),
	})

	pkg, upgraded, err := u.UpgradePackage(nil)
	if err != nil {
		log.Printf("fetch config failed: %v", err)
		return err
	}
	if !upgraded {
		log.Printf("The '%s' version is up to date\n", pkg.PackageName)
	} else {
		log.Printf("The '%s' is upgraded to version %s\n", pkg.PackageName, pkg.VersionId.String())
	}
	return nil
}

func UpdateRemoteConfigs() error {
	var lasterr error
	if err := fetchRemoteConfig("completion-config"); err != nil {
		log.Printf("Fetch failed: %v", err)
		lasterr = err
	}
	if err := fetchRemoteConfig("tokenizer-config"); err != nil {
		log.Printf("Fetch failed: %v", err)
		lasterr = err
	}
	if err := fetchRemoteConfig("hidden-scores-config"); err != nil {
		log.Printf("Fetch failed: %v", err)
		lasterr = err
	}
	return lasterr
}
