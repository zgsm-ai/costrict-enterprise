package models

type ServiceDetail struct {
	Name      string               `json:"name"`
	Pid       int                  `json:"pid"`
	Port      int                  `json:"port"`
	Status    RunStatus            `json:"status"`
	StartTime string               `json:"startTime"`
	Healthy   HealthyStatus        `json:"healthy"`
	Spec      ServiceSpecification `json:"spec"`
	Process   ProcessDetail        `json:"process,omitempty"`
	Tunnel    *TunnelDetail        `json:"tunnel,omitempty"`
	Component *ComponentDetail     `json:"component,omitempty"`
}
