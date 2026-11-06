package entity

import (
	"time"
)

type DeviceLog struct {
	BaseModel
	SessionID      string    `json:"session_id"`
	RememberToken  string    `json:"-" gorm:"-"`
	UserID         uint      `json:"user_id"`
	User           User      `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserAgent      string    `json:"user_agent"`
	IpAddress      string    `json:"ip_address"`
	LastActivityAt time.Time `json:"last_activity_at" gorm:"default:null"`
}
