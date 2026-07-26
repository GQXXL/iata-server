package auth

import (
	"time"

	"github.com/perfect-panel/server/internal/logic/auth"
	"github.com/perfect-panel/server/internal/middleware"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
	"github.com/perfect-panel/server/pkg/turnstile"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

// User register
func UserRegisterHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.UserRegisterRequest
		_ = c.ShouldBind(&req)
		// get client ip
		req.IP = middleware.PublicClientIP(c)
		req.UserAgent = c.Request.UserAgent()
		if err := middleware.GuardPublicRequest(
			c,
			svcCtx,
			"auth_register_email",
			req.UserAgent,
			middleware.RateLimitRule{Period: 3600, Quota: 5},
			middleware.RateLimitRule{Period: 86400, Quota: 10},
		); err != nil {
			result.HttpResult(c, nil, err)
			return
		}
		if svcCtx.Config.Verify.RegisterVerify {
			verifyTurns := turnstile.New(turnstile.Config{
				Secret:  svcCtx.Config.Verify.TurnstileSecret,
				Timeout: 3 * time.Second,
			})
			if verify, err := verifyTurns.Verify(c, req.CfToken, req.IP); err != nil || !verify {
				result.HttpResult(c, nil, errors.Wrapf(xerr.NewErrCode(xerr.TooManyRequests), "verify error: %v", err.Error()))
				return
			}
		}
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}
		if err := middleware.GuardPublicKey(
			c.Request.Context(),
			svcCtx,
			"auth_register_email_target",
			req.Email,
			middleware.RateLimitRule{Period: 3600, Quota: 3},
		); err != nil {
			result.HttpResult(c, nil, err)
			return
		}

		l := auth.NewUserRegisterLogic(c.Request.Context(), svcCtx)
		resp, err := l.UserRegister(&req)
		result.HttpResult(c, resp, err)
	}
}
