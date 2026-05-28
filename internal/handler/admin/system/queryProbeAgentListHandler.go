package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

func QueryProbeAgentListHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.QueryProbeAgentListRequest
		_ = c.ShouldBindQuery(&req)
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Size <= 0 {
			req.Size = 20
		}
		l := system.NewQueryProbeAgentListLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryProbeAgentList(&req)
		result.HttpResult(c, resp, err)
	}
}
