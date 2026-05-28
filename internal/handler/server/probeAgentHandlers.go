package server

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/probeagent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

func tokenSha256(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

func ProbeAgentHeartbeatHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.ProbeAgentHeartbeatRequest
		_ = c.ShouldBind(&req)
		if err := svcCtx.Validate(&req); err != nil {
			result.ParamErrorResult(c, err)
			return
		}
		h := tokenSha256(strings.TrimSpace(req.Token))
		agent, err := svcCtx.Store.ProbeAgent().FindAgentByTokenHash(c.Request.Context(), h)
		if err != nil {
			result.HttpResult(c, nil, err)
			return
		}
		_ = svcCtx.Store.ProbeAgent().TouchHeartbeat(c.Request.Context(), agent.ServerId, req.Version)
		result.HttpResult(c, map[string]any{"ok": true, "server_id": agent.ServerId, "ts": time.Now().UnixMilli()}, nil)
	}
}

func ProbeAgentGetConfigHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.ProbeAgentGetConfigRequest
		_ = c.ShouldBindQuery(&req)
		if err := svcCtx.Validate(&req); err != nil {
			result.ParamErrorResult(c, err)
			return
		}
		h := tokenSha256(strings.TrimSpace(req.Token))
		agent, err := svcCtx.Store.ProbeAgent().FindAgentByTokenHash(c.Request.Context(), h)
		if err != nil {
			result.HttpResult(c, nil, err)
			return
		}
		t, _ := svcCtx.Store.ProbeAgent().FindTargetByServerId(c.Request.Context(), agent.ServerId)
		if t == nil {
			t = &probeagent.Target{ServerId: agent.ServerId, Enabled: false, IntervalSeconds: 10}
		} else {
			t.IntervalSeconds = 10
		}
		result.HttpResult(c, &types.ProbeAgentGetConfigResponse{ServerId: agent.ServerId, TargetCt: t.TargetCt, TargetCu: t.TargetCu, TargetCm: t.TargetCm, Enabled: t.Enabled, IntervalSeconds: 10}, nil)
	}
}

func ProbeAgentReportHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.ProbeAgentReportRequest
		_ = c.ShouldBind(&req)
		if err := svcCtx.Validate(&req); err != nil {
			result.ParamErrorResult(c, err)
			return
		}
		h := tokenSha256(strings.TrimSpace(req.Token))
		agent, err := svcCtx.Store.ProbeAgent().FindAgentByTokenHash(c.Request.Context(), h)
		if err != nil {
			result.HttpResult(c, nil, err)
			return
		}
		for _, it := range req.Results {
			isp := strings.ToLower(strings.TrimSpace(it.ISP))
			if isp != "ct" && isp != "cu" && isp != "cm" {
				continue
			}
			status := strings.TrimSpace(it.Status)
			if status == "" {
				status = "ok"
			}
			mode := strings.TrimSpace(it.ProbeMode)
			if mode == "" {
				mode = "tcp"
			}
			_ = svcCtx.Store.ProbeAgent().UpsertResult(c.Request.Context(), &probeagent.Result{ServerId: agent.ServerId, ISP: isp, LatencyMs: it.LatencyMs, Status: status, ErrorMsg: it.ErrorMsg, ProbeMode: mode, CheckedAt: time.Now()})
		}
		_ = svcCtx.Store.ProbeAgent().TouchHeartbeat(c.Request.Context(), agent.ServerId, req.Version)
		result.HttpResult(c, map[string]any{"ok": true}, nil)
	}
}
