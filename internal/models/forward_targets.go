package model

import (
	"time"
)

type ForwardTargets struct {
	Id        int       `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(64);NOT NULL" json:"name"`
	Email     string    `gorm:"column:email;type:varchar(128);NOT NULL" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;NOT NULL" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP;NOT NULL" json:"updated_at"`
}

func (ForwardTargets) TableName() string {
	return "forward_targets"
}
