package auth

import (
	"github.com/perfect-panel/server/internal/logic/auth"
	"github.com/perfect-panel/server/internal/middleware"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Device Login
func DeviceLoginHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.DeviceLoginRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}
		if err := middleware.GuardPublicRequest(
			c,
			svcCtx,
			"auth_login_device",
			req.UserAgent,
			middleware.RateLimitRule{Period: 60, Quota: 3},
			middleware.RateLimitRule{Period: 3600, Quota: 10},
			middleware.RateLimitRule{Period: 86400, Quota: 20},
		); err != nil {
			result.HttpResult(c, nil, err)
			return
		}
		req.IP = middleware.PublicClientIP(c)
		if userAgent := c.Request.UserAgent(); userAgent != "" {
			req.UserAgent = userAgent
		}

		l := auth.NewDeviceLoginLogic(c.Request.Context(), svcCtx)
		resp, err := l.DeviceLogin(&req)
		result.HttpResult(c, resp, err)
	}
}
