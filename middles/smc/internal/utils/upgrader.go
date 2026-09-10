package utils

import (
	"archive/zip"
	"bufio"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

/**
 *	包类型枚举
 */
type PackageType string

const (
	PackageTypeExec PackageType = "exec"
	PackageTypeConf PackageType = "conf"
	PackageTypeZip  PackageType = "zip"
)

/**
 *	版本编号
 */
type VersionNumber struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Micro int `json:"micro"`
}

/**
 *	包版本的描述&签名信息，用于验证包的正确性
 */
type PackageVersion struct {
	PackageName  string        `json:"packageName"`           //包名字
	PackageType  PackageType   `json:"packageType"`           //包类型: exec/conf
	FileName     string        `json:"fileName"`              //被打包的文件的相对路径(相对.costrict目录,为空则安装到默认路径)
	Os           string        `json:"os"`                    //操作系统名:linux/windows
	Arch         string        `json:"arch"`                  //硬件架构
	Size         uint64        `json:"size"`                  //包文件大小
	Checksum     string        `json:"checksum"`              //Md5散列值
	Sign         string        `json:"sign"`                  //签名，使用私钥签的名，需要用对应公钥验证
	ChecksumAlgo string        `json:"checksumAlgo"`          //固定为“md5”
	VersionId    VersionNumber `json:"versionId"`             //版本号，采用SemVer标准
	Build        string        `json:"build"`                 //构建信息：Tag/Branch信息 CommitID BuildTime
	ReleaseTime  string        `json:"releaseTime,omitempty"` //包发布时间
	Description  string        `json:"description"`           //版本描述，含有更丰富的可读信息
}

/**
 *	一个package版本的地址信息
 */
type VersionAddr struct {
	VersionId VersionNumber `json:"versionId"` //版本的地址信息
	AppUrl    string        `json:"appUrl"`    //包地址
	InfoUrl   string        `json:"infoUrl"`   //包描述信息(PackageVersion)文件的地址
}

/**
 *	指定平台的关键信息，比如，最新版本，版本列表（描述一个硬件平台/操作系统对应的包列表）
 */
type PlatformInfo struct {
	PackageName string        `json:"packageName"`
	Os          string        `json:"os"`
	Arch        string        `json:"arch"`
	Newest      VersionAddr   `json:"newest"`
	Versions    []VersionAddr `json:"versions"`
}

type UpgradeConfig struct {
	PublicKey  string //用来验证包签名的公钥
	BaseUrl    string //保存安装包的服务器的基地址
	BaseDir    string //costrict数据所在的基路径
	Os         string //操作系统名
	Arch       string //硬件平台名
	TargetPath string //指定安装目标路径(及文件名)
	NoSetPath  bool   //不需要设置PATH。设置PATH可以让程序所在路径被自动搜索
	VerifyTLS  bool   //进行TLS/SSL相关的安全校验
}

type Upgrader struct {
	UpgradeConfig

	packageName string       //包名称
	installDir  string       //包文件的实际安装目录
	packageDir  string       //包信息目录
	client      *http.Client //HTTP客户端
}

// const SHENMA_PUBLIC_KEY = `-----BEGIN PUBLIC KEY-----
// MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwClPrRPGCOXcWPFMPIPc
// Hn5angPRwuIvwSGle/O7VaZfaTuplMVa2wUPzWv1AfmKpENMm0pf0uhnTyfH3gnR
// C46rNeMmBcLg8Jd7wTWXtik0IN7CREOQ6obIiMY4Sbx25EPHPf8SeqvPpFq8uOEM
// YqRUQbPaY5+mgkDZMy68hJDUUstapBQovjSlnLXjG2pULWKIJF2g0gGWvS4LGznP
// Uvrq2U1QVpsja3EtoLq8jF3UcLJWVZt2pMd5H9m3ULBKFzpu7ix+wb3ebRr6JtUI
// bMzLAZ0BM0wxlpDmp1GYVag+Ll3w2o3LXLEB08soABD0wdD03Sp7flkbebgAxd1b
// vwIDAQAB
// -----END PUBLIC KEY-----`

