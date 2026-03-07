package config

import (
	"errors"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		UserInfoHeader string
	}
	ProxyConfig *ProxyConfig `json:"proxy_config" yaml:"proxy_config"`
}

// Validate 实现 Validator 接口
func (c Config) Validate() error {
	if len(c.Name) == 0 {
		return errors.New("name 不能为空")
	}
	if c.ProxyConfig != nil {
		if err := c.ProxyConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}
