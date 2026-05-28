package server

import (
	"github.com/perfect-panel/server/internal/logic/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

func ServerPushUserNetworkActivityHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.ServerPushUserNetworkActivityRequest
		_ = c.ShouldBind(&req)
		_ = c.ShouldBindQuery(&req.ServerCommon)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := server.NewServerPushUserNetworkActivityLogic(c.Request.Context(), svcCtx)
		err := l.ServerPushUserNetworkActivity(&req)
		result.HttpResult(c, nil, err)
	}
}
