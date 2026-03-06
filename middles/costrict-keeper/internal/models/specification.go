package models

/**
 * Service configuration
 * @property {string} name - Service name
 * @property {string} startup - Startup mode: always/once/none
 * @property {string} command - Startup command
 * @property {string} protocol - Network protocol
 * @property {int} port - Service port
 * @property {string} metrics - Metrics endpoint path
 * @property {string} healthy - Health check endpoint path
 * @property {string} accessible - Accessible: remote/local
 */
type ServiceSpecification struct {
	Name       string   `json:"name"`
	Startup    string   `json:"startup"`
	Command    string   `json:"command,omitempty"`
	Args       []string `json:"args,omitempty"`
	Protocol   string   `json:"protocol,omitempty"`
	Port       int      `json:"port,omitempty"`
	Metrics    string   `json:"metrics,omitempty"`
	Healthy    string   `json:"healthy,omitempty"`
	Accessible string   `json:"accessible,omitempty"`
}

/**
 * Component configuration
 * @property {string} name - Component name
 * @property {string} version - Version compatibility range
 */
type ComponentSpecification struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type ManagerSpecification struct {
	Component ComponentSpecification `json:"component"`
	Service   ServiceSpecification   `json:"service"`
}

/**
 * System definition (system-spec.json)
 * @property {string} configuration - Configuration format version
 * @property {ManagerSpecification} manager - Service manager configuration
 * @property {[]ComponentSpecification} components - Component configurations
 * @property {[]ServiceSpecification} services - Service configurations
 * @property {[]ComponentSpecification} configurations - config file configurations
 */
type SystemSpecification struct {
	Configuration  string                   `json:"configuration"`
	Manager        ManagerSpecification     `json:"manager"`
	Components     []ComponentSpecification `json:"components"`
	Services       []ServiceSpecification   `json:"services"`
	Configurations []ComponentSpecification `json:"configurations,omitempty"`
}
