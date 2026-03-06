package models

type PackageDetail struct {
	PackageType string `json:"packageType"` //包类型: exec/conf
	FileName    string `json:"fileName"`    //被打包的文件的相对路径(相对.costrict目录,为空则安装到默认路径)
	Size        uint64 `json:"size"`        //包文件大小
	Version     string `json:"version"`     //版本号，采用SemVer标准
	Build       string `json:"build"`       //构建信息：Tag/Branch信息 CommitID BuildTime
	Description string `json:"description"` //版本描述，含有更丰富的可读信息
}

type PackageRepo struct {
	Newest   string   `json:"newest"`
	Versions []string `json:"versions"`
}

type ComponentDetail struct {
	Name        string                 `json:"name"`
	Spec        ComponentSpecification `json:"spec"`
	Local       PackageDetail          `json:"local"`
	Remote      PackageRepo            `json:"remote"`
	Installed   bool                   `json:"installed"`
	NeedUpgrade bool                   `json:"need_upgrade"`
}
