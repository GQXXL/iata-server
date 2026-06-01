package subscribe

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/model/probeagent"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryUserSubscribeNodeListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user subscribe node info
func NewQueryUserSubscribeNodeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserSubscribeNodeListLogic {
	return &QueryUserSubscribeNodeListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryUserSubscribeNodeListLogic) QueryUserSubscribeNodeList() (resp *types.QueryUserSubscribeNodeListResponse, err error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}

	// 保持鉴权语义：仍要求用户有可用订阅
	userSubscribes, err := l.svcCtx.Store.User().QueryUserSubscribe(l.ctx, u.Id, 1, 2)
	if err != nil {
		logger.Errorw("failed to query user subscribe", logger.Field("error", err.Error()), logger.Field("user_id", u.Id))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "DB_ERROR")
	}
	if len(userSubscribes) == 0 {
		return &types.QueryUserSubscribeNodeListResponse{List: []types.UserSubscribeInfo{}}, nil
	}

	nodes, err := l.getServers()
	if err != nil {
		return nil, err
	}

	resp = &types.QueryUserSubscribeNodeListResponse{
		List: []types.UserSubscribeInfo{{
			Nodes: nodes, // 全局节点，不按套餐分组
		}},
	}
	return resp, nil
}

func (l *QueryUserSubscribeNodeListLogic) getServers() (userSubscribeNodes []*types.UserSubscribeNodeInfo, err error) {
	userSubscribeNodes = make([]*types.UserSubscribeNodeInfo, 0)

	// 仅以探测管理（probe_agent）为节点来源
	_, agents, err := l.svcCtx.Store.ProbeAgent().ListAgents(l.ctx, 1, 10000)
	if err != nil {
		l.Errorw("[QueryUserSubscribeNodeList] query probe agent list error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query probe agent list error: %v", err.Error())
	}
	if len(agents) == 0 {
		return userSubscribeNodes, nil
	}

	serverIds := make([]int64, 0, len(agents))
	agentMap := make(map[int64]*probeagent.Agent, len(agents))
	for _, a := range agents {
		if a == nil {
			continue
		}
		serverIds = append(serverIds, a.ServerId)
		agentMap[a.ServerId] = a
	}
	if len(serverIds) == 0 {
		return userSubscribeNodes, nil
	}

	servers, err := l.svcCtx.Store.Node().QueryServerList(l.ctx, serverIds)
	if err != nil {
		l.Errorw("[QueryUserSubscribeNodeList] find server details error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find server details error: %v", err.Error())
	}
	serverMap := make(map[int64]*node.Server, len(servers))
	for _, s := range servers {
		serverMap[s.Id] = s
	}

	ctMap, _ := l.svcCtx.Store.ProbeAgent().GetLatestResultByServerIdsAndISP(l.ctx, serverIds, "ct")
	cuMap, _ := l.svcCtx.Store.ProbeAgent().GetLatestResultByServerIdsAndISP(l.ctx, serverIds, "cu")
	cmMap, _ := l.svcCtx.Store.ProbeAgent().GetLatestResultByServerIdsAndISP(l.ctx, serverIds, "cm")

	targetMap := make(map[int64]*probeagent.Target, len(serverIds))
	for _, sid := range serverIds {
		target, e := l.svcCtx.Store.ProbeAgent().FindTargetByServerId(l.ctx, sid)
		if e == nil && target != nil {
			targetMap[sid] = target
		}
	}

	for _, sid := range serverIds {
		a := agentMap[sid]
		if a == nil {
			continue
		}
		server := serverMap[sid]
		if server == nil {
			continue
		}

		item := &types.UserSubscribeNodeInfo{
			Id:        sid,
			Name:      strings.TrimSpace(a.Name), // 仅使用 probe_agent.name
			Country:   server.Country,
			City:      server.City,
			CreatedAt: a.CreatedAt.Unix(),
			Status:    "offline",
			Online:    false,
		}

		target := targetMap[sid]
		if target != nil {
			item.IntervalSeconds = target.IntervalSeconds
		}

		latestTs := time.Time{}
		if target != nil && target.Enabled {
			if r := ctMap[sid]; r != nil {
				item.CtLatencyMs = r.LatencyMs
				if !r.CheckedAt.IsZero() && r.CheckedAt.After(latestTs) {
					latestTs = r.CheckedAt
				}
			}
			if r := cuMap[sid]; r != nil {
				item.CuLatencyMs = r.LatencyMs
				if !r.CheckedAt.IsZero() && r.CheckedAt.After(latestTs) {
					latestTs = r.CheckedAt
				}
			}
			if r := cmMap[sid]; r != nil {
				item.CmLatencyMs = r.LatencyMs
				if !r.CheckedAt.IsZero() && r.CheckedAt.After(latestTs) {
					latestTs = r.CheckedAt
				}
			}

			if !latestTs.IsZero() {
				item.LatencyUpdatedAt = latestTs.UnixMilli()
			}
			if a.LastSeenAt != nil {
				if item.LatencyUpdatedAt == 0 || a.LastSeenAt.UnixMilli() > item.LatencyUpdatedAt {
					item.LatencyUpdatedAt = a.LastSeenAt.UnixMilli()
				}
			}
		}

		if server.LastReportedAt != nil {
			d := time.Since(*server.LastReportedAt)
			if d <= 3*time.Minute {
				item.Status = "online"
				item.Online = true
			} else if d <= 5*time.Minute {
				item.Status = "warning"
				item.Online = true
			}
		}

		userSubscribeNodes = append(userSubscribeNodes, item)
	}

	// 顺序仅同步探测管理对应服务器顺序（server.sort）
	sort.SliceStable(userSubscribeNodes, func(i, j int) bool {
		sa := 1 << 30
		sb := 1 << 30
		if s := serverMap[userSubscribeNodes[i].Id]; s != nil {
			sa = s.Sort
		}
		if s := serverMap[userSubscribeNodes[j].Id]; s != nil {
			sb = s.Sort
		}
		if sa != sb {
			return sa < sb
		}
		return userSubscribeNodes[i].Id < userSubscribeNodes[j].Id
	})

	return userSubscribeNodes, nil
}
