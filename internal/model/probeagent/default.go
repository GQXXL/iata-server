package probeagent

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Model interface {
	UpsertAgent(ctx context.Context, serverId int64, name, tokenHash, version string) error
	UpdateAgentNameByServerId(ctx context.Context, serverId int64, name string) error
	FindAgentByServerId(ctx context.Context, serverId int64) (*Agent, error)
	FindAgentByTokenHash(ctx context.Context, tokenHash string) (*Agent, error)
	TouchHeartbeat(ctx context.Context, serverId int64, version string) error
	ListAgents(ctx context.Context, page, size int) (int64, []*Agent, error)
	DeleteAgentByServerId(ctx context.Context, serverId int64) error
	UpsertTarget(ctx context.Context, data *Target) error
	FindTargetByServerId(ctx context.Context, serverId int64) (*Target, error)
	UpsertResult(ctx context.Context, data *Result) error
	GetLatestResultByServerIdsAndISP(ctx context.Context, serverIds []int64, isp string) (map[int64]*Result, error)
}

type customModel struct{ *defaultModel }
type defaultModel struct{ *gorm.DB }

func NewModel(db *gorm.DB) Model { return &customModel{defaultModel: &defaultModel{DB: db}} }

func (m *defaultModel) UpsertAgent(ctx context.Context, serverId int64, name, tokenHash, version string) error {
	now := time.Now()
	return m.WithContext(ctx).Exec(`
INSERT INTO probe_agent(server_id,name,token_hash,status,version,last_seen_at,created_at,updated_at)
VALUES(?,?,?,?,?,?,?,?)
ON DUPLICATE KEY UPDATE
name=VALUES(name),
token_hash=VALUES(token_hash),
status=IF(VALUES(version)='',status,'online'),
version=IF(VALUES(version)='',version,VALUES(version)),
last_seen_at=IF(VALUES(version)='',last_seen_at,VALUES(last_seen_at)),
updated_at=VALUES(updated_at)
`, serverId, name, tokenHash, "online", version, now, now, now).Error
}

func (m *defaultModel) UpdateAgentNameByServerId(ctx context.Context, serverId int64, name string) error {
	return m.WithContext(ctx).Model(&Agent{}).Where("server_id = ?", serverId).Updates(map[string]any{
		"name":       name,
		"updated_at": time.Now(),
	}).Error
}

func (m *defaultModel) FindAgentByServerId(ctx context.Context, serverId int64) (*Agent, error) {
	var v Agent
	err := m.WithContext(ctx).Model(&Agent{}).Where("server_id = ?", serverId).First(&v).Error
	return &v, err
}

func (m *defaultModel) FindAgentByTokenHash(ctx context.Context, tokenHash string) (*Agent, error) {
	var v Agent
	err := m.WithContext(ctx).Model(&Agent{}).Where("token_hash = ?", tokenHash).First(&v).Error
	return &v, err
}

func (m *defaultModel) TouchHeartbeat(ctx context.Context, serverId int64, version string) error {
	now := time.Now()
	updates := map[string]any{
		"status":       "online",
		"last_seen_at": now,
		"updated_at":   now,
	}
	if version != "" {
		updates["version"] = version
	}
	return m.WithContext(ctx).Model(&Agent{}).Where("server_id = ?", serverId).Updates(updates).Error
}

func (m *defaultModel) ListAgents(ctx context.Context, page, size int) (int64, []*Agent, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	var total int64
	query := m.WithContext(ctx).Model(&Agent{})
	if err := query.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var list []*Agent
	err := query.Order("updated_at DESC,id DESC").Limit(size).Offset((page - 1) * size).Find(&list).Error
	return total, list, err
}

func (m *defaultModel) DeleteAgentByServerId(ctx context.Context, serverId int64) error {
	return m.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("server_id = ?", serverId).Delete(&Result{}).Error; err != nil {
			return err
		}
		if err := tx.Where("server_id = ?", serverId).Delete(&Target{}).Error; err != nil {
			return err
		}
		if err := tx.Where("server_id = ?", serverId).Delete(&Agent{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (m *defaultModel) UpsertTarget(ctx context.Context, data *Target) error {
	now := time.Now()
	return m.WithContext(ctx).Exec(`
INSERT INTO probe_agent_target(server_id,target_ct,target_cu,target_cm,enabled,interval_seconds,created_at,updated_at)
VALUES(?,?,?,?,?,?,?,?)
ON DUPLICATE KEY UPDATE
 target_ct=VALUES(target_ct),target_cu=VALUES(target_cu),target_cm=VALUES(target_cm),enabled=VALUES(enabled),interval_seconds=VALUES(interval_seconds),updated_at=VALUES(updated_at)
`, data.ServerId, data.TargetCt, data.TargetCu, data.TargetCm, data.Enabled, data.IntervalSeconds, now, now).Error
}

func (m *defaultModel) FindTargetByServerId(ctx context.Context, serverId int64) (*Target, error) {
	var v Target
	err := m.WithContext(ctx).Model(&Target{}).Where("server_id = ?", serverId).First(&v).Error
	return &v, err
}

func (m *defaultModel) UpsertResult(ctx context.Context, data *Result) error {
	now := time.Now()
	if data.CheckedAt.IsZero() {
		data.CheckedAt = now
	}
	return m.WithContext(ctx).Exec(`
INSERT INTO probe_agent_result(server_id,isp,latency_ms,status,error_msg,probe_mode,checked_at,created_at,updated_at)
VALUES(?,?,?,?,?,?,?,?,?)
ON DUPLICATE KEY UPDATE
 latency_ms=VALUES(latency_ms),status=VALUES(status),error_msg=VALUES(error_msg),probe_mode=VALUES(probe_mode),checked_at=VALUES(checked_at),updated_at=VALUES(updated_at)
`, data.ServerId, data.ISP, data.LatencyMs, data.Status, data.ErrorMsg, data.ProbeMode, data.CheckedAt, now, now).Error
}

func (m *defaultModel) GetLatestResultByServerIdsAndISP(ctx context.Context, serverIds []int64, isp string) (map[int64]*Result, error) {
	out := map[int64]*Result{}
	if len(serverIds) == 0 {
		return out, nil
	}
	var list []*Result
	err := m.WithContext(ctx).Model(&Result{}).Where("server_id IN ? AND isp = ?", serverIds, isp).Find(&list).Error
	if err != nil {
		return nil, err
	}
	for _, it := range list {
		out[it.ServerId] = it
	}
	return out, nil
}
