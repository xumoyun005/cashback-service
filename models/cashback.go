package models

import "gorm.io/gorm"

type Cashback struct {
	gorm.Model
	CineramaUserId int     `json:"cinerama_user_id" gorm:"unique" binding:"required"`
	TuronUserId    int     `json:"turon_user_id"`
	CashbackAmount float64 `json:"cashback_amount" binding:"required"`
}
