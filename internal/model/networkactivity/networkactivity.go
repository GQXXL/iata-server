package networkactivity

import "time"

type UserNetworkActivity struct {
	Id              int64     `gorm:"primaryKey"`
	ServerId        int64     `gorm:"index:idx_server_id;not null;default:0"`
	UserId          int64     `gorm:"index:idx_user_id;not null"`
	SubscribeId     int64     `gorm:"index:idx_subscribe_id;not null;default:0"`
	UserSubscribeId int64     `gorm:"index:idx_user_subscribe_id;not null;default:0"`
	Domain          string    `gorm:"type:varchar(255);index:idx_domain;not null;default:''"`
	ClientIP        string    `gorm:"type:varchar(64);not null;default:''"`
	UserAgent       string    `gorm:"type:varchar(1024);not null;default:''"`
	Upload          int64     `gorm:"not null;default:0"`
	Download        int64     `gorm:"not null;default:0"`
	Timestamp       time.Time `gorm:"index:idx_timestamp;default:CURRENT_TIMESTAMP(3);not null"`
	CreatedAt       time.Time `gorm:"<-:create;default:CURRENT_TIMESTAMP(3);not null"`
}

func (UserNetworkActivity) TableName() string {
	return "user_network_activity"
}
