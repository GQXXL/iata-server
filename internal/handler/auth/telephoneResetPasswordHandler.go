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

// Reset password
func TelephoneResetPasswordHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.TelephoneResetPasswordRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}
		// get client ip
		req.IP = middleware.PublicClientIP(c)
		req.UserAgent = c.Request.UserAgent()
		if err := middleware.GuardPublicRequest(
			c,
			svcCtx,
			"auth_reset_password_phone",
			req.UserAgent,
			middleware.RateLimitRule{Period: 60, Quota: 5},
			middleware.RateLimitRule{Period: 3600, Quota: 20},
		); err != nil {
			result.HttpResult(c, nil, err)
			return
		}
		if svcCtx.Config.Verify.ResetPasswordVerify {
			verifyTurns := turnstile.New(turnstile.Config{
				Secret:  svcCtx.Config.Verify.TurnstileSecret,
				Timeout: 3 * time.Second,
			})
			if verify, err := verifyTurns.Verify(c.Request.Context(), req.CfToken, req.IP); err != nil || !verify {
				err = errors.Wrapf(xerr.NewErrCode(xerr.TooManyRequests), "error: %v, verify: %v", err, verify)
				result.HttpResult(c, nil, err)
				return
			}
		}
		if err := middleware.GuardPublicKey(
			c.Request.Context(),
			svcCtx,
			"auth_reset_password_phone_target",
			req.TelephoneAreaCode+":"+req.Telephone,
			middleware.RateLimitRule{Period: 600, Quota: 5},
			middleware.RateLimitRule{Period: 3600, Quota: 15},
		); err != nil {
			result.HttpResult(c, nil, err)
			return
		}
		l := auth.NewTelephoneResetPasswordLogic(c, svcCtx)
		resp, err := l.TelephoneResetPassword(&req)
		result.HttpResult(c, resp, err)
	}
}
