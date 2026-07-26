package common

import (
	"github.com/perfect-panel/server/internal/logic/common"
	"github.com/perfect-panel/server/internal/middleware"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get verification code
func SendEmailCodeHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.SendCodeRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}
		if err := middleware.GuardPublicRequest(
			c,
			svcCtx,
			"send_email_code",
			"",
			middleware.RateLimitRule{Period: 60, Quota: 3},
			middleware.RateLimitRule{Period: 3600, Quota: 10},
			middleware.RateLimitRule{Period: 86400, Quota: 30},
		); err != nil {
			result.HttpResult(c, nil, err)
			return
		}

		l := common.NewSendEmailCodeLogic(c.Request.Context(), svcCtx)
		resp, err := l.SendEmailCode(&req)
		result.HttpResult(c, resp, err)
	}
}
