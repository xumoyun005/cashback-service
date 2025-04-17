package models

import (
	"cashback-app/enum"

	"gorm.io/gorm"
)


type CashbackHistory struct {
	gorm.Model
	CashbackId uint `json:"cashback_id"`
	CashbackAmount float64 `json:"cashback_amount"`
	HostIp string `json:"host_ip"`
	Device string `json:"device"`
	Type  enum.Type `json:"type"` 
}