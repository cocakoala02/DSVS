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
	if err := c.ShouldBind(req); err != nil {
		ErrorResp(http.StatusBadRequest, c, errors.New("传入数据解析错误"))
		return
	}

	resp, validres, err := drynxhub.SurveyRun(req)
	if err != nil {
		ErrorResp(http.StatusBadRequest, c, err)
	}

	SuccessResp(http.StatusOK, c, resp, validres)
}
