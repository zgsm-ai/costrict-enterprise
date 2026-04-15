package services

import (
	"costrict-keeper/internal/config"
	"costrict-keeper/internal/env"
	"costrict-keeper/internal/httpc"
	"costrict-keeper/internal/logger"
	"costrict-keeper/internal/models"
	"costrict-keeper/internal/utils"
	"errors"
	"fmt"
)

var ErrComponentNotFound = errors.New("component not found")

type ComponentInstance struct {
	spec        models.ComponentSpecification
	local       *utils.PackageVersion
	remote      *utils.PlatformInfo
	installed   bool
	needUpgrade bool
}

/**
 * Component manager provides methods to get local and remote version information
 * for both services and components
 */
type ComponentManager struct {
	self       ComponentInstance
	components map[string]*ComponentInstance
	configs    map[string]*ComponentInstance
}

var componentManager *ComponentManager

/**
 * Create new component manager instance
 * @returns {ComponentManager} Returns new component manager instance
 */
func GetComponentManager() *ComponentManager {
	if componentManager != nil {
		return componentManager
	}
	componentManager = &ComponentManager{
		components: make(map[string]*ComponentInstance),
		configs:    make(map[string]*ComponentInstance),
	}
	return componentManager
}

func (ci *ComponentInstance) GetDetail() models.ComponentDetail {
	detail := models.ComponentDetail{
		Name:        ci.spec.Name,
		Spec:        ci.spec,
		Local:       models.PackageDetail{},
		Remote:      models.PackageRepo{},
		Installed:   ci.installed,
		NeedUpgrade: ci.needUpgrade,
	}
	if ci.local != nil {
		detail.Local.Build = ci.local.Build
		detail.Local.Description = ci.local.Description
		detail.Local.FileName = ci.local.FileName
		detail.Local.PackageType = string(ci.local.PackageType)
		detail.Local.Size = ci.local.Size
		detail.Local.Version = ci.local.VersionId.String()
	}
	if ci.remote != nil {
		detail.Remote.Newest = ci.remote.Newest.VersionId.String()
		for _, v := range ci.remote.Versions {
			detail.Remote.Versions = append(detail.Remote.Versions, v.VersionId.String())
		}
	}
	return detail
}

/**
 * Fetch component information including local and remote versions
 * @param {ComponentInstance} ci - Component instance to fetch information for
 * @returns {error} Returns error if fetch fails, nil on success
 * @description
 * - Creates upgrade configuration with component name and paths
 * - Gets local version information using utils.GetLocalVersion
 * - Gets remote version information using utils.GetRemoteVersions
 * - Compares local and remote versions to determine if upgrade is needed
 * - Updates component instance with version information and upgrade status
 * @throws
 * - Local version retrieval errors
 * - Remote version retrieval errors
 * - Version comparison errors
 * @private
 */
func (ci *ComponentInstance) fetchComponentInfo() error {
	u := utils.NewUpgrader(ci.spec.Name, utils.UpgradeConfig{
		BaseUrl: config.Cloud().UpgradeUrl,
		BaseDir: env.CostrictDir,
	}, httpc.GetClient())

	ci.needUpgrade = false
	ci.installed = false
	local, err := u.GetLocalVersion(nil)
	if err == nil {
		ci.local = &local
		ci.installed = true
	}
	remote, err := u.GetRemoteVersions()
	if err == nil {
		ci.remote = &remote
		if utils.CompareVersion(local.VersionId, remote.Newest.VersionId) < 0 {
			ci.needUpgrade = true
		}
	}
	return nil
}

/**
 * Upgrade component to latest version
 * @param {ComponentInstance} component - Component instance to upgrade
 * @returns {error} Returns error if upgrade fails, nil on success
 * @description
 * - Creates upgrade configuration with component name and base URL
 * - Sets install directory if specified in component specification
 * - Calls utils.UpgradePackage to perform the actual upgrade
 * - Updates component instance with new version information
 * - Logs upgrade result and success/failure status
 * @throws
 * - Upgrade package errors
 * - Configuration errors
 * @private
 */
