package networkactivity

import (
	"context"

	"gorm.io/gorm"
)

type Model interface {
	InsertBatch(ctx context.Context, data []*UserNetworkActivity) error
	QueryPage(ctx context.Context, page, size int, userId, subscribeId, userSubscribeId int64, domain string, startTs, endTs int64) ([]*UserNetworkActivity, int64, error)
	FindLatestDeviceUserAgent(ctx context.Context, userId int64) (string, error)
	SumTrafficLog(ctx context.Context, userId, subscribeId int64, start, end interface{}) (int64, int64, error)
	FindUserSubscribeBriefByUUID(ctx context.Context, uuid string) (id, userId, subscribeId int64, err error)
}

type defaultModel struct {
	Conn *gorm.DB
}

func NewModel(db *gorm.DB) Model {
	return &defaultModel{Conn: db}
}

func (m *defaultModel) InsertBatch(ctx context.Context, data []*UserNetworkActivity) error {
	if len(data) == 0 {
		return nil
	}
	return m.Conn.WithContext(ctx).Create(&data).Error
}

func (m *defaultModel) QueryPage(ctx context.Context, page, size int, userId, subscribeId, userSubscribeId int64, domain string, startTs, endTs int64) ([]*UserNetworkActivity, int64, error) {
	q := m.Conn.WithContext(ctx).Model(&UserNetworkActivity{})
	if userId > 0 {
		q = q.Where("user_id = ?", userId)
	}
	if userSubscribeId > 0 {
		q = q.Where("user_subscribe_id = ?", userSubscribeId)
	} else if subscribeId > 0 {
		q = q.Where("subscribe_id = ?", subscribeId)
	}
	if domain != "" {
		q = q.Where("domain LIKE ?", "%"+domain+"%")
	}
	if startTs > 0 {
		q = q.Where("timestamp >= FROM_UNIXTIME(?)", startTs)
	}
	if endTs > 0 {
		q = q.Where("timestamp <= FROM_UNIXTIME(?)", endTs)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*UserNetworkActivity
	err := q.Order("timestamp DESC").Limit(size).Offset((page - 1) * size).Find(&list).Error
	return list, total, err
}

type LatestDeviceUA struct{ UserAgent string }
type TrafficBucket struct {
	Upload   int64
	Download int64
}

func (m *defaultModel) FindLatestDeviceUserAgent(ctx context.Context, userId int64) (string, error) {
	var row struct{ UserAgent string }
	err := m.Conn.WithContext(ctx).Table("user_device").Select("user_agent").Where("user_id = ? AND user_agent <> ''", userId).Order("updated_at DESC, id DESC").Limit(1).Scan(&row).Error
	return row.UserAgent, err
}

func (m *defaultModel) SumTrafficLog(ctx context.Context, userId, subscribeId int64, start, end interface{}) (int64, int64, error) {
	var row struct {
		Upload   int64
		Download int64
	}
	err := m.Conn.WithContext(ctx).Table("traffic_log").Select("COALESCE(SUM(upload),0) as upload, COALESCE(SUM(download),0) as download").Where("user_id = ? AND subscribe_id = ? AND timestamp >= ? AND timestamp < ?", userId, subscribeId, start, end).Scan(&row).Error
	return row.Upload, row.Download, err
}

func (m *defaultModel) FindUserSubscribeBriefByUUID(ctx context.Context, uuid string) (id, userId, subscribeId int64, err error) {
	var row struct {
		Id          int64
		UserId      int64
		SubscribeId int64
	}
	err = m.Conn.WithContext(ctx).Table("user_subscribe").Select("id, user_id, subscribe_id").Where("uuid = ?", uuid).Order("id DESC").Limit(1).Scan(&row).Error
	return row.Id, row.UserId, row.SubscribeId, err
}
