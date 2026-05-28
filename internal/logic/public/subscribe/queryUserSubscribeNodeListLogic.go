package subscribe

import (
	"context"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/model/probeagent"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
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

	userSubscribes, err := l.svcCtx.Store.User().QueryUserSubscribe(l.ctx, u.Id, 1, 2)
	if err != nil {
		logger.Errorw("failed to query user subscribe", logger.Field("error", err.Error()), logger.Field("user_id", u.Id))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "DB_ERROR")
	}

	resp = &types.QueryUserSubscribeNodeListResponse{}
	for _, us := range userSubscribes {
		userSubscribe, err := l.getUserSubscribe(us.Token)
		if err != nil {
			l.Errorw("[SubscribeLogic] Get user subscribe failed", logger.Field("error", err.Error()), logger.Field("token", userSubscribe.Token))
			return nil, err
		}
		nodes, err := l.getServers(userSubscribe)
		if err != nil {
			return nil, err
		}
		userSubscribeInfo := types.UserSubscribeInfo{
			Id:          userSubscribe.Id,
			Nodes:       nodes,
			Traffic:     userSubscribe.Traffic,
			Upload:      userSubscribe.Upload,
			Download:    userSubscribe.Download,
			Token:       userSubscribe.Token,
			UserId:      userSubscribe.UserId,
			OrderId:     userSubscribe.OrderId,
			SubscribeId: userSubscribe.SubscribeId,
			StartTime:   userSubscribe.StartTime.Unix(),
			ExpireTime:  userSubscribe.ExpireTime.Unix(),
			Status:      userSubscribe.Status,
			CreatedAt:   userSubscribe.CreatedAt.Unix(),
			UpdatedAt:   userSubscribe.UpdatedAt.Unix(),
		}

		if userSubscribe.FinishedAt != nil {
			userSubscribeInfo.FinishedAt = userSubscribe.FinishedAt.Unix()
		}

		if l.svcCtx.Config.Register.EnableTrial && l.svcCtx.Config.Register.TrialSubscribe == userSubscribe.SubscribeId {
			userSubscribeInfo.IsTryOut = true
		}

		resp.List = append(resp.List, userSubscribeInfo)
	}

	return
}

