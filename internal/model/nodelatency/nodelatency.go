package nodelatency

import "time"

type MonitorTask struct {
	Id              int64  `gorm:"primaryKey;autoIncrement"`
	Name            string `gorm:"type:varchar(128);not null;default:''"`
	MonitorType     string `gorm:"type:varchar(32);not null;default:'tcp'"`
	Target          string `gorm:"type:varchar(255);not null;default:''"`
	TargetCt        string `gorm:"column:target_ct;type:varchar(255);not null;default:''"`
	TargetCu        string `gorm:"column:target_cu;type:varchar(255);not null;default:''"`
	TargetCm        string `gorm:"column:target_cm;type:varchar(255);not null;default:''"`
	NodeIds         string `gorm:"type:text;not null"`
	IntervalSeconds int    `gorm:"not null;default:60"`
	Enabled         bool   `gorm:"not null;default:true"`
	LastRunAt       *time.Time
	CreatedAt       time.Time `gorm:"<-:create"`
	UpdatedAt       time.Time
}

func (MonitorTask) TableName() string { return "node_latency_monitor_task" }

type MonitorResult struct {
	Id        int64     `gorm:"primaryKey;autoIncrement"`
	TaskId    int64     `gorm:"not null"`
	NodeId    int64     `gorm:"not null"`
	ISP       string    `gorm:"column:isp;type:varchar(16);not null;default:''"`
	LatencyMs int64     `gorm:"not null;default:-1"`
	Status    string    `gorm:"type:varchar(16);not null;default:'offline'"`
	ErrorMsg  string    `gorm:"type:varchar(255);not null;default:''"`
	CheckedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP(3)"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}

func (MonitorResult) TableName() string { return "node_latency_monitor_result" }
