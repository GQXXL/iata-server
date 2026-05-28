package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

func DeleteProbeAgentHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.DeleteProbeAgentRequest
		_ = c.ShouldBind(&req)
		if err := svcCtx.Validate(&req); err != nil {
			result.ParamErrorResult(c, err)
			return
		}
		l := system.NewDeleteProbeAgentLogic(c.Request.Context(), svcCtx)
		err := l.DeleteProbeAgent(&req)
		result.HttpResult(c, nil, err)
	}
}
