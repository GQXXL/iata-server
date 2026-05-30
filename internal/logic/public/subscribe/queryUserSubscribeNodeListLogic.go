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
	enable := true

	_, nodes, err := l.svcCtx.Store.Node().FilterNodeList(l.ctx, &node.FilterNodeParams{
		Page:    0,
		Size:    1000,
		Enabled: &enable,
	})
	if err != nil {
		l.Errorw("[QueryUserSubscribeNodeList] find node list error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find node list error: %v", err.Error())
	}

	if len(nodes) == 0 {
		return userSubscribeNodes, nil
	}

	serverMap := make(map[int64]*node.Server)
	nodesByID := make(map[int64]*node.Node)
	serverIDSet := make(map[int64]struct{})
	for _, n := range nodes {
		if n == nil {
			continue
		}
		nodesByID[n.Id] = n
		serverIDSet[n.ServerId] = struct{}{}
	}

	serverIds := make([]int64, 0, len(serverIDSet))
	for sid := range serverIDSet {
		serverIds = append(serverIds, sid)
	}

	servers, err := l.svcCtx.Store.Node().QueryServerList(l.ctx, serverIds)
	if err != nil {
		l.Errorw("[QueryUserSubscribeNodeList] find server details error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find server details error: %v", err.Error())
	}
	for _, s := range servers {
		serverMap[s.Id] = s
	}

	ctMap, _ := l.svcCtx.Store.ProbeAgent().GetLatestResultByServerIdsAndISP(l.ctx, serverIds, "ct")
	cuMap, _ := l.svcCtx.Store.ProbeAgent().GetLatestResultByServerIdsAndISP(l.ctx, serverIds, "cu")
	cmMap, _ := l.svcCtx.Store.ProbeAgent().GetLatestResultByServerIdsAndISP(l.ctx, serverIds, "cm")

	agentMap := make(map[int64]*probeagent.Agent, len(serverIds))
	targetEnabledMap := make(map[int64]bool, len(serverIds))
	for _, sid := range serverIds {
		a, e := l.svcCtx.Store.ProbeAgent().FindAgentByServerId(l.ctx, sid)
		if e == nil && a != nil {
			agentMap[sid] = a
		}
		target, e := l.svcCtx.Store.ProbeAgent().FindTargetByServerId(l.ctx, sid)
		if e == nil && target != nil {
			targetEnabledMap[sid] = target.Enabled
		}
	}

	for _, n := range nodes {
		if n == nil {
			continue
		}
		server := serverMap[n.ServerId]
		if server == nil {
			continue
		}

		displayName := n.Name
		if a := agentMap[n.ServerId]; a != nil && strings.TrimSpace(a.Name) != "" {
			displayName = a.Name
		}

		item := &types.UserSubscribeNodeInfo{
			Id:        n.Id,
			Name:      displayName,
			Protocol:  n.Protocol,
			Port:      n.Port,
			Address:   n.Address,
			Tags:      strings.Split(n.Tags, ","),
			Country:   server.Country,
			City:      server.City,
			CreatedAt: n.CreatedAt.Unix(),
			Status:    "offline",
			Online:    false,
		}

		latestTs := time.Time{}
		if targetEnabledMap[n.ServerId] {
			if r := ctMap[n.ServerId]; r != nil {
				item.CtLatencyMs = r.LatencyMs
				if !r.CheckedAt.IsZero() && r.CheckedAt.After(latestTs) {
					latestTs = r.CheckedAt
				}
			}
			if r := cuMap[n.ServerId]; r != nil {
				item.CuLatencyMs = r.LatencyMs
				if !r.CheckedAt.IsZero() && r.CheckedAt.After(latestTs) {
					latestTs = r.CheckedAt
				}
			}
			if r := cmMap[n.ServerId]; r != nil {
				item.CmLatencyMs = r.LatencyMs
				if !r.CheckedAt.IsZero() && r.CheckedAt.After(latestTs) {
					latestTs = r.CheckedAt
				}
			}

			if !latestTs.IsZero() {
				item.LatencyUpdatedAt = latestTs.UnixMilli()
			}
			if a := agentMap[n.ServerId]; a != nil && a.LastSeenAt != nil {
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

	// 去重：同一节点ID仅保留1条
	uniq := make(map[int64]*types.UserSubscribeNodeInfo, len(userSubscribeNodes))
	for _, n := range userSubscribeNodes {
		if n == nil {
			continue
		}
		uniq[n.Id] = n
	}
	userSubscribeNodes = userSubscribeNodes[:0]
	for _, n := range uniq {
		userSubscribeNodes = append(userSubscribeNodes, n)
	}

	// 排序：同步维护/探测管理的服务器排序，其次节点排序
	sort.SliceStable(userSubscribeNodes, func(i, j int) bool {
		a := nodesByID[userSubscribeNodes[i].Id]
		b := nodesByID[userSubscribeNodes[j].Id]
		if a == nil || b == nil {
			return userSubscribeNodes[i].Id < userSubscribeNodes[j].Id
		}
		sa := 1 << 30
		sb := 1 << 30
		if s := serverMap[a.ServerId]; s != nil {
			sa = s.Sort
		}
		if s := serverMap[b.ServerId]; s != nil {
			sb = s.Sort
		}
		if sa != sb {
			return sa < sb
		}
		if a.Sort != b.Sort {
			return a.Sort < b.Sort
		}
		return a.Id < b.Id
	})

	return userSubscribeNodes, nil
}