const SHENMA_PUBLIC_KEY = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp/yvHEtGy09fNgZO2a/e
oyjEvBqVEjNf9RRf8r5QLeXI/InJGS323faqrVAtEjbOhq1R0KuAYISyFRzPvJYa
aBdlaDpXOY0UJxz6C/hLSAl2ohn/SvCycYVucrjnPUAwCqDNaLLjyqyTdsSXNh3d
QHgyBM16LD8oqFHj+/dxlMNxv+FIcc6WeN9F7BmTmvbHt5jBqBxBhXtlR8lx7F/H
AIMDOcw+6STgS2RFFnTRrBl8ZgJPBUavczm0TY4a9gUErfTnb8zBHtH6K4OPsvEF
Nimo+oDprwaVnIIPm1UvZtc/Qe/6OD0emoVovSzRYhbaqVPWgKqPNiitW9JZvuV3
nwIDAQAB
-----END PUBLIC KEY-----`

const SHENMA_BASE_URL = "https://zgsm.sangfor.com/costrict"

// ------------------------------------------------------------------------------
//
//	VersionNumber
//
// ------------------------------------------------------------------------------
func NewVersion(verStr string) (v VersionNumber, err error) {
	err = v.Parse(verStr)
	return
}

func (ver *VersionNumber) String() string {
	return fmt.Sprintf("%d.%d.%d", ver.Major, ver.Minor, ver.Micro)
}

func (ver *VersionNumber) Parse(verstr string) error {
	var err error
	var major, minor, micro int

	if verstr == "" {
		ver.Major = 0
		ver.Minor = 0
		ver.Micro = 0
		return nil
	}
	vers := strings.Split(verstr, ".")
	if len(vers) != 3 {
		return fmt.Errorf("invalid version string")
	}
	major, err = strconv.Atoi(vers[0])
	if err != nil {
		return err
	}
	minor, err = strconv.Atoi(vers[1])
	if err != nil {
		return err
	}
	micro, err = strconv.Atoi(vers[2])
	if err != nil {
		return err
	}
	ver.Major = major
	ver.Minor = minor
	ver.Micro = micro
	return nil
}

/**
 *	比较版本
 */
func CompareVersion(local, remote VersionNumber) int {
	if local.Major != remote.Major {
		return local.Major - remote.Major
	}
	if local.Minor != remote.Minor {
		return local.Minor - remote.Minor
	}
	return local.Micro - remote.Micro
}

//------------------------------------------------------------------------------
//	PackageVersion
//------------------------------------------------------------------------------

func (pkg *PackageVersion) Verify() error {
	if pkg.PackageType != "exec" && pkg.PackageType != "conf" && pkg.PackageType != "zip" {
		return fmt.Errorf("invalid package type: %s", pkg.PackageType)
	}
	if pkg.FileName == "" {
		return fmt.Errorf("invalid FileName: %s", pkg.FileName)
	}
	if filepath.IsAbs(pkg.FileName) {
		return fmt.Errorf("invalid FileName: %s", pkg.FileName)
	}
	return nil
}

func (pkg *PackageVersion) Load(fname string) error {
	bytes, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, pkg); err != nil {
		return err
	}
	return nil
}

func (pkg *PackageVersion) Save(fname string) error {
	bytes, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(fname, bytes, 0644); err != nil {
		log.Printf("Save package file '%s' failed: %v\n", fname, err)
		return err
	}
	return nil
}

//------------------------------------------------------------------------------
//	Upgrader
//------------------------------------------------------------------------------

func NewUpgrader(packageName string, cfg UpgradeConfig, client *http.Client) *Upgrader {
	u := &Upgrader{}
	u.UpgradeConfig = cfg
	u.packageName = packageName
	u.correct()
	u.client = client
	return u
}

func (u *Upgrader) getHttpClient() *http.Client {
	if u.client == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !u.VerifyTLS},
			MaxIdleConns:    5,
			IdleConnTimeout: 30,
		}
		u.client = &http.Client{Transport: tr}
	}
	return u.client
}

/**
 *	Close the HTTP client and clean up resources
 * @returns {error} Returns error if closing fails, nil on success
 *	@description
 *	- Closes idle HTTP connections to prevent resource leaks
 *	- Should be called when the Upgrader is no longer needed
 *	- Safe to call multiple times
 *	@example
 *	upgrader := NewUpgrader("mypkg", config)
 *	defer upgrader.Close()
 */
func (u *Upgrader) Close() error {
	if u.client != nil {
		u.client.CloseIdleConnections()
	}
	return nil
}

/**
 *	从云端获取一个文件的内容
 */
func (u *Upgrader) GetBytes(urlStr string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("GetBytes: %v", err)
	}
	vals := make(url.Values)
	for k, v := range params {
		vals.Set(k, v)
	}
	req.URL.RawQuery = vals.Encode()

	rsp, err := u.getHttpClient().Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("GetBytes: %v", err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		rspBody, _ := io.ReadAll(rsp.Body)
		return rspBody, fmt.Errorf("GetBytes('%s?%s') code:%d, error:%s",
			urlStr, req.URL.RawQuery, rsp.StatusCode, string(rspBody))
	}
	return io.ReadAll(rsp.Body)
}

/**
 *	从服务器获取一个文件
 */
func (u *Upgrader) GetFile(urlStr string, params map[string]string, savePath string) error {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return fmt.Errorf("GetFile('%s') failed: %v", urlStr, err)
	}
	vals := make(url.Values)
	for k, v := range params {
		vals.Set(k, v)
	}
	req.URL.RawQuery = vals.Encode()

	rsp, err := u.getHttpClient().Do(req)
	if err != nil {
		return fmt.Errorf("GetFile('%s') failed: %v", urlStr, err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		rspBody, _ := io.ReadAll(rsp.Body)
		return fmt.Errorf("GetFile('%s', '%s') code: %d, error:%s",
			urlStr, req.URL.RawQuery, rsp.StatusCode, string(rspBody))
	}

	// 创建一个文件用于保存
	if err = os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		return fmt.Errorf("GetFile('%s'): MkdirAll('%s') error:%v", urlStr, savePath, err)
	}
	out, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("GetFile('%s'): create('%s') error: %v", urlStr, savePath, err)
	}
	defer out.Close()

	// 然后将响应流和文件流对接起来
	_, err = io.Copy(out, rsp.Body)
	if err != nil {
		return fmt.Errorf("GetFile('%s'): copy error: %v", urlStr, err)
	}
	return err
}

func (u *Upgrader) GetInstallPaths(pkg *PackageVersion) (dir, fname, fullpath string) {
	if u.TargetPath != "" {
		fullpath = u.TargetPath
		dir, fname = filepath.Split(fullpath)
	} else {
		if pkg.PackageType == PackageTypeZip {
			dir = filepath.Join(u.BaseDir, "components")
			fname = pkg.PackageName
		} else {
			dir, fname = filepath.Split(pkg.FileName)
			if dir != "" {
				dir = filepath.Join(u.BaseDir, dir)
			} else {
				switch pkg.PackageType {
				case PackageTypeExec:
					dir = u.installDir
				case PackageTypeConf:
					dir = filepath.Join(u.BaseDir, "config")
				default:
					dir = filepath.Join(u.BaseDir, "cache", "components")
				}
			}
		}
		fullpath = filepath.Join(dir, fname)
	}
	return
}

/**
 *	获取本地包信息
 *	如果指定了版本，则获取指定版本包信息，否则获取最新版本
 */
func (u *Upgrader) GetLocalVersion(ver *VersionNumber) (pkg PackageVersion, err error) {
	var pkgFile string
	if ver != nil {
		pkgFile = filepath.Join(u.packageDir, "caches", u.packageName, fmt.Sprintf("%s.json", ver.String()))
	} else {
		pkgFile = filepath.Join(u.packageDir, fmt.Sprintf("%s.json", u.packageName))
	}
	err = pkg.Load(pkgFile)
	return
}

/**
 *	从远程库获取包版本
 *	如果具体的os/arch不存在，则尝试获取common/common的通用包
 */
func (u *Upgrader) GetRemoteVersions() (PlatformInfo, error) {
	//	<base-url>/<package>/<os>/<arch>/platform.json
	urlStr := fmt.Sprintf("%s/%s/%s/%s/platform.json", u.BaseUrl, u.packageName, u.Os, u.Arch)

	bytes, err := u.GetBytes(urlStr, nil)
	if err != nil {
		urlStr = fmt.Sprintf("%s/%s/common/common/platform.json", u.BaseUrl, u.packageName)
		bytes, err = u.GetBytes(urlStr, nil)
		if err != nil {
			return PlatformInfo{}, err
		}
	}
	vers := &PlatformInfo{}
	if err = json.Unmarshal(bytes, vers); err != nil {
		return *vers, fmt.Errorf("GetRemoteVersions('%s') unmarshal error: %v", urlStr, err)
	}
	return *vers, nil
}

/**
 *	固定版本，令自动升级忽略该包
 */
func (u *Upgrader) AddPinned(pkg PackageVersion) error {
	pinsDir := filepath.Join(u.packageDir, "pins")
	if err := os.MkdirAll(pinsDir, 0775); err != nil {
		log.Printf("Create directory '%s' failed: %v\n", pinsDir, err)
		return err
	}
	//	把包描述文件保存到包文件目录
	pkgFile := filepath.Join(pinsDir, fmt.Sprintf("%s.json", u.packageName))
	return pkg.Save(pkgFile)
}

func (u *Upgrader) RemovePinned() {
	pkgFile := filepath.Join(u.packageDir, "pins", fmt.Sprintf("%s.json", u.packageName))
	if _, err := os.Stat(pkgFile); err == nil {
		if err := os.Remove(pkgFile); err != nil {
			log.Printf("Remove '%s' failed: %v", pkgFile, err)
		}
	}
}

func (u *Upgrader) GetPinned() (pkg PackageVersion, err error) {
	pkgFile := filepath.Join(u.packageDir, "pins", fmt.Sprintf("%s.json", u.packageName))
	err = pkg.Load(pkgFile)
	return
}

func (u *Upgrader) AddTodo(pkg PackageVersion) error {
	todosDir := filepath.Join(u.packageDir, "todos")
	if err := os.MkdirAll(todosDir, 0775); err != nil {
		log.Printf("Create directory '%s' failed: %v\n", todosDir, err)
		return err
	}
	pkgFile := filepath.Join(todosDir, fmt.Sprintf("%s.json", u.packageName))
	return pkg.Save(pkgFile)
}

func (u *Upgrader) RemoveTodo() {
	pkgFile := filepath.Join(u.packageDir, "todos", fmt.Sprintf("%s.json", u.packageName))
	if _, err := os.Stat(pkgFile); err == nil {
		if err := os.Remove(pkgFile); err != nil {
			log.Printf("Remove '%s' failed: %v", pkgFile, err)
		}
	}
}

func (u *Upgrader) GetTodo() (pkg PackageVersion, err error) {
	pkgFile := filepath.Join(u.packageDir, "todos", fmt.Sprintf("%s.json", u.packageName))
	err = pkg.Load(pkgFile)
	return
}

/**
 *	获取包(需要校验保证包的合法性)
 */
func (u *Upgrader) GetPackage(specVer *VersionNumber) (PackageVersion, bool, error) {
	var curVer VersionNumber

	//	获取本地版本信息
	pkg, err := u.GetLocalVersion(nil)
	if err == nil {
		curVer = pkg.VersionId
		if specVer != nil && CompareVersion(curVer, *specVer) == 0 {
			return pkg, false, nil
		}
	}
	//	获取云端的版本列表
	vers, err := u.GetRemoteVersions()
	if err != nil {
		log.Printf("Get remote versions for package '%s' failed: %v\n", u.packageName, err)
		return pkg, false, err
	}

	addr := VersionAddr{}
	if specVer != nil { //升级指定版本
		//	检查指定版本specVer在不在版本列表中
		found := false
		for _, v := range vers.Versions {
			if CompareVersion(v.VersionId, *specVer) == 0 {
				addr = v
				found = true
				break
			}
		}
		if !found {
			log.Printf("Specified version %s not found for package '%s'\n", specVer.String(), u.packageName)
			return pkg, false, fmt.Errorf("version %s isn't exist", specVer.String())
		}
	} else { //升级最新版本
		ret := CompareVersion(curVer, vers.Newest.VersionId)
		if ret >= 0 {
			return pkg, false, nil
		}
		addr = vers.Newest
	}
	if pkg, err := u.checkLocalPackage(addr.VersionId); err == nil {
		return pkg, true, nil
	}
	//	获取云端升级包的描述信息
	data, err := u.GetBytes(u.BaseUrl+addr.InfoUrl, nil)
	if err != nil {
		log.Printf("Get package info from '%s' failed: %v\n", addr.InfoUrl, err)
		return pkg, false, err
	}
	if err = json.Unmarshal(data, &pkg); err != nil {
		log.Printf("Unmarshal package info from '%s' failed: %v\n", addr.InfoUrl, err)
		return pkg, false, err
	}
	if err = pkg.Verify(); err != nil {
		log.Printf("Invalid package file '%s': %v\n", addr.InfoUrl, err)
		return pkg, false, err
	}
	cacheDir := filepath.Join(u.packageDir, "caches", u.packageName, addr.VersionId.String())
	if err = os.MkdirAll(cacheDir, 0775); err != nil {
		log.Printf("Create cache directory '%s' failed: %v\n", cacheDir, err)
		return pkg, false, err
	}
	//	下载包
	_, fname := filepath.Split(pkg.FileName)
	cacheFname := filepath.Join(cacheDir, fname)
	if err = u.GetFile(u.BaseUrl+addr.AppUrl, nil, cacheFname); err != nil {
		log.Printf("Download package from '%s' to '%s' failed: %v\n", u.BaseUrl+addr.AppUrl, cacheFname, err)
		return pkg, false, err
	}
	//	验证下载文件的完整性，防止丢失、篡改等
	if err := u.verifyIntegrity(pkg, cacheFname); err != nil {
		return pkg, false, err
	}
	//	把包描述文件保存到包文件目录
	pkgFile := filepath.Join(u.packageDir, "caches", u.packageName, fmt.Sprintf("%s.json", pkg.VersionId.String()))
	if err := os.WriteFile(pkgFile, data, 0644); err != nil {
		log.Printf("Write package info file '%s' failed: %v\n", pkgFile, err)
		return pkg, false, err
	}
	//	清理掉太老的版本包
	u.CleanupPackage(3)
	return pkg, true, nil
}

/**
 *	激活版本ver的包，令其成为当前版本
 */
func (u *Upgrader) ActivatePackage(pkg PackageVersion) error {
	if err := u.activatePackage(pkg); err != nil {
		return err
	}
	u.AddPinned(pkg)
	return nil
}

/**
 *	升级包
 */
func (u *Upgrader) UpgradePackage(specVer *VersionNumber) (PackageVersion, bool, error) {
	pkg, upgraded, err := u.GetPackage(specVer)
	if err != nil {
		return pkg, false, err
	}
	if !upgraded { //不需要更新，所以不需要激活
		return pkg, false, nil
	}
	u.AddTodo(pkg)
	if err := u.activatePackage(pkg); err != nil {
		return pkg, false, err
	}
	u.RemoveTodo()
	u.RemovePinned()
	return pkg, true, nil
}

/**
 *	移除指定版本或当前版本的包
 */
func (u *Upgrader) RemovePackage(ver *VersionNumber) error {
	if ver != nil {
		return u.removeSpecialVersion(*ver)
	}
	// 读取包描述文件
	pkgFile := filepath.Join(u.packageDir, fmt.Sprintf("%s.json", u.packageName))
	var pkg PackageVersion
	if err := pkg.Load(pkgFile); err != nil {
		return nil
	}
	u.removeSpecialVersion(pkg.VersionId)
	// 删除包安装后的数据文件或目录
	_, _, fullpath := u.GetInstallPaths(&pkg)
	// 检查文件或目录是否存在，如果存在则删除
	if _, err := os.Stat(fullpath); err == nil {
		if err := os.Remove(fullpath); err != nil {
			return fmt.Errorf("RemovePackage: remove package '%s' failed: %v", fullpath, err)
		}
		log.Printf("Package '%s' removed successfully\n", fullpath)
	}

	// 删除包描述文件
	if err := os.Remove(pkgFile); err != nil {
		return fmt.Errorf("RemovePackage: remove package description file '%s' failed: %v", pkgFile, err)
	}

	log.Printf("Package '%s' removed successfully\n", u.packageName)
	return nil
}

/**
 *	清理包缓存中太老的版本包
 */
func (u *Upgrader) CleanupPackage(reserveNum int) {
	packageCacheDir := filepath.Join(u.packageDir, "caches", u.packageName)
	cleanupPackageOlders(packageCacheDir, reserveNum)
}

/**
 * 清理package目录下过老的版本包数据
 * @param {string} baseDir - costrict数据所在的基路径，如果为空则使用默认路径
 * @returns {error} 返回错误对象，成功时返回nil
 * @description
 * - 扫描版本描述文件package/x-{ver}.json文件，提取文件中保存的版本信息
 * - 保证每个模块只保留最新的三个包，过老的包需要清除
 * - 删除过老的包描述文件x-{ver}.json和package/{ver}/{targetFile}
 * - 支持自定义baseDir，如果为空则使用默认的.costrict目录
 * - 按包名分组处理，每个包保留最新的三个版本
 * @throws
 * - 读取package目录失败
 * - 解析版本描述文件失败
 * - 删除包文件或描述文件失败
 */
func (u *Upgrader) CleanupOlders(reserveNum int) error {
	cacheDir := filepath.Join(u.packageDir, "caches")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		log.Printf("Cleanup: package directory '%s' does not exist\n", cacheDir)
		return err
	}
	dirs, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Printf("Cleanup: package directory '%s' read failed: %v\n", cacheDir, err)
		return err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		dirname := dir.Name()
		if dirname == "." || dirname == ".." {
			continue
		}
		packageCacheDir := filepath.Join(cacheDir, dirname)
		cleanupPackageOlders(packageCacheDir, reserveNum)
	}
	return nil
}

/**
 * Remove oldest package versions from cache directory, keeping only the latest reserveNum versions
 * @param {string} packageCacheDir - Package cache directory path
 * @param {int} reserveNum - Number of latest versions to keep
 * @description
 * - Scans packageCacheDir for version files (format: x.x.x.json)
 * - Parses version numbers from filenames
 * - Sorts versions from newest to oldest using CompareVersion
 * - Keeps the latest reserveNum versions, deletes the rest
 * - Deletes both the version JSON file and its corresponding subdirectory
 * - Skips invalid version format files
 * @throws
 * - Directory reading failure (logged and continues)
 * - Version parsing failure (skips invalid versions)
 * - File/directory removal failure (logged)
 * @example
 * cleanupPackageOlders("/path/to/cache", 3)
 */
func cleanupPackageOlders(packageCacheDir string, reserveNum int) {
	if _, err := os.Stat(packageCacheDir); os.IsNotExist(err) {
		log.Printf("Cleanup: directory '%s' does not exist\n", packageCacheDir)
		return
	}

	// Read all files in the directory
	files, err := os.ReadDir(packageCacheDir)
	if err != nil {
		log.Printf("Cleanup: read directory '%s' failed: %v\n", packageCacheDir, err)
		return
	}

	// Collect version information from *.json files
	var versions []struct {
		version  VersionNumber
		jsonFile string
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		// Check if file matches version format: x.x.x.json
		if !strings.HasSuffix(filename, ".json") {
			continue
		}

		// Remove .json suffix to get version string
		versionStr := filename[:len(filename)-5]
		// Parse version string
		var ver VersionNumber
		if err := ver.Parse(versionStr); err != nil {
			log.Printf("Cleanup: invalid version format '%s': %v\n", versionStr, err)
			continue
		}

		versions = append(versions, struct {
			version  VersionNumber
			jsonFile string
		}{
			version:  ver,
			jsonFile: filepath.Join(packageCacheDir, filename),
		})
	}

	// Sort versions from newest to oldest
	sort.Slice(versions, func(i, j int) bool {
		return CompareVersion(versions[i].version, versions[j].version) > 0
	})

	// Delete old versions (keep only the first reserveNum versions)
	for i := reserveNum; i < len(versions); i++ {
		old := versions[i]
		versionStr := old.version.String()
		versionDir := filepath.Join(packageCacheDir, versionStr)

		// Delete version JSON file
		if err := os.Remove(old.jsonFile); err != nil {
			log.Printf("Cleanup: remove JSON file '%s' failed: %v\n", old.jsonFile, err)
		} else {
			log.Printf("Cleanup: JSON file '%s' removed\n", old.jsonFile)
		}

		// Delete version directory
		if _, err := os.Stat(versionDir); err == nil {
			if err := os.RemoveAll(versionDir); err != nil {
				log.Printf("Cleanup: remove directory '%s' failed: %v\n", versionDir, err)
			} else {
				log.Printf("Cleanup: directory '%s' removed\n", versionDir)
			}
		}
	}
}

func (u *Upgrader) checkLocalPackage(ver VersionNumber) (PackageVersion, error) {
	cacheDir := filepath.Join(u.packageDir, "caches", u.packageName)
	pkgFile := filepath.Join(cacheDir, fmt.Sprintf("%s.json", ver.String()))
	var pkg PackageVersion
	if err := pkg.Load(pkgFile); err != nil {
		return pkg, err
	}
	_, fname := filepath.Split(pkg.FileName)
	cacheFname := filepath.Join(cacheDir, ver.String(), fname)
	if err := u.verifyIntegrity(pkg, cacheFname); err != nil {
		return pkg, err
	}
	return pkg, nil
}

func (u *Upgrader) verifyIntegrity(pkg PackageVersion, fname string) error {
	_, md5str, err := CalcFileMd5(fname)
	if err != nil {
		log.Printf("Calculate MD5 for file '%s' failed: %v\n", fname, err)
		return err
	}
	if md5str != pkg.Checksum {
		log.Printf("MD5 checksum mismatch for package '%s'. Expected: %s, Actual: %s\n", pkg.PackageName, pkg.Checksum, md5str)
		return fmt.Errorf("checksum error")
	}
	//	检查签名，防止包被篡改
	sig, err := hex.DecodeString(pkg.Sign)
	if err != nil {
		log.Printf("Decode signature for package '%s' failed: %v\n", pkg.PackageName, err)
		return err
	}
	if err = VerifySign([]byte(u.PublicKey), sig, []byte(md5str)); err != nil {
		log.Printf("Verify signature for package '%s' failed: %v\n", pkg.PackageName, err)
		return err
	}
	return nil
}

/**
 *	激活版本ver的包，令其成为当前版本
 */
func (u *Upgrader) activatePackage(pkg PackageVersion) error {
	_, fname := filepath.Split(pkg.FileName)
	cacheDir := filepath.Join(u.packageDir, "caches", u.packageName, pkg.VersionId.String())
	cacheFname := filepath.Join(cacheDir, fname)
	//	把下载的包安装到正式目录
	if err := u.installPackage(pkg, cacheFname); err != nil {
		log.Printf("Install package '%s' failed: %v\n", cacheFname, err)
		return err
	}
	pkgFile := filepath.Join(u.packageDir, fmt.Sprintf("%s.json", u.packageName))
	return pkg.Save(pkgFile)
}

/**
 * Unzip package data from cache file to install path
 * @param {*PackageVersion} pkg - Package version information
 * @param {string} cacheFname - Path to the cached zip file
 * @returns {error} Returns error if unzip fails, nil on success
 * @description
 * - Removes all existing content in the target directory before unzipping
 * - Creates the target directory if it doesn't exist
 * - Extracts all files and directories from the zip archive
 * - Preserves file and directory permissions from the zip
 * - Handles both regular files and subdirectories
 * @throws
 * - Failed to open or read zip file
 * - Failed to create destination directory
 * - Failed to create or write destination files
 * @example
 * err := upgrader.unzipPackage(&pkg, "/path/to/cache/file.zip")
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (u *Upgrader) unzipPackage(pkg *PackageVersion, cacheFname string) error {
	_, _, fullpath := u.GetInstallPaths(pkg)

	if _, err := os.Stat(fullpath); err == nil {
		if err := os.RemoveAll(fullpath); err != nil {
			return fmt.Errorf("unzipPackage: failed to remove directory '%s': %v", fullpath, err)
		}
		log.Printf("Extract and remove the existing directory '%s'\n", fullpath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("unzipPackage: failed to stat directory '%s': %v", fullpath, err)
	}

	// Create target directory
	if err := os.MkdirAll(fullpath, 0755); err != nil {
		return fmt.Errorf("unzipPackage: failed to create directory '%s': %v", fullpath, err)
	}

	// Open zip file
	reader, err := zip.OpenReader(cacheFname)
	if err != nil {
		return fmt.Errorf("unzipPackage: failed to open zip file '%s': %v", cacheFname, err)
	}
	defer reader.Close()

	// Extract all files from zip
	for _, file := range reader.File {
		// Construct destination file path
		destFilePath := filepath.Join(fullpath, file.Name)

		// Security check: prevent zip slip attack
		if !strings.HasPrefix(destFilePath, filepath.Clean(fullpath)+string(os.PathSeparator)) {
			return fmt.Errorf("unzipPackage: invalid file path '%s': path traversal detected", destFilePath)
		}

		// Create directory if entry is a directory
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destFilePath, file.Mode()); err != nil {
				return fmt.Errorf("unzipPackage: failed to create directory '%s': %v", destFilePath, err)
			}
			continue
		}

		// Create directory for the file if needed
		destDir := filepath.Dir(destFilePath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("unzipPackage: failed to create directory '%s': %v", destDir, err)
		}

		// Open file in zip
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("unzipPackage: failed to open file '%s' in zip: %v", file.Name, err)
		}
		defer srcFile.Close()

		// Create destination file
		dstFile, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("unzipPackage: failed to create file '%s': %v", destFilePath, err)
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("unzipPackage: failed to copy file '%s': %v", file.Name, err)
		}

		log.Printf("Extracted '%s' to '%s'\n", file.Name, destFilePath)
	}

	log.Printf("Successfully unzipped '%s' to '%s'\n", cacheFname, fullpath)
	return nil
}

/**
 *	保存包数据文件
 */
func (u *Upgrader) savePackageData(pkg PackageVersion, cacheFname string) error {
	if pkg.PackageType == PackageTypeZip {
		return u.unzipPackage(&pkg, cacheFname)
	}
	dir, _, fullpath := u.GetInstallPaths(&pkg)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.Remove(fullpath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 拷贝文件而不是重命名
	srcFile, err := os.Open(cacheFname)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	if pkg.PackageType != PackageTypeExec {
		return nil
	}
	return os.Chmod(fullpath, 0755)
}

/**
 *	在windows上设置PATH变量，让新安装的程序可以被执行
 */
func windowsSetPATH(installDir string) error {
	paths := os.Getenv("PATH")
	if !strings.Contains(paths, installDir) {
		newPath := fmt.Sprintf("%s;%s", paths, installDir)
		cmd := exec.Command("setx", "PATH", newPath)
		// cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // 隐藏命令窗口
		if err := cmd.Run(); err != nil {
			return err
		}
		os.Setenv("PATH", newPath)
	}
	return nil
}

/**
 *	在linux上设置PATH变量，让新安装的程序可以被执行
 */
func linuxSetPATH(installDir string) error {
	currentPath := os.Getenv("PATH")
	// 检查是否已经包含该路径
	currentPathStr := strings.TrimSpace(currentPath)
	if strings.Contains(currentPathStr, installDir) {
		log.Println("The path is already in PATH.")
		return nil
	}
	// 将新路径添加到 PATH
	newPathStr := fmt.Sprintf("%s:%s", currentPathStr, installDir)
	err := os.Setenv("PATH", newPathStr)
	if err != nil {
		log.Printf("Failed to set PATH for current process: %v\n", err)
		return err
	}
	// 获取当前用户的主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get user home directory: %v\n", err)
		return err
	}
	envLine := fmt.Sprintf("export PATH=$PATH:%s", installDir)

	bashrcPath := homeDir + "/.bashrc"
	// 检查是否已经包含该环境变量
	file, err := os.Open(bashrcPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to open ~/.bashrc: %v\n", err)
			return err
		}
		// 文件不存在，创建一个空文件
		file, err = os.Create(bashrcPath)
		if err != nil {
			log.Printf("Failed to create ~/.bashrc: %v\n", err)
			return err
		}
		file.Close()
	} else {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), envLine) {
				file.Close()
				log.Println("Environment variable already exists in ~/.bashrc.")
				return nil
			}
		}
		file.Close()
		if err := scanner.Err(); err != nil {
			log.Printf("Failed to read ~/.bashrc: %v\n", err)
			return err
		}
	}
	// 将环境变量追加到 ~/.bashrc 文件
	file, err = os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open ~/.bashrc for appending: %v\n", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(envLine + "\n")
	if err != nil {
		log.Printf("Failed to write environment variable to ~/.bashrc: %v\n", err)
		return err
	}

	log.Println("Environment variable added to ~/.bashrc successfully.")
	return nil
}

/**
 *	安装包数据
 */
func (u *Upgrader) installPackage(pkg PackageVersion, cacheFname string) error {
	if err := u.savePackageData(pkg, cacheFname); err != nil {
		return err
	}
	if pkg.PackageType != PackageTypeExec {
		return nil
	}
	if u.NoSetPath {
		return nil
	}
	if runtime.GOOS == "windows" {
		return windowsSetPATH(u.installDir)
	} else {
		return linuxSetPATH(u.installDir)
	}
}

func (u *Upgrader) removeSpecialVersion(ver VersionNumber) error {
	cacheDir := filepath.Join(u.packageDir, "caches", u.packageName)
	// 读取包描述文件
	pkgFile := filepath.Join(cacheDir, fmt.Sprintf("%s.json", ver.String()))
	var pkg PackageVersion
	if err := pkg.Load(pkgFile); err != nil {
		//认为包已移除，不报错
		return nil
	}

	// 检查文件package/caches/x.x.x/xx是否存在，如果存在则删除
	_, fname := filepath.Split(pkg.FileName)
	cacheFname := filepath.Join(cacheDir, ver.String(), fname)
	if _, err := os.Stat(cacheFname); err == nil {
		if err := os.Remove(cacheFname); err != nil {
			return err
		}
	}

	// 删除包描述文件 package/caches/x.x.x.json
	if err := os.Remove(pkgFile); err != nil {
		return err
	}
	if isDirEmpty(cacheDir) {
		if err := os.Remove(cacheDir); err != nil {
			log.Printf("Package directory '%s' remove failed: %v\n", cacheDir, err)
		} else {
			log.Printf("Package directory '%s' removed\n", cacheDir)
		}
	}
	log.Printf("Package '%s-%s' removed successfully\n", u.packageName, ver.String())
	return nil
}

/**
 * 检查目录是否为空
 * @param {string} dirPath - 目录路径
 * @returns {bool} 目录为空返回true，否则返回false
 * @description
 * - 检查指定目录是否为空（不包含任何文件或子目录）
 * - 如果目录不存在，返回true
 * - 如果目录存在但为空，返回true
 * - 如果目录存在且包含文件或子目录，返回false
 * @throws
 * - 读取目录失败时返回false
 * @example
 * if isDirEmpty("/path/to/dir") {
 *     os.Remove("/path/to/dir")
 * }
 */
func isDirEmpty(dirPath string) bool {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return true
	}
	file, err := os.Open(dirPath)
	if err != nil {
		return false
	}
	defer file.Close()

	_, err = file.Readdirnames(1)
	return err == io.EOF
}

/**
 *	获取costrict目录结构设定
 */
func getCostrictDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".costrict")
}

func (u *Upgrader) correct() {
	if u.Arch == "" {
		u.Arch = runtime.GOARCH
	}
	if u.Os == "" {
		u.Os = runtime.GOOS
	}
	if u.BaseUrl == "" {
		u.BaseUrl = SHENMA_BASE_URL
	}
	if u.PublicKey == "" {
		u.PublicKey = SHENMA_PUBLIC_KEY
	}
	if u.BaseDir == "" {
		u.BaseDir = getCostrictDir()
	}
	u.installDir = filepath.Join(u.BaseDir, "bin")
	u.packageDir = filepath.Join(u.BaseDir, "package")
}
