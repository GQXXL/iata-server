package system

import (
	"context"

	"github.com/perfect-panel/server/internal/model/probeagent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GenerateProbeAgentTokenLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateProbeAgentTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateProbeAgentTokenLogic {
	return &GenerateProbeAgentTokenLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GenerateProbeAgentTokenLogic) GenerateProbeAgentToken(req *types.GenerateProbeAgentTokenRequest, tokenHash, rawToken string) (*types.GenerateProbeAgentTokenResponse, error) {
	if _, err := l.svcCtx.Store.Node().FindOneServer(l.ctx, req.ServerId); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "server not found")
	}
	if err := l.svcCtx.Store.ProbeAgent().UpsertAgent(l.ctx, req.ServerId, req.Name, tokenHash, ""); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "upsert probe agent failed")
	}
	// 默认预设三网目标；间隔固定 10 秒
	if err := l.svcCtx.Store.ProbeAgent().UpsertTarget(l.ctx, &probeagent.Target{
		ServerId:        req.ServerId,
		TargetCt:        "gd-ct-v4.ip.zstaticcdn.com:443",
		TargetCu:        "gd-cu-v4.ip.zstaticcdn.com:443",
		TargetCm:        "gd-cm-v4.ip.zstaticcdn.com:443",
		Enabled:         true,
		IntervalSeconds: 10,
	}); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "init probe target failed")
	}
	return &types.GenerateProbeAgentTokenResponse{ServerId: req.ServerId, Token: rawToken}, nil
}
