package config

/**
 * Service object (serialized to JSON format)
 * @property {string} name - Service name
 * @property {string} version - Service version
 * @property {bool} installed - Whether the service is installed
 * @property {string} startup - Startup mode: always/once/none
 * @property {string} status - Service status: exited/running/stopped/error
 * @property {string} protocol - Service protocol
 * @property {int} port - Service port
 * @property {string} command - Startup command
 * @property {string} metrics - Metrics endpoint path
 * @property {string} healthy - Health check endpoint path
 * @property {string} accessible - Accessible: remote/local
 */
type ServiceKnowledge struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Installed  bool   `json:"installed"`
	Startup    string `json:"startup"`
	Status     string `json:"status"`
	Protocol   string `json:"protocol,omitempty"`
	Port       int    `json:"port,omitempty"`
	Command    string `json:"command,omitempty"`
	Metrics    string `json:"metrics,omitempty"`
	Healthy    string `json:"healthy,omitempty"`
	Accessible string `json:"accessible,omitempty"`
}

/**
 * Log configuration (part of SystemKnowledge)
 * @property {string} dir - Log directory
 * @property {string} level - Log level
 */
type LogKnowledge struct {
	Dir   string `json:"dir"`
	Level string `json:"level"`
}

/**
 * SystemKnowledge structure (serialized to .well-known.json)
 * @property {LogKnowledge} logs - Log configuration
 * @property {[]ServiceKnowledge} services - Service information
 * @property {[]InterfaceInfo} interfaces - Interface information
 */
type SystemKnowledge struct {
	Logs     LogKnowledge       `json:"logs"`
	Services []ServiceKnowledge `json:"services"`
}
