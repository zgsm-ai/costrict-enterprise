package providers

import (
	"errors"
	"sync"
)

var (
	managerInstance *OAuthManager
	managerOnce     sync.Once
)

func GetManager() *OAuthManager {
	managerOnce.Do(func() {
		managerInstance = NewOAuthManager()
		managerInstance.RegisterFactory("casdoor", NewCasdoorFactory())
	})
	return managerInstance
}

func InitializeProviders(configs map[string]*ProviderConfig) error {
	manager := GetManager()
	if manager == nil {
		return errors.New("GetManager failed")
	}
	for name, config := range configs {
		manager.SetConfig(name, config)
	}
	return nil
}