func (l *QueryUserSubscribeNodeListLogic) getServers(userSub *user.Subscribe) (userSubscribeNodes []*types.UserSubscribeNodeInfo, err error) {
	userSubscribeNodes = make([]*types.UserSubscribeNodeInfo, 0)
	if l.isSubscriptionExpired(userSub) {
		return l.createExpiredServers(), nil
	}

	subDetails, err := l.svcCtx.Store.Subscribe().FindOne(l.ctx, userSub.SubscribeId)
	if err != nil {
		l.Errorw("[Generate Subscribe]find subscribe details error: %v", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe details error: %v", err.Error())
	}
	nodeIds := tool.StringToInt64Slice(subDetails.Nodes)
	tags := strings.Split(subDetails.NodeTags, ",")
	cleanTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}
	tags = cleanTags

	// 节点状态页改为全局节点池：不再按订阅绑定的 nodeIds / nodeTags 过滤
	// 这里直接查询全部启用节点，保证所有用户看到一致的全局节点状态。
	nodeIds = []int64{}
	tags = []string{}

	l.Debugf("[Generate Subscribe]nodes(global): %v, NodeTags(global): %v", nodeIds, tags)

	enable := true

	nodes, err := l.filterSubscribeNodes(nodeIds, tags, enable)
	if err != nil {
		l.Errorw("[Generate Subscribe]find server details error: %v", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find server details error: %v", err.Error())
	}

	if len(nodes) > 0 {
		var serverMapIds = make(map[int64]*node.Server)
		for _, n := range nodes {
			serverMapIds[n.ServerId] = nil
		}
		var serverIds []int64
		for k := range serverMapIds {
			serverIds = append(serverIds, k)
		}

		servers, err := l.svcCtx.Store.Node().QueryServerList(l.ctx, serverIds)
		if err != nil {
			l.Errorw("[Generate Subscribe]find server details error: %v", logger.Field("error", err.Error()))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find server details error: %v", err.Error())
		}

		for _, s := range servers {
			serverMapIds[s.Id] = s
		}

		ctMap, _ := l.svcCtx.Store.ProbeAgent().GetLatestResultByServerIdsAndISP(l.ctx, serverIds, "ct")
		cuMap, _ := l.svcCtx.Store.ProbeAgent().GetLatestResultByServerIdsAndISP(l.ctx, serverIds, "cu")
		cmMap, _ := l.svcCtx.Store.ProbeAgent().GetLatestResultByServerIdsAndISP(l.ctx, serverIds, "cm")
		agentMap := make(map[int64]*probeagent.Agent, len(serverIds))
		for _, sid := range serverIds {
			a, err := l.svcCtx.Store.ProbeAgent().FindAgentByServerId(l.ctx, sid)
			if err == nil && a != nil {
				agentMap[sid] = a
			}
		}
		for _, n := range nodes {
			server := serverMapIds[n.ServerId]
			if server == nil {
				continue
			}
			userSubscribeNode := &types.UserSubscribeNodeInfo{
				Id:        n.Id,
				Name:      n.Name,
				Uuid:      userSub.UUID,
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

			latestMs := int64(0)
			latestTs := time.Time{}

			if r := ctMap[n.ServerId]; r != nil {
				userSubscribeNode.CtLatencyMs = r.LatencyMs
				if !r.CheckedAt.IsZero() && r.CheckedAt.After(latestTs) {
					latestTs = r.CheckedAt
				}
			}
			if r := cuMap[n.ServerId]; r != nil {
				userSubscribeNode.CuLatencyMs = r.LatencyMs
				if !r.CheckedAt.IsZero() && r.CheckedAt.After(latestTs) {
					latestTs = r.CheckedAt
				}
			}
			if r := cmMap[n.ServerId]; r != nil {
				userSubscribeNode.CmLatencyMs = r.LatencyMs
				if !r.CheckedAt.IsZero() && r.CheckedAt.After(latestTs) {
					latestTs = r.CheckedAt
				}
			}

			if !latestTs.IsZero() {
				latestMs = latestTs.UnixMilli()
				userSubscribeNode.LatencyUpdatedAt = latestMs
			}

			if a := agentMap[n.ServerId]; a != nil && a.LastSeenAt != nil {
				if userSubscribeNode.LatencyUpdatedAt == 0 || a.LastSeenAt.UnixMilli() > userSubscribeNode.LatencyUpdatedAt {
					userSubscribeNode.LatencyUpdatedAt = a.LastSeenAt.UnixMilli()
				}
			}

			// 状态与管理端一致：按 servers.last_reported_at 判定
			if server.LastReportedAt != nil {
				d := time.Since(*server.LastReportedAt)
				if d <= 3*time.Minute {
					userSubscribeNode.Status = "online"
					userSubscribeNode.Online = true
				} else if d <= 5*time.Minute {
					userSubscribeNode.Status = "warning"
					userSubscribeNode.Online = true
				}
			}

			userSubscribeNodes = append(userSubscribeNodes, userSubscribeNode)
		}
	}

	l.Debugf("[Query Subscribe]found servers: %v", len(nodes))
	logger.Debugf("[Generate Subscribe]found servers: %v", len(nodes))
	return userSubscribeNodes, nil
}

func (l *QueryUserSubscribeNodeListLogic) filterSubscribeNodes(nodeIds []int64, tags []string, enable bool) ([]*node.Node, error) {
	addNodes := func(result []*node.Node, seen map[int64]struct{}, items []*node.Node) []*node.Node {
		for _, item := range items {
			if item == nil {
				continue
			}
			if _, ok := seen[item.Id]; ok {
				continue
			}
			seen[item.Id] = struct{}{}
			result = append(result, item)
		}
		return result
	}

	if len(nodeIds) == 0 && len(tags) == 0 {
		_, nodes, err := l.svcCtx.Store.Node().FilterNodeList(l.ctx, &node.FilterNodeParams{
			Page:    0,
			Size:    1000,
			Enabled: &enable,
		})
		return nodes, err
	}

	seen := make(map[int64]struct{})
	nodes := make([]*node.Node, 0)
	if len(nodeIds) > 0 {
		_, directNodes, err := l.svcCtx.Store.Node().FilterNodeList(l.ctx, &node.FilterNodeParams{
			Page:    0,
			Size:    1000,
			NodeId:  nodeIds,
			Enabled: &enable,
		})
		if err != nil {
			return nil, err
		}
		nodes = addNodes(nodes, seen, directNodes)
	}
	if len(tags) > 0 {
		_, tagNodes, err := l.svcCtx.Store.Node().FilterNodeList(l.ctx, &node.FilterNodeParams{
			Page:    0,
			Size:    1000,
			Tag:     tags,
			Enabled: &enable,
		})
		if err != nil {
			return nil, err
		}
		nodes = addNodes(nodes, seen, tagNodes)
	}
	return nodes, nil
}

func (l *QueryUserSubscribeNodeListLogic) isSubscriptionExpired(userSub *user.Subscribe) bool {
	return userSub.ExpireTime.Unix() < time.Now().Unix() && userSub.ExpireTime.Unix() != 0
}

func (l *QueryUserSubscribeNodeListLogic) createExpiredServers() []*types.UserSubscribeNodeInfo {
	return nil
}

func (l *QueryUserSubscribeNodeListLogic) getFirstHostLine() string {
	host := l.svcCtx.Config.Host
	lines := strings.Split(host, "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return host
}
func (l *QueryUserSubscribeNodeListLogic) getUserSubscribe(token string) (*user.Subscribe, error) {
	userSub, err := l.svcCtx.Store.User().FindOneSubscribeByToken(l.ctx, token)
	if err != nil {
		l.Infow("[Generate Subscribe]find subscribe error: %v", logger.Field("error", err.Error()), logger.Field("token", token))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe error: %v", err.Error())
	}

	//  Ignore expiration check
	//if userSub.Status > 1 {
	//	l.Infow("[Generate Subscribe]subscribe is not available", logger.Field("status", int(userSub.Status)), logger.Field("token", token))
	//	return nil, errors.Wrapf(xerr.NewErrCode(xerr.SubscribeNotAvailable), "subscribe is not available")
	//}

	return userSub, nil
}
