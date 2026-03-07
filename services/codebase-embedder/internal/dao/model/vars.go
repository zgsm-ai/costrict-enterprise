package model

type PublishStatus string

var (
	PublishStatusPending PublishStatus = "pending"
	PublishStatusSuccess PublishStatus = "success"
	PublishStatusFailed  PublishStatus = "failed"
)

type CodebaseStatus string

var (
	CodebaseStatusActive  CodebaseStatus = "active"
	CodebaseStatusExpired CodebaseStatus = "expired"
)
