package order

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryOrderDetailLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get order
func NewQueryOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOrderDetailLogic {
	return &QueryOrderDetailLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryOrderDetailLogic) QueryOrderDetail(req *types.QueryOrderDetailRequest) (resp *types.OrderDetail, err error) {
	orderInfo, err := l.svcCtx.Store.Order().FindOneDetailsByOrderNo(l.ctx, req.OrderNo)
	if err != nil {
		l.Errorw("[QueryOrderDetail] Database query error", logger.Field("error", err.Error()), logger.Field("order_no", req.OrderNo))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find order error: %v", err.Error())
	}
	if closed, e := NewCloseOrderLogic(l.ctx, l.svcCtx).CloseExpiredPendingOrder(orderInfo.OrderNo, orderInfo.Status, orderInfo.CreatedAt); e != nil {
		l.Errorw("[QueryOrderDetail] Close expired order failed", logger.Field("error", e.Error()), logger.Field("orderNo", orderInfo.OrderNo))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "close expired order failed: %v", e.Error())
	} else if closed {
		orderInfo.Status = 3
	}
	resp = &types.OrderDetail{}
	tool.DeepCopy(resp, orderInfo)
	// Prevent commission amount leakage
	resp.Commission = 0
	return
}
