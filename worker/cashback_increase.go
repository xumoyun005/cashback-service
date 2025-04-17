package worker

import (
	"cashback-app/config"
	"cashback-app/enum"
	"cashback-app/models"
	"cashback-app/requests"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var CashbackQueue = make(chan requests.CashbackRequest, 100)

func EnqueueCashback(req requests.CashbackRequest) {
	CashbackQueue <- req
}

func StartCashbackWorker() {
	for req := range CashbackQueue {
		go processCashback(req)
	}
}

func processCashback(req requests.CashbackRequest) {
	db := config.DB
	cashbackAmount := req.TariffPrice * 0.01

	if cashbackAmount <= 0 {
		logrus.WithField("trace_code", req.TraceCode).Warn("Cashback amount is zero, skipping")
		sendResult(req.TraceCode, fmt.Errorf("cashback amount must be greater than 0"))
		return
	}

	tx := db.Begin()

	var cashback models.Cashback
	err := tx.Where("cinerama_user_id = ?", req.CineramaUserID).First(&cashback).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cashback = models.Cashback{
				CineramaUserId: req.CineramaUserID,
				CashbackAmount: cashbackAmount,
				TuronUserId:    0,
			}
			if err := tx.Create(&cashback).Error; err != nil {
				tx.Rollback()
				logrus.WithField("trace_code", req.TraceCode).Error("Error creating cashback")
				sendResult(req.TraceCode, fmt.Errorf("failed to create cashback: %w", err))
				return
			}
		} else {
			tx.Rollback()
			logrus.WithField("trace_code", req.TraceCode).Error("Error finding cashback")
			sendResult(req.TraceCode, fmt.Errorf("failed to find cashback: %w", err))
			return
		}
	} else {
		cashback.CashbackAmount += cashbackAmount
		if err := tx.Save(&cashback).Error; err != nil {
			tx.Rollback()
			logrus.WithField("trace_code", req.TraceCode).Error("Error updating cashback")
			sendResult(req.TraceCode, fmt.Errorf("failed to update cashback: %w", err))
			return
		}
	}

	history := models.CashbackHistory{
		CashbackId:     cashback.ID,
		Device:         req.Device,
		HostIp:         req.HostIP,
		CashbackAmount: cashbackAmount,
		Type:           enum.Increased,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		logrus.WithField("trace_code", req.TraceCode).Error("Error creating history")
		sendResult(req.TraceCode, fmt.Errorf("failed to create history: %w", err))
		return
	}

	if err := tx.Commit().Error; err != nil {
		logrus.WithField("trace_code", req.TraceCode).Error("Commit failed")
		sendResult(req.TraceCode, fmt.Errorf("commit error: %w", err))
		return
	}

	sendResult(req.TraceCode, nil, cashbackAmount)
}
