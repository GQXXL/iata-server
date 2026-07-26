package middleware

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/limit"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type RateLimitRule struct {
	Period int
	Quota  int
}

func PublicClientIP(c *hertzx.Context) string {
	for _, header := range []string{
		"CF-Connecting-IP",
		"X-Real-IP",
		"X-Forwarded-For",
		"X-Original-Forwarded-For",
	} {
		value := strings.TrimSpace(c.GetHeader(header))
		if value == "" {
			continue
		}
		for _, part := range strings.Split(value, ",") {
			if ip := normalizeIP(part); ip != "" {
				return ip
			}
		}
	}

	if ip := normalizeIP(c.ClientIP()); ip != "" {
		return ip
	}
	return "unknown"
}

func GuardPublicRequest(c *hertzx.Context, svcCtx *svc.ServiceContext, action, suppliedUserAgent string, rules ...RateLimitRule) error {
	userAgent := strings.TrimSpace(c.Request.UserAgent())
	if userAgent == "" {
		userAgent = strings.TrimSpace(suppliedUserAgent)
	}
	if isSuspiciousUserAgent(userAgent) {
		return errors.Wrapf(xerr.NewErrCode(xerr.TooManyRequests), "blocked suspicious client")
	}

	return GuardPublicKey(c.Request.Context(), svcCtx, action, fmt.Sprintf("ip:%s", PublicClientIP(c)), rules...)
}

func GuardPublicKey(ctx context.Context, svcCtx *svc.ServiceContext, action, key string, rules ...RateLimitRule) error {
	key = strings.TrimSpace(strings.ToLower(key))
	if key == "" {
		key = "unknown"
	}

	for _, rule := range rules {
		if rule.Period <= 0 || rule.Quota <= 0 {
			continue
		}
		limiter := limit.NewPeriodLimit(
			rule.Period,
			rule.Quota,
			svcCtx.Redis,
			fmt.Sprintf("abuse:%s:%d:", action, rule.Period),
		)
		permit, err := limiter.TakeCtx(ctx, key)
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "take public request limit failed")
		}
		if !limiter.ParsePermitState(permit) {
			return errors.Wrapf(xerr.NewErrCode(xerr.TooManyRequests), "too many public requests")
		}
	}

	return nil
}

func normalizeIP(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	value = strings.Trim(value, "[]")
	if ip := net.ParseIP(value); ip != nil {
		return ip.String()
	}
	return ""
}

func isSuspiciousUserAgent(userAgent string) bool {
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if ua == "" || ua == "t" || ua == "test" {
		return true
	}

	for _, pattern := range []string{
		"python",
		"python-requests",
		"curl/",
		"wget/",
		"httpie",
		"aiohttp",
		"urllib",
		"scrapy",
		"go-http-client",
		"postmanruntime",
		"insomnia",
		"pentest",
		"testagent",
	} {
		if strings.Contains(ua, pattern) {
			return true
		}
	}

	return false
}
