package controller

import (
	"github.com/gin-gonic/gin"
)

type Controller struct {
	Drynx
}

func ControllerSetup() *Controller {
	return &Controller{}
}

type Response struct {
	Success string `json:"success"`
	Data    any    `json:"data,omitempty"`
}

func SuccessResp(code int, c *gin.Context, obj any, valres string) {
	resp := Response{
		Success: valres,
		Data:    obj,
	}

	c.JSON(code, resp)
}

func ErrorResp(code int, c *gin.Context, err error) {
	resp := Response{
		Success: "false",
		Data:    err,
	}

	c.JSON(code, resp)
}
