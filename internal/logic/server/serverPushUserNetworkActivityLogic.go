package server

import (
	"context"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/networkactivity"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/pkg/errors"
)

type ServerPushUserNetworkActivityLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewServerPushUserNetworkActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ServerPushUserNetworkActivityLogic {
	return &ServerPushUserNetworkActivityLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ServerPushUserNetworkActivityLogic) ServerPushUserNetworkActivity(req *types.ServerPushUserNetworkActivityRequest) error {
	serverInfo, err := l.svcCtx.Store.Node().FindOneServer(l.ctx, req.ServerId)
	if err != nil {
		l.Errorw("[ServerPushUserNetworkActivity] server not found", logger.Field("error", err))
		return errors.New("server not found")
	}

	rows := make([]*networkactivity.UserNetworkActivity, 0, len(req.Records))
	for _, r := range req.Records {
		domain := strings.TrimSpace(strings.ToLower(r.Domain))
		if domain == "" {
			continue
		}
		ts := time.Now()
		if r.Timestamp > 0 {
			ts = time.Unix(r.Timestamp, 0)
		}

		uid := r.UserId
		sid := r.SubscribeId
		usid := r.UserSubscribeId

		uuid := strings.TrimSpace(r.UserSubscribeUuid)
		if usid <= 0 && uuid != "" {
			id, userId, subscribeId, err := l.svcCtx.Store.NetworkActivity().FindUserSubscribeBriefByUUID(l.ctx, uuid)
			if err == nil && id > 0 {
				usid = id
				if uid <= 0 {
					uid = userId
				}
				if sid <= 0 {
					sid = subscribeId
				}
			}
		}

		rows = append(rows, &networkactivity.UserNetworkActivity{
			ServerId:        serverInfo.Id,
			UserId:          uid,
			SubscribeId:     sid,
			UserSubscribeId: usid,
			Domain:          domain,
			ClientIP:        strings.TrimSpace(r.ClientIP),
			UserAgent:       strings.TrimSpace(r.UserAgent),
			Upload:          r.Upload,
			Download:        r.Download,
			Timestamp:       ts,
		})
	}

	if err := l.svcCtx.Store.NetworkActivity().InsertBatch(l.ctx, rows); err != nil {
		l.Errorw("[ServerPushUserNetworkActivity] insert failed", logger.Field("error", err))
		return err
	}
	return nil
}
