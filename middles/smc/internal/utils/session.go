package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/zgsm-ai/smc/internal/env"
)

/**
 * HTTP Session
 */
type Session struct {
	HostUrl string `json:"url"`
	client  *http.Client
}

type Json = map[string]interface{}
type KeyValues = map[string]string

type RespData struct {
	Code    string      `json:"code"`
	Message string      `json:"message,omitempty"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

/**
 * Build API path from template and path parameters
 */
func ApiPath(pattern string, paras ...string) string {
	str := pattern
	for i, para := range paras {
		pos := fmt.Sprintf("{%d}", i)
		str = strings.ReplaceAll(str, pos, para)
	}
	return str
}

/**
 * Convert time duration string to seconds
 */
func Time2Sec(time string) (uint32, error) {
	suffix := time[len(time)-1:]
	digits := time[0 : len(time)-1]
	val, err := strconv.Atoi(digits)
	if err != nil {
		return 0, err
	}
	if suffix == "s" {
		return uint32(val), nil
	} else if suffix == "m" {
		return uint32(val * 60), nil
	} else if suffix == "h" {
		return uint32(val * 3600), nil
	} else if suffix == "d" {
		return uint32(val * 3600 * 24), nil
	} else {
		return 0, os.ErrInvalid
	}
}

/**
 * Get error message from server response
 */
func acquireServerError(rsp *http.Response, rspBody []byte) error {
	var errTag string
	if rsp.StatusCode >= 400 && rsp.StatusCode < 500 {
		errTag = "Client Error"
	} else if rsp.StatusCode >= 500 && rsp.StatusCode < 600 {
		errTag = "Server Error"
	} else {
		errTag = "Other Error"
	}
	log.Printf("%s(%d): %s %s, response: %s\n",
		errTag, rsp.StatusCode, rsp.Request.Method, rsp.Request.URL.Path, string(rspBody))
	rspJson := &RespData{}
	if err := json.Unmarshal(rspBody, rspJson); err != nil {
		return fmt.Errorf("%s(%d): Server response: %s", errTag, rsp.StatusCode, string(rspBody))
	}
	return fmt.Errorf("%s(%d): %s", errTag, rsp.StatusCode, rspJson.Message)
}

/**
 * Create new Session for communicating with AI platform
 */
func NewSession(url string) *Session {
	ss := new(Session)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: env.SkipSSL},
	}
	ss.client = &http.Client{Transport: tr}
	ss.HostUrl = url
	return ss
}

/**
 * Create new Session using proxy
 */
func NewProxySession(hostUrl, proxy string) *Session {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		log.Println("Failed to parse proxy URL:", err)
		return nil
	}
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL), // Set proxy URL
		TLSClientConfig: &tls.Config{InsecureSkipVerify: env.SkipSSL},
	}
	ss := new(Session)
	ss.client = &http.Client{Transport: tr}
	ss.HostUrl = hostUrl
	return ss
}

/**
 * Build complete URL from session HOST, API relative path and parameters
 */
func (ss *Session) mkUrl(apiPath string, paras ...string) string {
	str := apiPath
	for i, para := range paras {
		pos := fmt.Sprintf("{%d}", i)
		str = strings.ReplaceAll(str, pos, para)
	}
	return ss.HostUrl + str
}

/**
 * Serialize objects to byte stream
 */
func transformInterfaceIntoByte(value interface{}) []byte {
	k := reflect.TypeOf(value).Kind()
	switch k {
	case reflect.String:
		return []byte(value.(string))
	case reflect.Bool:
		return []byte(fmt.Sprintf("%v", value.(bool)))
	case reflect.Int32:
		return []byte(fmt.Sprintf("%v", value.(int32)))
	case reflect.Int64:
		return []byte(fmt.Sprintf("%v", value.(int64)))
	case reflect.Int:
		return []byte(fmt.Sprintf("%v", value.(int)))
	case reflect.Uint:
		return []byte(fmt.Sprintf("%v", value.(uint)))
	case reflect.Uint32:
		return []byte(fmt.Sprintf("%v", value.(uint32)))
	case reflect.Uint64:
		return []byte(fmt.Sprintf("%v", value.(uint64)))
	case reflect.Float32:
		return []byte(fmt.Sprintf("%v", value.(float32)))
	case reflect.Float64:
		return []byte(fmt.Sprintf("%v", value.(float64)))
	default:
		log.Printf("unknown type: %s\n", reflect.TypeOf(value).Name())
	}
	return nil
}

/**
 * Build complete URL from session HOST, API relative path and parameters
 */
func (ss *Session) mkUrlByKvs(apiPath string, kvs map[string]string) string {
	str := apiPath
	for k, v := range kvs {
		placeholder := fmt.Sprintf("{%s}", k)
		str = strings.ReplaceAll(str, placeholder, url.PathEscape(v))
	}
	return ss.HostUrl + str
}

func (ss *Session) Request(method, apiPath string, paths, querys, headers map[string]string, body []byte) ([]byte, error) {
	var rd io.Reader
	if len(body) > 0 {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, ss.mkUrlByKvs(apiPath, paths), rd)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	// req.Header.Add("Content-Type", "application/json")
	if querys != nil {
		values := make(url.Values)
		for k, v := range querys {
			values.Set(k, string(transformInterfaceIntoByte(v)))
		}
		req.URL.RawQuery = values.Encode()
	}
	rsp, err := ss.client.Do(req)
	if err != nil {
		log.Printf("%s %s, query: %s, error: %v\n", method, apiPath, req.URL.RawQuery, err)
		return nil, err
	}
	defer rsp.Body.Close()
	rspBody, err := io.ReadAll(rsp.Body)
	env.LogDbg.Printf("%s %s, response %d: %s\n",
		method, req.URL.String(), rsp.StatusCode, string(rspBody))
	if rsp.StatusCode != 200 {
		return rspBody, acquireServerError(rsp, rspBody)
	}
	return rspBody, err
}

/**
 * Send GET request
 */
func (ss *Session) Get(apiPath string, params Json) ([]byte, error) {
	req, err := http.NewRequest("GET", ss.mkUrl(apiPath), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	if params != nil {
		values := make(url.Values)
		for k, v := range params {
			values.Set(k, string(transformInterfaceIntoByte(v)))
		}
		req.URL.RawQuery = values.Encode()
	}
	rsp, err := ss.client.Do(req)
	if err != nil {
		log.Printf("GET %s, query: %s, error: %v\n", apiPath, req.URL.RawQuery, err)
		return nil, err
	}
	defer rsp.Body.Close()
	rspBody, err := io.ReadAll(rsp.Body)
	env.LogDbg.Printf("GET %s, response %d: %s\n",
		req.URL.String(), rsp.StatusCode, string(rspBody))
	if rsp.StatusCode != 200 {
		return rspBody, acquireServerError(rsp, rspBody)
	}
	return rspBody, err
}

/**
 * Send raw POST request
 */
func (ss *Session) Post(apiPath string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", ss.mkUrl(apiPath), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	// req.Header.Add("Cookie", ss.Cookie)
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	rsp, err := ss.client.Do(req)
	if err != nil {
		log.Printf("POST %s, body: %s, error: %v\n", apiPath, string(body), err)
		return nil, err
	}
	defer rsp.Body.Close()
	rspBody, err := io.ReadAll(rsp.Body)
	env.LogDbg.Printf("POST %s, body: %s, status code: %d, response: %s\n",
		apiPath, string(body), rsp.StatusCode, string(rspBody))
	if rsp.StatusCode != 200 {
		return rspBody, acquireServerError(rsp, rspBody)
	}
	return rspBody, err
}

/**
 * Send raw PUT request
 */
func (ss *Session) Put(apiPath string, data Json) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", ss.mkUrl(apiPath), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	// req.Header.Add("Cookie", ss.Cookie)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	rsp, err := ss.client.Do(req)
	if err != nil {
		log.Printf("PUT %s, data: %s, error: %v\n", apiPath, string(body), err)
		return nil, err
	}
	defer rsp.Body.Close()
	rspBody, err := io.ReadAll(rsp.Body)
	env.LogDbg.Printf("PUT %s, body: %s, status code: %d, response: %s\n",
		apiPath, string(body), rsp.StatusCode, string(rspBody))
	if rsp.StatusCode != 200 {
		return rspBody, acquireServerError(rsp, rspBody)
	}
	return rspBody, err
}

/**
 * Send raw DELETE request
 */
func (ss *Session) Delete(apiPath string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", ss.mkUrl(apiPath), nil)
	if err != nil {
		return nil, err
	}
	// req.Header.Add("Cookie", ss.Cookie)
	req.Header.Add("Accept", "*/*")
	rsp, err := ss.client.Do(req)
	if err != nil {
		log.Printf("DELETE %s, error: %v\n", apiPath, err)
		return nil, err
	}
	defer rsp.Body.Close()
	rspBody, err := io.ReadAll(rsp.Body)
	env.LogDbg.Printf("DELETE %s, status code: %d, response: %s\n",
		apiPath, rsp.StatusCode, string(rspBody))
	if rsp.StatusCode != 200 {
		return rspBody, acquireServerError(rsp, rspBody)
	}
	return rspBody, err
}

/**
 * Send raw DELETE request
 */
func (ss *Session) DeleteJson(apiPath string, data Json) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("PostJson(%s, ...) Marshal error: %v\n", apiPath, err)
		return []byte{}, err
	}
	req, err := http.NewRequest("DELETE", ss.mkUrl(apiPath), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	rsp, err := ss.client.Do(req)
	if err != nil {
		log.Printf("DELETE %s, error: %v\n", apiPath, err)
		return nil, err
	}
	defer rsp.Body.Close()
	rspBody, err := io.ReadAll(rsp.Body)
	env.LogDbg.Printf("DELETE %s, status code: %d, response: %s\n",
		apiPath, rsp.StatusCode, string(rspBody))
	if rsp.StatusCode != 200 {
		return rspBody, acquireServerError(rsp, rspBody)
	}
	return rspBody, err
}

/**
 * POST data to AIP
 */
func (ss *Session) PostJson(apiPath string, jsonData Json) ([]byte, error) {
	body, err := json.Marshal(jsonData)
	if err != nil {
		log.Printf("PostJson(%s, ...) Marshal error: %v\n", apiPath, err)
		return []byte{}, err
	}
	rspBody, err := ss.Post(apiPath, body)
	if err != nil {
		return rspBody, err
	}
	rspJson := RespData{}
	if err = json.Unmarshal(rspBody, &rspJson); err != nil {
		return rspBody, err
	}
	data, err := json.Marshal(rspJson.Data)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

/**
 * Get JSON data from AIP
 */
func (ss *Session) GetData(apiPath string, params Json) ([]byte, error) {
	rspBody, err := ss.Get(apiPath, params)
	if err != nil {
		return nil, err
	}
	rspJson := RespData{}
	if err = json.Unmarshal(rspBody, &rspJson); err != nil {
		log.Printf("GetData %s decode error: %v\n", apiPath, err)
		return nil, err
	}
	if !rspJson.Success {
		log.Printf("GetData %s failed, message: %s\n", apiPath, rspJson.Message)
		return nil, fmt.Errorf("Server response: %s", rspJson.Message)
	}
	if rspJson.Data == nil {
		log.Printf("GetData %s missing '.data' field\n", apiPath)
		return nil, fmt.Errorf("Server response missing '.data' field")
	}
	return json.MarshalIndent(rspJson.Data, "", "  ")
}

/**
 * Send GET request
 */
func (ss *Session) GetBody(apiPath string, params Json) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", ss.mkUrl(apiPath), nil)
	if err != nil {
		return nil, err
	}
	// req.Header.Add("Cookie", ss.Cookie)
	req.Header.Add("Content-Type", "application/json")
	if params != nil {
		vals := make(url.Values)
		for k, v := range params {
			vals.Set(k, string(transformInterfaceIntoByte(v)))
		}
		req.URL.RawQuery = vals.Encode()
	}
	// Set long polling timeout
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Keep-Alive", "timeout=0")

	rsp, err := ss.client.Do(req)
	if err != nil {
		log.Printf("GET %s, query: %s, error: %v\n", apiPath, req.URL.RawQuery, err)
		return nil, err
	}
	env.LogDbg.Printf("GET %s, response %d, header: %+v\n",
		req.URL.String(), rsp.StatusCode, req.Header)
	return rsp.Body, err
}
