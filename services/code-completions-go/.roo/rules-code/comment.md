# 注释

所有注释内容，要求使用英文进行描述。

RESTful API实现函数，采用swagger注释标准进行注释，保证能生成所有API的swagger文档。

范例：

```go
// RenderPrompt render prompt template
// @Summary Render prompt template
// @Description Get prompt template by ID and render with input args
// @Tags Render
// @Accept json
// @Produce json 
// @Param prompt_id path string true "Prompt template ID"
// @Param args body string false "Template args" SchemaExample({"args":{"text":"Singleton pattern implementation"}})
// @Success 200 {object} map[string]interface{} "Rendered result"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Template not found"
// @Failure 500 {object} map[string]interface{} "Render failed"
// @Router /v1/prompts/{prompt_id}/render [post]
func (pc *PromptController) RenderPrompt(c *gin.Context) {
}
```

其它函数，采用jsdoc风格给每个函数进行注释，说明函数的功能，参数，返回值，用法，注意事项等。

范例:

```go
/**
 * Upload file to server
 * @param {string} serverPath - Target storage path on server
 * @param {*resource.AI_File} file - File object to upload, contains metadata like Size
 * @returns {error} Returns error object, nil on success
 * @description
 * - Auto handles file size: small files upload directly, large files(>DEF_PART_SIZE) call PostHugeFile
 * - Sets HTTP headers: Content-Type, Cookie and Accept
 * - Processes server response and updates progress bar
 * @throws
 * - File stream error (createFileBuffer)
 * - POST request creation failure (http.NewRequest)
 * - HTTP request error (client.Do)
 * - Non-200 status code (statusToError)
 * @example
 * err := session.PostFile("/upload/path", file)
 * if err != nil {
 *     log.Fatal(err)
 * }
 */
func (ss *AI_Session) PostFile(serverPath string, file *resource.AI_File) error {
...
}
```
