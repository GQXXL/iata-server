package system

import (
	"context"
	"time"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryProbeAgentListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProbeAgentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProbeAgentListLogic {
	return &QueryProbeAgentListLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *QueryProbeAgentListLogic) QueryProbeAgentList(req *types.QueryProbeAgentListRequest) (*types.QueryProbeAgentListResponse, error) {
	total, agents, err := l.svcCtx.Store.ProbeAgent().ListAgents(l.ctx, req.Page, req.Size)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query probe agent list failed")
	}

	serverIds := make([]int64, 0, len(agents))
	for _, a := range agents {
		serverIds = append(serverIds, a.ServerId)
	}
	_, servers, err := l.svcCtx.Store.Node().FilterServerList(l.ctx, &node.FilterParams{Page: 1, Size: 10000, Ids: serverIds})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query server list failed")
	}
	serverMap := make(map[int64]*node.Server, len(servers))
	for _, s := range servers {
		serverMap[s.Id] = s
	}

	resp := &types.QueryProbeAgentListResponse{Total: total, List: make([]types.ProbeAgentItem, 0, len(agents))}
	for _, agent := range agents {
		item := types.ProbeAgentItem{ServerId: agent.ServerId, Name: agent.Name, Status: "offline", IntervalSeconds: 30, Enabled: false}
		if item.Name == "" {
			if s := serverMap[agent.ServerId]; s != nil {
				item.Name = s.Name
			}
		}
		item.Status = agent.Status
		item.Version = agent.Version
		if agent.LastSeenAt != nil {
			item.LastSeenAt = agent.LastSeenAt.UnixMilli()
			if time.Since(*agent.LastSeenAt) > 2*time.Minute {
				item.Status = "offline"
			}
		}
		if s := serverMap[agent.ServerId]; s != nil {
			item.ServerSort = s.Sort
		}
		target, _ := l.svcCtx.Store.ProbeAgent().FindTargetByServerId(l.ctx, agent.ServerId)
		if target != nil {
			item.TargetCt = target.TargetCt
			item.TargetCu = target.TargetCu
			item.TargetCm = target.TargetCm
			item.Enabled = target.Enabled
			if target.IntervalSeconds > 0 {
				item.IntervalSeconds = target.IntervalSeconds
			}
		}
		resp.List = append(resp.List, item)
	}
	return resp, nil
}
