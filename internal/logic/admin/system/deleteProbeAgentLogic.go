package system

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type DeleteProbeAgentLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProbeAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProbeAgentLogic {
	return &DeleteProbeAgentLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteProbeAgentLogic) DeleteProbeAgent(req *types.DeleteProbeAgentRequest) error {
	if req.ServerId <= 0 {
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "invalid server_id")
	}
	if err := l.svcCtx.Store.ProbeAgent().DeleteAgentByServerId(l.ctx, req.ServerId); err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete probe agent failed")
	}
	return nil
}
