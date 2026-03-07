package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	Empty = ""
	AT    = "@"
	NAME  = "name"
	EMAIL = "email"
	EXT   = "email"
)

func ParseJWTUserInfo(r *http.Request, userInfoHeader string) string {

	defer func() {
		if r := recover(); r != nil {
		}
	}()

	userInfo := r.Header.Get(userInfoHeader)

	if userInfo == Empty {
		return Empty
	}

	payloadMap, err := parseJWTPayload(userInfo)
	if err != nil {
		return Empty
	}

	email, emailOk := payloadMap[EMAIL].(string)
	name, nameOk := payloadMap[NAME].(string)
	if !emailOk && !nameOk {
		ext, extOk := payloadMap[EXT].(map[string]interface{})
		if !extOk {
			return Empty
		}

		email, emailOk = ext[EMAIL].(string)
		name, nameOk = ext[NAME].(string)
	}

	atIndex := strings.Index(email, AT)

	number := Empty
	if atIndex > 0 {
		number = email[:atIndex]
	}

	username := name + number
	return username
}

func parseJWTPayload(jwtStr string) (map[string]interface{}, error) {
	decoded, err := base64.StdEncoding.DecodeString(jwtStr)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %w", err)
	}
	var payloadMap map[string]interface{}
	if err := json.Unmarshal(decoded, &payloadMap); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}
	return payloadMap, nil
}
