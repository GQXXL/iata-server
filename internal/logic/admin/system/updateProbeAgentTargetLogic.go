package system

import (
	"context"

	"github.com/perfect-panel/server/internal/model/probeagent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdateProbeAgentTargetLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProbeAgentTargetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProbeAgentTargetLogic {
	return &UpdateProbeAgentTargetLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpdateProbeAgentTargetLogic) UpdateProbeAgentTarget(v *probeagent.Target, name string) error {
	v.IntervalSeconds = 10
	if _, err := l.svcCtx.Store.Node().FindOneServer(l.ctx, v.ServerId); err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "server not found")
	}
	if err := l.svcCtx.Store.ProbeAgent().UpsertTarget(l.ctx, v); err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "upsert target failed")
	}
	if name != "" {
		if err := l.svcCtx.Store.ProbeAgent().UpdateAgentNameByServerId(l.ctx, v.ServerId, name); err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "update probe agent name failed")
		}
	}
	return nil
}
