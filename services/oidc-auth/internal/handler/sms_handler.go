package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zgsm-ai/oidc-auth/internal/service"
	"github.com/zgsm-ai/oidc-auth/pkg/log"
	"github.com/zgsm-ai/oidc-auth/pkg/response"
)

type ResponseBody struct {
	Status string `json:"status"`
	Msg    string `json:"msg,omitempty"`
}

func (s *Server) SMSHandler(c *gin.Context) {
	contentType := c.ContentType()

	var phoneNumber string
	var messageContent string
	var data map[string]any

	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		strings.HasPrefix(contentType, "multipart/form-data") {

		err := c.Request.ParseForm()
		if err != nil {
			log.Error(c, "Error parsing form: %v", err)
			c.JSON(http.StatusBadRequest, ResponseBody{Status: "error", Msg: "failed to parse form data"})
			return
		}
		phoneNumber = c.PostForm("phoneNumber")
		messageContent = c.PostForm("code")
	} else {
		log.Error(c, "unsupported Content-Type: %s", contentType)
		response.JSONError(c, http.StatusUnsupportedMediaType, "", "unsupported Content-Type")
		return
	}
	if phoneNumber == "" || messageContent == "" {
		errmsg := "Missing required fields. "
		if phoneNumber == "" {
			errmsg += "Expected 'phoneNumber' (from form) or 'receiver' (from JSON). "
		}
		if messageContent == "" {
			errmsg += "Expected 'code' (from form) or 'content' (from JSON). "
		}
		log.Error(c, "Error: %s Received data: %v", errmsg, data)
		response.JSONError(c, http.StatusBadRequest, "", errmsg)
		return
	}
	SMSCfg := service.GetSMSCfg(nil)
	if SMSCfg.EnabledTest {
		log.Info(c, "processed request to send SMS to: %s, content: %s", phoneNumber, messageContent)
		log.Info(c, "simulating SMS sent to %s with content: %s", phoneNumber, messageContent)
	} else {
		err := service.SendSMS(phoneNumber, messageContent)
		if err != nil {
			log.Error(c, "failed to send SMS to %s, error: %v", phoneNumber, err)
			response.JSONError(c, http.StatusInternalServerError, "", "failed to send sms")
			return
		}
		log.Info(c, "successfully sent SMS to %s for verification", phoneNumber)
	}
	c.JSON(http.StatusOK, ResponseBody{Status: "ok", Msg: "simulated SMS sent successfully"})
}
