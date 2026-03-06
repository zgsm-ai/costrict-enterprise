package models

import (
	"time"
)

type MidnightRoosterState struct {
	Status        string    `json:"status" example:"active" description:"检查状态"`
	NextCheckTime time.Time `json:"nextCheckTime" example:"2024-01-02T03:30:00Z" description:"下次检查时间"`
	LastCheckTime time.Time `json:"lastCheckTime" example:"2024-01-01T03:30:00Z" description:"最后检查时间"`
}

type PortAllocState struct {
	Min       int
	Max       int
	Allocates []int
}

type EnvConfig struct {
	Daemon      bool   `json:"deamon"`
	ListenPort  int    `json:"listenPort"`
	Version     string `json:"version"`
	CostrictDir string `json:"costrictDir"`
}

type ServerConfig struct {
	SystemSpec string `json:"systemSpec"`
	Auth       string `json:"auth"`
	Software   string `json:"software"`
	Cloud      string `json:"cloud"`
}

type ServerState struct {
	StartTime       time.Time            `json:"startTime"`
	MidnightRooster MidnightRoosterState `json:"midnightRooster"`
	PortAlloc       PortAllocState       `json:"portAlloc"`
	Env             EnvConfig            `json:"env"`
	Config          ServerConfig         `json:"config"`
}
