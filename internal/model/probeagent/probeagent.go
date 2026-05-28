package probeagent

import "time"

type Agent struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`
	ServerId   int64  `gorm:"uniqueIndex;not null"`
	Name       string `gorm:"type:varchar(128);not null;default:''"`
	TokenHash  string `gorm:"type:varchar(128);not null;uniqueIndex;default:''"`
	Status     string `gorm:"type:varchar(16);not null;default:'offline'"`
	Version    string `gorm:"type:varchar(32);not null;default:''"`
	LastSeenAt *time.Time
	CreatedAt  time.Time `gorm:"<-:create"`
	UpdatedAt  time.Time
}

func (Agent) TableName() string { return "probe_agent" }

type Target struct {
	Id              int64     `gorm:"primaryKey;autoIncrement"`
	ServerId        int64     `gorm:"uniqueIndex;not null"`
	TargetCt        string    `gorm:"column:target_ct;type:varchar(255);not null;default:''"`
	TargetCu        string    `gorm:"column:target_cu;type:varchar(255);not null;default:''"`
	TargetCm        string    `gorm:"column:target_cm;type:varchar(255);not null;default:''"`
	Enabled         bool      `gorm:"not null;default:true"`
	IntervalSeconds int       `gorm:"not null;default:30"`
	CreatedAt       time.Time `gorm:"<-:create"`
	UpdatedAt       time.Time
}

func (Target) TableName() string { return "probe_agent_target" }

type Result struct {
	Id        int64     `gorm:"primaryKey;autoIncrement"`
	ServerId  int64     `gorm:"not null;uniqueIndex:uk_server_isp,priority:1"`
	ISP       string    `gorm:"type:varchar(16);not null;uniqueIndex:uk_server_isp,priority:2;default:''"`
	LatencyMs int64     `gorm:"not null;default:-1"`
	Status    string    `gorm:"type:varchar(16);not null;default:'offline'"`
	ErrorMsg  string    `gorm:"type:varchar(255);not null;default:''"`
	ProbeMode string    `gorm:"type:varchar(32);not null;default:'tcp'"`
	CheckedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP(3)"`
	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}

func (Result) TableName() string { return "probe_agent_result" }
