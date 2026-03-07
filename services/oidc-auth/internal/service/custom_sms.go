package service

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zgsm-ai/oidc-auth/internal/config"
	"github.com/zgsm-ai/oidc-auth/pkg/log"
)

var (
	smsCfg *config.SMSConfig
	once   sync.Once
)

type smsResponse struct {
	TaskId    string `json:"taskId"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"`
	Count     int    `json:"count"`
}

func GetSMSCfg(cfg *config.SMSConfig) *config.SMSConfig {
	once.Do(func() {
		if cfg == nil {
			log.Error(nil, "SMS configuration is nil")
			return
		}
		smsCfg = cfg
		log.Info(nil, "SMS configuration initialized successfully")
	})
	return smsCfg
}

func getMD5(params map[string]interface{}, appKey string) string {
	excludedKeys := map[string]struct{}{
		"Signature":         {},
		"PhoneNumberSet":    {},
		"SessionContext":    {},
		"TemplateParamSet":  {},
		"SessionContextSet": {},
		"ContextParamSet":   {},
		"PhoneList":         {},
		"phoneSet":          {},
	}

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var builder strings.Builder

	for _, key := range keys {
		if _, ok := excludedKeys[key]; ok {
			continue
		}

		valueStr := fmt.Sprint(params[key])

		if valueStr == "" {
			continue
		}

		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(valueStr)
		builder.WriteString("&")
	}

	builder.WriteString("key=")
	builder.WriteString(appKey)

	signStr := builder.String()

	hasher := md5.New()
	hasher.Write([]byte(signStr))
	digest := hex.EncodeToString(hasher.Sum(nil))

	return strings.ToUpper(digest)
}

func SendSMS(phoneNum, code string) error {
	smsCfg := GetSMSCfg(nil)
	requestParams := map[string]interface{}{
		"AppId":            smsCfg.AppID,
		"MchId":            smsCfg.MchID,
		"SignType":         "MD5",
		"TemplateId":       smsCfg.TemplateID,
		"TimeStamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"Type":             "3",
		"Version":          "1.1.0",
		"TemplateParamSet": []string{code},
		"PhoneNumberSet":   []string{strings.TrimPrefix(phoneNum, "+86")},
	}
	fmt.Println(requestParams)
	calculatedSignature := getMD5(requestParams, smsCfg.ApiKey)
	requestParams["Signature"] = calculatedSignature

	jsonPayload, err := json.Marshal(requestParams)
	if err != nil {
		return fmt.Errorf("")
	}
	client := smsCfg.HTTPClient
	req, err := http.NewRequest("POST", smsCfg.SendURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var smsRes smsResponse
	err = json.Unmarshal(bodyBytes, &smsRes)
	if err != nil {
		return err
	}
	if smsRes.Status != "00" {
		return fmt.Errorf(smsRes.Message)
	}
	return nil
}
