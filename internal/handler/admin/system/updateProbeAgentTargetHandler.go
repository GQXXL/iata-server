package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/model/probeagent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

func UpdateProbeAgentTargetHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.UpdateProbeAgentTargetRequest
		_ = c.ShouldBind(&req)
		if err := svcCtx.Validate(&req); err != nil {
			result.ParamErrorResult(c, err)
			return
		}
		if req.IntervalSeconds <= 0 {
			req.IntervalSeconds = 30
		}
		l := system.NewUpdateProbeAgentTargetLogic(c.Request.Context(), svcCtx)
		err := l.UpdateProbeAgentTarget(&probeagent.Target{ServerId: req.ServerId, TargetCt: req.TargetCt, TargetCu: req.TargetCu, TargetCm: req.TargetCm, Enabled: req.Enabled, IntervalSeconds: req.IntervalSeconds}, req.Name)
		result.HttpResult(c, nil, err)
	}
}
