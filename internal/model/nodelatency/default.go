package nodelatency

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Model interface {
	CreateTask(ctx context.Context, t *MonitorTask) error
	UpdateTask(ctx context.Context, t *MonitorTask) error
	DeleteTask(ctx context.Context, id int64) error
	FindTask(ctx context.Context, id int64) (*MonitorTask, error)
	ListTask(ctx context.Context, page, size int, search string) ([]*MonitorTask, int64, error)
	ListEnabledDueTasks(ctx context.Context, now time.Time) ([]*MonitorTask, error)
	TouchTaskRunAt(ctx context.Context, id int64, now time.Time) error
	UpsertResult(ctx context.Context, r *MonitorResult) error
	GetLatestResultByNodeIdsAndISP(ctx context.Context, nodeIds []int64, isp string) (map[int64]*MonitorResult, error)
}

type defaultModel struct{ DB *gorm.DB }

func NewModel(db *gorm.DB) Model { return &defaultModel{DB: db} }

func (m *defaultModel) CreateTask(ctx context.Context, t *MonitorTask) error {
	return m.DB.WithContext(ctx).Create(t).Error
}
func (m *defaultModel) UpdateTask(ctx context.Context, t *MonitorTask) error {
	return m.DB.WithContext(ctx).Model(&MonitorTask{}).Where("id = ?", t.Id).Updates(map[string]any{
		"name": t.Name, "monitor_type": t.MonitorType, "target": t.Target,
		"target_ct": t.TargetCt, "target_cu": t.TargetCu, "target_cm": t.TargetCm, "node_ids": t.NodeIds,
		"interval_seconds": t.IntervalSeconds, "enabled": t.Enabled,
	}).Error
}
func (m *defaultModel) DeleteTask(ctx context.Context, id int64) error {
	return m.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ?", id).Delete(&MonitorResult{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&MonitorTask{}).Error
	})
}
func (m *defaultModel) FindTask(ctx context.Context, id int64) (*MonitorTask, error) {
	var t MonitorTask
	err := m.DB.WithContext(ctx).Where("id = ?", id).First(&t).Error
	return &t, err
}
func (m *defaultModel) ListTask(ctx context.Context, page, size int, search string) ([]*MonitorTask, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	q := m.DB.WithContext(ctx).Model(&MonitorTask{})
	if search != "" {
		q = q.Where("name LIKE ?", "%"+search+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*MonitorTask
	err := q.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&list).Error
	return list, total, err
}
func (m *defaultModel) ListEnabledDueTasks(ctx context.Context, now time.Time) ([]*MonitorTask, error) {
	var list []*MonitorTask
	err := m.DB.WithContext(ctx).Where("enabled = 1 AND (last_run_at IS NULL OR TIMESTAMPDIFF(SECOND,last_run_at,?) >= interval_seconds)", now).Find(&list).Error
	return list, err
}
func (m *defaultModel) TouchTaskRunAt(ctx context.Context, id int64, now time.Time) error {
	return m.DB.WithContext(ctx).Model(&MonitorTask{}).Where("id = ?", id).Update("last_run_at", now).Error
}
func (m *defaultModel) UpsertResult(ctx context.Context, r *MonitorResult) error {
	return m.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var old MonitorResult
		err := tx.Where("task_id = ? AND node_id = ? AND isp = ?", r.TaskId, r.NodeId, r.ISP).First(&old).Error
		if err == nil {
			return tx.Model(&MonitorResult{}).Where("id = ?", old.Id).Updates(map[string]any{
				"latency_ms": r.LatencyMs, "status": r.Status, "error_msg": r.ErrorMsg, "checked_at": r.CheckedAt,
			}).Error
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		return tx.Create(r).Error
	})
}
func (m *defaultModel) GetLatestResultByNodeIdsAndISP(ctx context.Context, nodeIds []int64, isp string) (map[int64]*MonitorResult, error) {
	res := map[int64]*MonitorResult{}
	if len(nodeIds) == 0 {
		return res, nil
	}
	if isp == "" {
		return res, nil
	}
	var list []*MonitorResult
	err := m.DB.WithContext(ctx).
		Raw(`SELECT r.* FROM node_latency_monitor_result r
		JOIN node_latency_monitor_task mt ON mt.id = r.task_id AND mt.enabled = 1
		JOIN (
			SELECT r2.node_id, MAX(r2.id) max_id
			FROM node_latency_monitor_result r2
			JOIN node_latency_monitor_task mt2 ON mt2.id = r2.task_id AND mt2.enabled = 1
			WHERE r2.node_id IN ? AND r2.isp = ?
			GROUP BY r2.node_id
		) t ON r.node_id=t.node_id AND r.id=t.max_id
		WHERE r.isp = ?`, nodeIds, isp, isp).
		Scan(&list).Error
	if err != nil {
		return nil, err
	}
	for _, it := range list {
		res[it.NodeId] = it
	}
	return res, nil
}