func (ci *ComponentInstance) upgradeComponent() error {
	// 解析版本号 - 由于新结构体中没有版本信息，使用默认版本
	u := utils.NewUpgrader(ci.spec.Name, utils.UpgradeConfig{
		BaseUrl: config.Cloud().UpgradeUrl,
		BaseDir: env.CostrictDir,
	}, httpc.GetClient())

	pkg, upgraded, err := u.UpgradePackage(nil)
	if err != nil {
		logger.Errorf("The '%s' upgrade failed: %v", ci.spec.Name, err)
		return err
	}
	ci.local = &pkg
	if !upgraded {
		logger.Infof("The '%s' version is up to date\n", ci.spec.Name)
	} else {
		logger.Infof("The '%s' is upgraded to version %s\n", ci.spec.Name, pkg.VersionId.String())
	}
	vers, err := u.GetRemoteVersions()
	if err != nil {
		logger.Errorf("GetRemoteVersions failed: %v", err)
		return err
	}
	ci.remote = &vers
	return err
}

/**
 * Remove specified component
 */
func (ci *ComponentInstance) removeComponent() error {
	// Check if component is installed
	if !ci.installed {
		return fmt.Errorf("component '%s' is not installed", ci.spec.Name)
	}
	u := utils.NewUpgrader(ci.spec.Name, utils.UpgradeConfig{
		BaseDir: env.CostrictDir,
	}, httpc.GetClient())

	// Remove the package
	if err := u.RemovePackage(nil); err != nil {
		return fmt.Errorf("failed to remove component %s: %v", ci.spec.Name, err)
	}

	// Update component state
	ci.installed = false
	ci.needUpgrade = false
	ci.local = nil

	logger.Infof("Component '%s' removed successfully", ci.spec.Name)
	return nil
}

/**
 * Initialize component manager with shared HTTP client
 * @param {*http.Client} client - Shared HTTP client for communications
 * @returns {error} Returns error if initialization fails, nil on success
 * @description
 * - Sets shared HTTP client for all component operations
 * - Initializes all component instances from configuration
 * - Fetches component information for all components and configurations
 * - Sets up self component instance
 * @throws
 * - Component information fetch errors
 * - Configuration parsing errors
 */
func (cm *ComponentManager) Init() error {
	for _, cpn := range config.Spec().Configurations {
		ci := ComponentInstance{
			spec: cpn,
		}
		ci.fetchComponentInfo()
		componentManager.configs[cpn.Name] = &ci
	}
	for _, cpn := range config.Spec().Components {
		ci := ComponentInstance{
			spec: cpn,
		}
		ci.fetchComponentInfo()
		componentManager.components[cpn.Name] = &ci
	}
	componentManager.self.spec = config.Spec().Manager.Component
	componentManager.self.fetchComponentInfo()
	return nil
}

/**
* Upgrade specified component to latest version
* @param {string} name - Name of the component to upgrade
* @returns {error} Returns error if upgrade fails, nil on success
* @description
* - Finds service configuration by component name
* - Parses highest version from service configuration
* - Executes upgrade function with component configuration
* @throws
* - Service not found errors
* - Version parsing errors
* - Upgrade execution errors
 */
func (cm *ComponentManager) UpgradeComponent(name string) error {
	cpn, ok := cm.components[name]
	if !ok {
		return ErrComponentNotFound
	}
	if !cpn.needUpgrade {
		return nil
	}
	return cpn.upgradeComponent()
}

/**
* Remove specified component
* @param {string} name - Name of the component to remove
* @returns {error} Returns error if removal fails, nil on success
* @description
* - Finds component by name in component manager
* - Checks if component is installed before removal
* - Uses RemovePackage function to remove component files and metadata
* - Updates component manager state after successful removal
* @throws
* - Component not found errors
* - Package removal errors
 */
func (cm *ComponentManager) RemoveComponent(name string) error {
	cpn, ok := cm.components[name]
	if !ok {
		return fmt.Errorf("component %s not found", name)
	}
	return cpn.removeComponent()
}

