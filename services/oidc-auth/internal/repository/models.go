package repository

import (
	"time"

	"github.com/google/uuid"
)

// StarUser URecord the user who logged in to GitHub and starred
type StarUser struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string `gorm:"uniqueIndex;not null;size:100" json:"name"`
	GitHubID   string `gorm:"uniqueIndex;not null" json:"github_id"`
	GitHubName string `gorm:"size:100;not null" json:"github_name"`
	StarredAt  string `gorm:"type:timestamp" json:"starred_at"`
}

// SyncLock Resolving multi-instance conflicts
type SyncLock struct {
	Name     string    `gorm:"primaryKey; unique; not null"`
	LockedAt time.Time `gorm:"type:timestamptz;" json:"locked_at"`
}

type AuthUser struct {
	ID               uuid.UUID  `gorm:"type:uuid; primaryKey" json:"id"`
	CreatedAt        time.Time  `gorm:"type:timestamptz" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"type:timestamptz" json:"updated_at"`
	Name             string     `gorm:"index; size:100" json:"name"`
	GithubID         string     `gorm:"size:100;" json:"github_id"`
	GithubName       string     `gorm:"size:100" json:"github_name"`
	Vip              int        `gorm:"default:0" json:"vip"`
	Phone            string     `gorm:"size:20" json:"phone"`
	Email            string     `gorm:"size:100;index" json:"email"`
	Password         string     `gorm:"size:100" json:"password"`
	Company          string     `gorm:"size:100" json:"company"`
	Location         string     `gorm:"size:100" json:"location"`
	UserCode         string     `gorm:"size:100" json:"user_code"`
	ExternalAccounts uuid.UUID  `gorm:"size:100" json:"external_accounts"`
	EmployeeNumber   string     `gorm:"size:100" json:"employee_number"`
	GithubStar       string     `gorm:"type:text" json:"github_star"`
	Devices          []Device   `gorm:"type:jsonb;serializer:json" json:"devices"`
	AccessTime       time.Time  `gorm:"type:timestamptz" json:"access_time"`
	InviteCode       string     `gorm:"size:10;index" json:"invite_code"`
	InviterID        *uuid.UUID `gorm:"type:uuid" json:"inviter_id"`
}

type Device struct {
	ID               uuid.UUID `json:"id"`
	CreatedAt        time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt        time.Time `gorm:"type:timestamp" json:"updated_at"`
	MachineCode      string    `json:"machine_code"`
	VSCodeVersion    string    `json:"vscode_version"`
	PluginVersion    string    `json:"plugin_version"`
	State            string    `json:"state"`
	RefreshTokenHash string    `json:"refresh_token_hash"`
	RefreshToken     string    `json:"refresh_token"`
	AccessToken      string    `json:"access_token"`
	AccessTokenHash  string    `json:"access_token_hash"`
	UriScheme        string    `json:"uri_scheme"`
	Status           string    `json:"status"`
	Provider         string    `json:"provider"`
	Platform         string    `json:"platform"`
	DeviceCode       string    `json:"device_code"`
	TokenProvider    string    `gorm:"size:20" json:"token_provider"`
}
