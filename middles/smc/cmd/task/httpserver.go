/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package task

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zgsm-ai/smc/internal/env"
)

var CallbackChan chan TaskFinishedCallback = make(chan TaskFinishedCallback)

type ResponseData struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func respOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, ResponseData{
		Code:    "0",
		Message: "OK",
		Success: true,
		Data:    data,
	})
}

func respError(c *gin.Context, code int, err error) {
	c.JSON(code, ResponseData{
		Code:    strconv.Itoa(code),
		Message: err.Error(),
		Success: false,
		Data:    nil,
	})
}

type TaskFinishedCallback struct {
	Name    string `json:"name"`
	Uuid    string `json:"uuid"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func doCallback(c *gin.Context) {
	var req TaskFinishedCallback
	if err := c.ShouldBindJSON(&req); err != nil {
		respError(c, http.StatusBadRequest, err)
		return
	}
	respOK(c, "OK")
	CallbackChan <- req
}

/*
 * Start HTTP server and register routes
 */
func RunHttpServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	//	Task related routes
	r.POST("/callback", doCallback)
	err := r.Run(env.Listen)
	if err != nil {
		log.Fatal(err)
	}
}
