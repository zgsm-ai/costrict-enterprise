package prompt

import (
	"encoding/json"
	"log"

	"github.com/zgsm-ai/smc/internal/utils"
)

const (
	REQ_PROMPT_CHAT   = "/ai-prompt-shell/api/v1/prompts/{0}/chat"
	REQ_PROMPT_RENDER = "/ai-prompt-shell/api/v1/prompts/{0}/render"
)

/**
 * call prompt chat api with given prompt and model
 * @param ss session object for API calls
 * @param promptId ID of the prompt to use
 * @param model model name to use for chat
 * @param args arguments to pass to prompt
 * @return response data and error if any
 */
func ChatWithPrompt(ss *utils.Session, promptId, model string, args map[string]interface{}) ([]byte, error) {
	jsonData := utils.Json{
		"model": model,
		"args":  args,
	}

	data, err := ss.PostJson(utils.ApiPath(REQ_PROMPT_CHAT, promptId), jsonData)
	if err != nil {
		return data, err
	}
	return data, err
}

/**
 * render prompt content with given arguments via API
 * @param ss session object for API calls
 * @param promptId ID of the prompt to render
 * @param args arguments to use for rendering
 * @return rendered content data and error if any
 */
func RenderPrompt(ss *utils.Session, promptId string, args map[string]interface{}) ([]byte, error) {
	jsonData := utils.Json{
		"args": args,
	}
	apiPath := utils.ApiPath(REQ_PROMPT_RENDER, promptId)
	body, err := json.Marshal(jsonData)
	if err != nil {
		log.Printf("RenderPrompt(%s) Marshal error: %v\n", promptId, err)
		return []byte{}, err
	}

	data, err := ss.Post(apiPath, body)
	if err != nil {
		return data, err
	}
	return data, err
}
