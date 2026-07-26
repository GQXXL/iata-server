package auth

import (
	"github.com/perfect-panel/server/internal/logic/auth"
	"github.com/perfect-panel/server/internal/middleware"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Check user telephone is exist
func CheckUserTelephoneHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.TelephoneCheckUserRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}
		if err := middleware.GuardPublicRequest(
			c,
			svcCtx,
			"auth_check_phone",
			"",
			middleware.RateLimitRule{Period: 60, Quota: 60},
			middleware.RateLimitRule{Period: 3600, Quota: 500},
		); err != nil {
			result.HttpResult(c, nil, err)
			return
		}

		l := auth.NewCheckUserTelephoneLogic(c.Request.Context(), svcCtx)
		resp, err := l.CheckUserTelephone(&req)
		result.HttpResult(c, resp, err)
	}
}
