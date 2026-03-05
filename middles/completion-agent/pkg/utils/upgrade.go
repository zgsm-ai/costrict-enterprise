package utils

import (
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
	PackageName  string        `json:"packageName"`  //包名字
	PackageType  PackageType   `json:"packageType"`  //包类型: exec/conf
	FileName     string        `json:"fileName"`     //被打包的文件的相对路径(相对.costrict目录,为空则安装到默认路径)
	Os           string        `json:"os"`           //操作系统名:linux/windows
	Arch         string        `json:"arch"`         //硬件架构
	Size         uint64        `json:"size"`         //包文件大小
	Checksum     string        `json:"checksum"`     //Md5散列值
	Sign         string        `json:"sign"`         //签名，使用私钥签的名，需要用对应公钥验证
	ChecksumAlgo string        `json:"checksumAlgo"` //固定为“md5”
	VersionId    VersionNumber `json:"versionId"`    //版本号，采用SemVer标准
	Build        string        `json:"build"`        //构建信息：Tag/Branch信息 CommitID BuildTime
	Description  string        `json:"description"`  //版本描述，含有更丰富的可读信息
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
}

type Upgrader struct {
	UpgradeConfig

	packageName string //包名称
	installDir  string
	packageDir  string
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

//------------------------------------------------------------------------------
//	Get data from cloud
//------------------------------------------------------------------------------
/**
 *	从云端获取一个文件的内容
 */
func GetBytes(urlStr string, params map[string]string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("GetBytes: %v", err)
	}
	vals := make(url.Values)
	for k, v := range params {
		vals.Set(k, v)
	}
	req.URL.RawQuery = vals.Encode()

	rsp, err := client.Do(req)
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
func GetFile(urlStr string, params map[string]string, savePath string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return fmt.Errorf("GetFile('%s') failed: %v", urlStr, err)
	}
	vals := make(url.Values)
	for k, v := range params {
		vals.Set(k, v)
	}
	req.URL.RawQuery = vals.Encode()

	rsp, err := client.Do(req)
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

//------------------------------------------------------------------------------
//	VersionNumber
//------------------------------------------------------------------------------

func (ver *VersionNumber) String() string {
	return fmt.Sprintf("%d.%d.%d", ver.Major, ver.Minor, ver.Micro)
}

func (ver *VersionNumber) Parse(verstr string) error {
	var err error
	var major, minor, micro int

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
	if pkg.PackageType != "exec" && pkg.PackageType != "conf" {
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

func NewUpgrader(packageName string, cfg UpgradeConfig) *Upgrader {
	u := &Upgrader{}
	u.UpgradeConfig = cfg
	u.packageName = packageName
	u.correct()
	return u
}

/**
 *	获取本地包信息
 *	如果指定了版本，则获取指定版本包信息，否则获取最新版本
 */
func (u *Upgrader) GetLocalVersion(ver *VersionNumber) (pkg PackageVersion, err error) {
	var pkgFile string
	if ver != nil {
		pkgFile = filepath.Join(u.packageDir, fmt.Sprintf("%s-%s.json", u.packageName, ver.String()))
	} else {
		pkgFile = filepath.Join(u.packageDir, fmt.Sprintf("%s.json", u.packageName))
	}
	err = pkg.Load(pkgFile)
	return
}

/**
 *	从远程库获取包版本
 */
func (u *Upgrader) GetRemoteVersions() (PlatformInfo, error) {
	//	<base-url>/<package>/<os>/<arch>/platform.json
	urlStr := fmt.Sprintf("%s/%s/%s/%s/platform.json", u.BaseUrl, u.packageName, u.Os, u.Arch)

	bytes, err := GetBytes(urlStr, nil)
	if err != nil {
		return PlatformInfo{}, err
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
	var pkg PackageVersion
	var curVer VersionNumber

	//	获取本地版本信息
	pkgFile := filepath.Join(u.packageDir, fmt.Sprintf("%s.json", u.packageName))
	if err := pkg.Load(pkgFile); err == nil {
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
	data, err := GetBytes(u.BaseUrl+addr.InfoUrl, nil)
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
	cacheDir := filepath.Join(u.packageDir, addr.VersionId.String())
	if err = os.MkdirAll(cacheDir, 0775); err != nil {
		log.Printf("Create cache directory '%s' failed: %v\n", cacheDir, err)
		return pkg, false, err
	}
	//	下载包
	_, fname := filepath.Split(pkg.FileName)
	cacheFname := filepath.Join(cacheDir, fname)
	if err = GetFile(u.BaseUrl+addr.AppUrl, nil, cacheFname); err != nil {
		log.Printf("Download package from '%s' to '%s' failed: %v\n", addr.AppUrl, cacheFname, err)
		return pkg, false, err
	}
	//	验证下载文件的完整性，防止丢失、篡改等
	if err := u.verifyIntegrity(pkg, cacheFname); err != nil {
		return pkg, false, err
	}
	//	把包描述文件保存到包文件目录
	pkgFile = filepath.Join(u.packageDir, fmt.Sprintf("%s-%s.json", u.packageName, pkg.VersionId.String()))
	if err := os.WriteFile(pkgFile, data, 0644); err != nil {
		log.Printf("Write package info file '%s' failed: %v\n", pkgFile, err)
		return pkg, false, err
	}
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
 *	移除指定名字的包
 *	@param {string} packageName - 要移除的包名称
 *	@param {string} baseDir - costrict数据所在的基路径，如果为空则使用默认路径
 *	@returns {error} 返回错误对象，成功时返回nil
 *	@description
 *	- 移除指定包的所有相关文件，包括包描述文件和安装的包文件
 *	- 首先读取包描述信息以确定需要删除的文件位置
 *	- 支持自定义baseDir，如果为空则使用默认的.costrict目录
 *	- 如果包不存在或已删除，不会报错
 *	@throws
 *	- 读取包描述文件失败
 *	- 删除包文件失败
 *	- 删除包描述文件失败
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
	// 删除包数据文件
	var dataPath string
	dir, fname := filepath.Split(pkg.FileName)
	if dir != "" {
		dataPath = filepath.Join(u.BaseDir, pkg.FileName)
	} else {
		dataPath = filepath.Join(u.installDir, fname)
	}

	// 检查文件是否存在，如果存在则删除
	if _, err := os.Stat(dataPath); err == nil {
		if err := os.Remove(dataPath); err != nil {
			return fmt.Errorf("RemovePackage: remove package file '%s' failed: %v", dataPath, err)
		}
		log.Printf("Package file '%s' removed successfully\n", dataPath)
	}

	// 删除包描述文件
	if err := os.Remove(pkgFile); err != nil {
		return fmt.Errorf("RemovePackage: remove package description file '%s' failed: %v", pkgFile, err)
	}

	log.Printf("Package '%s' removed successfully\n", u.packageName)
	return nil
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
 * @example
 * err := CleanupOldVersions()
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (u *Upgrader) CleanupOldVersions() error {
	// 检查package目录是否存在
	if _, err := os.Stat(u.packageDir); os.IsNotExist(err) {
		log.Printf("Cleanup: package directory '%s' does not exist\n", u.packageDir)
		return err
	}

	// 读取package目录下的所有文件
	files, err := os.ReadDir(u.packageDir)
	if err != nil {
		log.Printf("Cleanup: package directory '%s' read failed: %v\n", u.packageDir, err)
		return err
	}

	// 按包名分组的版本信息
	packageVersions := make(map[string][]VersionSummary)

	// 遍历文件，找出所有版本描述文件（格式：x-{ver}.json）
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		// 匹配格式：{packageName}-{version}.json
		if !strings.HasSuffix(filename, ".json") {
			continue
		}
		// 关注中间带‘-’的版本描述文件
		parts := strings.Split(filename, "-")
		if len(parts) < 2 {
			continue
		}
		// 读取包描述文件
		filePath := filepath.Join(u.packageDir, filename)
		// 解析包描述信息
		var pkg PackageVersion
		if err := pkg.Load(filePath); err != nil {
			log.Printf("Cleanup: Load '%s' failed: %v\n", filePath, err)
			continue
		}
		versionStr := pkg.VersionId.String()
		_, fname := filepath.Split(pkg.FileName)
		// 保存版本信息
		versionInfo := VersionSummary{
			PackageName: pkg.PackageName,
			Version:     pkg.VersionId,
			PackageDir:  filepath.Join(u.packageDir, versionStr),
			DescPath:    filePath,
			DataPath:    filepath.Join(u.packageDir, versionStr, fname),
		}

		packageVersions[pkg.PackageName] = append(packageVersions[pkg.PackageName], versionInfo)
	}

	// 对每个包的版本进行排序，并删除过老的版本
	for _, versions := range packageVersions {
		// 按版本号从新到旧排序
		sort.Slice(versions, func(i, j int) bool {
			return CompareVersion(versions[i].Version, versions[j].Version) > 0
		})
		removeOldestVersions(versions, 3)
	}

	return nil
}

// VersionSummary 包版本的摘要，用于清理过老版本
type VersionSummary struct {
	PackageName string        // 包名
	Version     VersionNumber // 版本号
	DescPath    string        // 包描述文件路径
	PackageDir  string        // 包目录路径
	DataPath    string        // 包数据文件路径
}

/**
 *	删除过老版本，但保留开头即最新的reserveNum个版本
 */
func removeOldestVersions(versions []VersionSummary, reserveNum int) {
	// 如果版本数量超过保留数目，则删除过老的版本
	for i := reserveNum; i < len(versions); i++ {
		old := versions[i]

		// 删除包描述文件
		if err := os.Remove(old.DescPath); err != nil {
			log.Printf("Cleanup: remove description file '%s' failed: %v\n", old.DescPath, err)
		} else {
			log.Printf("Cleanup: description file '%s' removed\n", old.DescPath)
		}

		// 删除包数据文件
		if err := os.Remove(old.DataPath); err != nil {
			log.Printf("Cleanup: remove data file '%s' failed: %v\n", old.DataPath, err)
		} else {
			log.Printf("Cleanup: data file '%s' removed\n", old.DataPath)
		}

		// 检查目录是否为空，如果为空则删除目录
		if isDirEmpty(old.PackageDir) {
			if err := os.Remove(old.PackageDir); err != nil {
				log.Printf("Cleanup: remove directory '%s' failed: %v\n", old.PackageDir, err)
			} else {
				log.Printf("Cleanup: package directory '%s' removed\n", old.PackageDir)
			}
		}
	}
}

func (u *Upgrader) checkLocalPackage(ver VersionNumber) (PackageVersion, error) {
	pkgFile := filepath.Join(u.packageDir, fmt.Sprintf("%s-%s.json", u.packageName, ver.String()))
	var pkg PackageVersion
	if err := pkg.Load(pkgFile); err != nil {
		return pkg, err
	}
	_, fname := filepath.Split(pkg.FileName)
	cacheFname := filepath.Join(u.packageDir, ver.String(), fname)
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
	cacheDir := filepath.Join(u.packageDir, pkg.VersionId.String())
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
 *	保存包数据文件
 */
func (u *Upgrader) savePackageData(pkg PackageVersion, cacheFname string) error {
	var dataPath string
	if u.TargetPath != "" {
		dataPath = u.TargetPath
	} else {
		dir, fname := filepath.Split(pkg.FileName)
		if dir != "" {
			dataPath = filepath.Join(u.BaseDir, pkg.FileName)
		} else {
			dataPath = filepath.Join(u.installDir, fname)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dataPath), 0755); err != nil {
		return err
	}
	if err := os.Remove(dataPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 拷贝文件而不是重命名
	srcFile, err := os.Open(cacheFname)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dataPath)
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
	return os.Chmod(dataPath, 0755)
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
	// 读取包描述文件
	pkgFile := filepath.Join(u.packageDir, fmt.Sprintf("%s-%s.json", u.packageName, ver.String()))
	var pkg PackageVersion
	if err := pkg.Load(pkgFile); err != nil {
		//认为包已移除，不报错
		return nil
	}

	_, fname := filepath.Split(pkg.FileName)
	cacheDir := filepath.Join(u.packageDir, ver.String())
	cacheFname := filepath.Join(cacheDir, fname)
	// 检查文件是否存在，如果存在则删除
	if _, err := os.Stat(cacheFname); err == nil {
		if err := os.Remove(cacheFname); err != nil {
			return err
		}
	}

	// 删除包描述文件
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
