package task

import (
	"encoding/json"

	"github.com/zgsm-ai/smc/internal/utils"
)

const (
	REQ_TEMPLATES      = "/taskd/api/v1/templates"
	REQ_TEMPLATES_ONCE = "/taskd/api/v1/templates/{0}"
)

/**
 * Task template metadata structure
 */
type TemplateMetadata struct {
	Name        string `json:"name,omitempty"`
	Title       string `json:"title,omitempty"`
	Engine      string `json:"engine,omitempty"`
	Schema      string `json:"schema,omitempty"`
	Extra       string `json:"extra,omitempty"`
	Description string `json:"description,omitempty"`
}

/**
 * Parameters for listing templates
 */
type ListTemplatesArgs struct {
	Name     string `json:"name,omitempty"`
	Engine   string `json:"engine,omitempty"`
	Title    string `json:"title,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	Verbose  bool   `json:"verbose,omitempty"`
}

/**
 * Result of template listing operation
 */
type ListTemplatesResult struct {
	Total int                `json:"total"`
	List  []TemplateMetadata `json:"list"`
}

/**
 * Get template by name
 * @param ss active session
 * @param name template identifier
 * @return TemplateMetadata template configuration
 * @return error if operation fails
 */
func GetTemplate(ss *utils.Session, name string) (TemplateMetadata, error) {
	var md TemplateMetadata
	rspData, err := ss.GetData(utils.ApiPath(REQ_TEMPLATES_ONCE, name), nil)
	if err != nil {
		return md, err
	}
	if err := json.Unmarshal(rspData, &md); err != nil {
		return md, err
	}

	return md, nil
}

/**
 * Add new task template
 * @param ss active session
 * @param meta template configuration to add
 * @return error if operation fails
 */
func AddTemplate(ss *utils.Session, meta TemplateMetadata) error {
	data, err := json.Marshal(&meta)
	if err != nil {
		return err
	}
	_, err = ss.Post(utils.ApiPath(REQ_TEMPLATES), data)
	return err
}

/**
 * Remove existing task template
 * @param ss active session
 * @param name template identifier
 * @return error if operation fails
 */
func RemoveTemplate(ss *utils.Session, name string) error {
	_, err := ss.Delete(utils.ApiPath(REQ_TEMPLATES_ONCE, name))
	return err
}

/**
 * List templates with filtering and pagination
 * @param ss active session
 * @param arg query parameters
 * @return []TemplateMetadata list of matching templates
 * @return error if operation fails
 */
func ListTemplates(ss *utils.Session, arg *ListTemplatesArgs) ([]TemplateMetadata, error) {
	jsonData := utils.Json{
		"engine":   arg.Engine,
		"name":     arg.Name,
		"page":     arg.Page,
		"pageSize": arg.PageSize,
		"sort":     "id DESC",
		"verbose":  arg.Verbose,
	}
	rspData, err := ss.GetData(utils.ApiPath(REQ_TEMPLATES), jsonData)
	if err != nil {
		return nil, err
	}
	var result []TemplateMetadata
	if err := json.Unmarshal(rspData, &result); err != nil {
		return nil, err
	}
	return result, nil
}
