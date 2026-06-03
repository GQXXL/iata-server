package notify

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/perfect-panel/server/internal/logic/notify"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/payment"
	"github.com/perfect-panel/server/pkg/result"
)

// PaymentNotifyHandler Payment Notify
func PaymentNotifyHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		platform, ok := c.Request.Context().Value(constant.CtxKeyPlatform).(string)
		if !ok {
			logger.WithContext(c.Request.Context()).Errorf("platform not found")
			result.HttpResult(c, nil, fmt.Errorf("platform not found"))
			return
		}

		switch payment.ParsePlatform(platform) {
		case payment.EPay, payment.CryptoSaaS:
			if err := c.Request.ParseForm(); err != nil {
				logger.WithContext(c.Request.Context()).Errorw("[PaymentNotifyHandler] ParseForm failed", logger.Field("error", err.Error()))
			}

			params := make(map[string]string)
			// POST form args (Hertz wrapper check)
			if c.Request.PostForm != nil {
				for key, values := range c.Request.PostForm {
					if len(values) > 0 {
						params[key] = values[0]
					}
				}
			}
			// GET query args
			if c.Request.URL != nil {
				for key, values := range c.Request.URL.Query() {
					if len(values) > 0 {
						params[key] = values[0]
					}
				}
			}

			pid, err := strconv.ParseInt(params["pid"], 10, 64)
			if err != nil {
				logger.WithContext(c.Request.Context()).Errorw("[PaymentNotifyHandler] Parse pid failed", logger.Field("error", err.Error()), logger.Field("pid", params["pid"]))
			}
			req := &types.EPayNotifyRequest{
				Pid:         pid,
				TradeNo:     params["trade_no"],
				OutTradeNo:  params["out_trade_no"],
				Type:        params["type"],
				Name:        params["name"],
				Money:       params["money"],
				TradeStatus: params["trade_status"],
				Param:       params["param"],
				Sign:        params["sign"],
				SignType:    params["sign_type"],
			}

			l := notify.NewEPayNotifyLogic(c.Request.Context(), svcCtx, notify.EPayNotifyMeta{
				Method: c.Request.Method,
				Params: params,
			})
			if err := l.EPayNotify(req); err != nil {
				logger.WithContext(c.Request.Context()).Errorf("EPayNotify failed: %v", err.Error())
				c.String(http.StatusBadRequest, "%v", err)
				return
			}
			c.String(http.StatusOK, "%s", "success")
		case payment.Stripe:
			l := notify.NewStripeNotifyLogic(c.Request.Context(), svcCtx)
			if err := l.StripeNotify(c.Request, c.Writer); err != nil {
				result.HttpResult(c, nil, err)
				return
			}
			result.HttpResult(c, nil, nil)

		case payment.AlipayF2F:
			l := notify.NewAlipayNotifyLogic(c.Request.Context(), svcCtx)
			if err := l.AlipayNotify(c.Request); err != nil {
				result.HttpResult(c, nil, err)
				return
			}
			// Return success to alipay
			c.String(http.StatusOK, "%s", "success")

		default:
			logger.WithContext(c.Request.Context()).Errorf("platform %s not support", platform)
		}
	}
}

func formValues(values url.Values) map[string]string {
	params := make(map[string]string)
	for key, value := range values {
		if len(value) > 0 {
			params[key] = value[0]
		}
	}
	return params
}
