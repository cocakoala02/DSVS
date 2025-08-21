package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ldsec/drynx/ginsrv/datastruct"
	"github.com/ldsec/drynx/ginsrv/drynxhub"
)

type Drynx struct {
}

func (d *Drynx) TripartiteSurvey(c *gin.Context) {
	req := &datastruct.TriSurReq{}
	if err := c.ShouldBind(req); err != nil { //bangdingshuju
		ErrorResp(http.StatusBadRequest, c, errors.New("传入数据解析错误"))
		return
	}

	resp, validres, err := drynxhub.SurveyRun(req)
	if err != nil {
		ErrorResp(http.StatusBadRequest, c, err)
		return // 很重要
	}

	SuccessResp(http.StatusOK, c, resp, validres)
}

// 解析请求参数 → 失败则返回 400。
// 调用业务逻辑 SurveyRun → 如果出错也返回 400。
// 返回统一的 JSON 响应 → 方便前端处理。
// 它结合了你之前写的 SuccessResp 和 ErrorResp，实现了一个 规范的 API 接口
