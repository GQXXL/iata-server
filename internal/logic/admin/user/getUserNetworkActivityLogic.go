package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/networkactivity"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetUserNetworkActivityLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserNetworkActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserNetworkActivityLogic {
	return &GetUserNetworkActivityLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetUserNetworkActivityLogic) GetUserNetworkActivity(req *types.GetUserNetworkActivityRequest) (*types.GetUserNetworkActivityResponse, error) {
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}

	list, total, err := l.svcCtx.Store.NetworkActivity().QueryPage(l.ctx, page, size, req.UserId, req.SubscribeId, req.UserSubscribeId, req.Domain, req.StartTime, req.EndTime)
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "doesn't exist") || strings.Contains(msg, "does not exist") || strings.Contains(msg, "1146") {
			return &types.GetUserNetworkActivityResponse{Total: 0, List: []types.UserNetworkActivity{}}, nil
		}
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "GetUserNetworkActivity failed: %v", err.Error())
	}

	// bucket key: user_subscribe_minute
	bucketCount := make(map[string]int64)
	for _, v := range list {
		k := bucketKey(v)
		bucketCount[k]++
	}

	deviceUACache := make(map[int64]string)
	subscribeUACache := make(map[int64]string)
	bucketTrafficCache := make(map[string][2]int64)

	respList := make([]types.UserNetworkActivity, 0, len(list))
	for _, v := range list {
		ua := strings.TrimSpace(v.UserAgent)
		if ua == "" || strings.HasPrefix(ua, "PPanel-node/journalctl") {
			if v.UserSubscribeId > 0 {
				if cached, ok := subscribeUACache[v.UserSubscribeId]; ok {
					if cached != "" {
						ua = cached
					}
				} else {
					subUA, err := l.svcCtx.Store.Log().FindLatestSubscribeUAByUserSubscribeID(l.ctx, v.UserSubscribeId)
					if err == nil {
						subscribeUACache[v.UserSubscribeId] = subUA
						if subUA != "" {
							ua = subUA
						}
					} else {
						subscribeUACache[v.UserSubscribeId] = ""
					}
				}
			}
			if ua == "" || strings.HasPrefix(ua, "PPanel-node/journalctl") {
				if cached, ok := deviceUACache[v.UserId]; ok {
					if cached != "" {
						ua = cached
					}
				} else {
					deviceUA, err := l.svcCtx.Store.NetworkActivity().FindLatestDeviceUserAgent(l.ctx, v.UserId)
					if err == nil {
						deviceUACache[v.UserId] = deviceUA
						if deviceUA != "" {
							ua = deviceUA
						}
					} else {
						deviceUACache[v.UserId] = ""
					}
				}
			}
		}

		upload := v.Upload
		download := v.Download
		if upload == 0 && download == 0 {
			k := bucketKey(v)
			tf, ok := bucketTrafficCache[k]
			if !ok {
				start := v.Timestamp.Truncate(time.Minute)
				end := start.Add(time.Minute)
				uploadSum, downloadSum, err := l.svcCtx.Store.NetworkActivity().SumTrafficLog(l.ctx, v.UserId, v.SubscribeId, start, end)
				if err == nil {
					tf = [2]int64{uploadSum, downloadSum}
				} else {
					tf = [2]int64{0, 0}
				}
				bucketTrafficCache[k] = tf
			}
			if c := bucketCount[k]; c > 0 {
				upload = tf[0] / c
				download = tf[1] / c
			}
		}

		respList = append(respList, types.UserNetworkActivity{
			Id:              v.Id,
			ServerId:        v.ServerId,
			UserId:          v.UserId,
			SubscribeId:     v.SubscribeId,
			UserSubscribeId: v.UserSubscribeId,
			Domain:          v.Domain,
			ClientIP:        v.ClientIP,
			UserAgent:       ua,
			Upload:          upload,
			Download:        download,
			Timestamp:       v.Timestamp.Unix(),
		})
	}
	return &types.GetUserNetworkActivityResponse{Total: total, List: respList}, nil
}

func bucketKey(v *networkactivity.UserNetworkActivity) string {
	return fmt.Sprintf("%d_%d_%d_%d", v.UserId, v.SubscribeId, v.UserSubscribeId, v.Timestamp.Unix()/60)
}
