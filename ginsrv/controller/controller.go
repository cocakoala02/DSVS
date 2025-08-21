package controller

import (
	"github.com/gin-gonic/gin"
)

type Controller struct {
	Drynx
}

// 定义了一个 Controller 类型，里面嵌入了 Drynx。
// 嵌入意味着 Controller 可以直接使用 Drynx 的方法（类似于继承）。

func ControllerSetup() *Controller {
	return &Controller{}
}

// 返回一个 Controller 的实例指针。
// 一般在程序启动时调用，用于初始化控制器。
type Response struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
	Valres  bool `json:"valres"`
}

// Response 用来统一 API 的返回格式。
// Success：字符串，表示请求是否成功（不是 bool，而是 "true" 或 "false"）。
// Data：任意类型 (any)，表示实际返回的数据。
// omitempty：如果 Data 没有值，序列化为 JSON 时会省略这个字段。

func SuccessResp(code int, c *gin.Context, obj any, valres bool) {
	resp := Response{
		Success: true,
		Data:    obj,
		Valres:  valres,
	}

	c.JSON(code, resp)
}

// 用来返回 成功的 JSON 响应。
// 参数：
// code：HTTP 状态码（如 200）。
// c：Gin 的上下文，负责请求和响应。
// obj：要返回的数据。
// valres：成功的标志字符串（通常传 "true"）。
// 内部逻辑：构造 Response → 用 c.JSON 返回。

func ErrorResp(code int, c *gin.Context, err error) {
	resp := Response{
		Success: false,
		Data:    err,
		Valres:  false,
	}

	c.JSON(code, resp)
}

// 用来返回 失败的 JSON 响应。
// 参数：
// code：HTTP 状态码（如 400/500）。
// c：Gin 的上下文。
// err：错误信息。
// 内部逻辑：构造一个 Response，Success 固定为 "false"，然后返回错误信息。