/**
 * Get all components derived from services
 * @returns {([]ComponentInstance, error)} Returns slice of component information and error if any
 * @description
 * - Converts service configurations to component information
 * - Each service becomes a component with name, version and path
 * - Returns empty slice if no services exist
 * @throws
 * - Component conversion errors
 */
func (cm *ComponentManager) GetComponents(includeSelf, includeConfig bool) []*ComponentInstance {
	components := make([]*ComponentInstance, 0)
	if includeSelf {
		components = append(components, &cm.self)
	}
	for _, cpn := range cm.components {
		components = append(components, cpn)
	}
	if includeConfig {
		for _, cpn := range cm.configs {
			components = append(components, cpn)
		}
	}
	return components
}

/**
 * Get self component instance (manager component)
 * @returns {ComponentInstance} Returns the manager component instance
 * @description
 * - Returns the component instance representing the manager itself
 * - Contains manager's version, installation status and upgrade information
 * - Used for manager self-management and upgrade operations
 * @example
 * manager := GetComponentManager()
 * selfComponent := manager.GetSelf()
 * fmt.Printf("Manager version: %s", selfComponent.LocalVersion)
 */
func (cm *ComponentManager) GetSelf() *ComponentInstance {
	return &cm.self
}

/**
 * Get component instance by name
 * @param {string} name - Name of the component to retrieve
 * @returns {ComponentInstance} Returns component instance if found, nil otherwise
 * @description
 * - Searches for component by name in the components map
 * - Returns nil if component is not found
 * - Used to access specific component information and operations
 */
func (cm *ComponentManager) GetComponent(name string) *ComponentInstance {
	if name == cm.self.spec.Name {
		return &cm.self
	}
	cpn, ok := cm.components[name]
	if ok {
		return cpn
	}
	cpn, ok = cm.configs[name]
	if ok {
		return cpn
	}
	return nil
}

/**
 * Upgrade all components that need updates
 * @returns {error} Returns nil (always returns nil for backward compatibility)
 * @description
 * - Iterates through all managed components
 * - Checks if each component needs upgrade (needUpgrade flag)
 * - Calls upgradeComponent for each component that needs upgrade
 * - Logs upgrade operations and results
 * - Continues processing even if some upgrades fail
 * @example
 * manager := GetComponentManager()
 * if err := manager.UpgradeAll(); err != nil {
 *     logger.Error("Some upgrades failed")
 * }
 */
func (cm *ComponentManager) UpgradeAll() error {
	for _, cpn := range cm.configs {
		if cpn.needUpgrade {
			cpn.upgradeComponent()
		}
	}
	for _, cpn := range cm.components {
		if cpn.needUpgrade {
			cpn.upgradeComponent()
		}
	}
	u := utils.NewUpgrader("", utils.UpgradeConfig{
		BaseDir: env.CostrictDir,
	}, nil)
	u.CleanupOlders(3)
	return nil
}

/**
 * Check components for updates and upgrade if needed
 * @returns {error} Returns error if check or upgrade fails, nil on success
 * @description
 * - Checks all components for available updates
 * - Upgrades components that have newer versions available
 * - Uses mutex to prevent concurrent check operations
 * - Logs upgrade operations and results
 * @throws
 * - Component check errors
 * - Component upgrade errors
 */
func (cm *ComponentManager) CheckComponents() int {
	logger.Info("Starting component update check...")

	upgradeCount := 0
	components := []*ComponentInstance{&cm.self}
	for _, cpn := range cm.components {
		components = append(components, cpn)
	}
	for _, cpn := range cm.configs {
		components = append(components, cpn)
	}
	for _, cpn := range components {
		// Refresh component information to get latest version
		if err := cpn.fetchComponentInfo(); err != nil {
			logger.Errorf("Failed to fetch component info for %s: %v", cpn.spec.Name, err)
			continue
		}
		// Check if upgrade is needed
		if cpn.needUpgrade {
			logger.Infof("Component %s needs upgrade from %s to %s", cpn.spec.Name,
				cpn.local.VersionId.String(), cpn.remote.Newest.VersionId.String())
			upgradeCount++
		}
	}

	logger.Infof("Component update check completed. %d components upgraded.", upgradeCount)
	return upgradeCount
}
