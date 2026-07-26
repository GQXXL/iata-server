package common

import (
	"github.com/perfect-panel/server/internal/logic/common"
	"github.com/perfect-panel/server/internal/middleware"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get sms verification code
func SendSmsCodeHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.SendSmsCodeRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}
		if err := middleware.GuardPublicRequest(
			c,
			svcCtx,
			"send_sms_code",
			"",
			middleware.RateLimitRule{Period: 60, Quota: 2},
			middleware.RateLimitRule{Period: 3600, Quota: 6},
			middleware.RateLimitRule{Period: 86400, Quota: 20},
		); err != nil {
			result.HttpResult(c, nil, err)
			return
		}

		l := common.NewSendSmsCodeLogic(c.Request.Context(), svcCtx)
		resp, err := l.SendSmsCode(&req)
		result.HttpResult(c, resp, err)
	}
}
