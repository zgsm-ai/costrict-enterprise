package models

import "time"

type PortPair struct {
	LocalPort   int `json:"localPort"`   // local port
	MappingPort int `json:"mappingPort"` // mapping port to cloud
}

type TunnelDetail struct {
	Name        string        `json:"name"`        // service name
	Status      RunStatus     `json:"status"`      // tunnel status(running/stopped/error/exited)
	Pairs       []PortPair    `json:"pairs"`       // Port pairs
	CreatedTime time.Time     `json:"createdTime"` // creation time
	Pid         int           `json:"pid"`         // process ID of the tunnel
	Healthy     HealthyStatus `json:"healthy"`     // Works fine
}
