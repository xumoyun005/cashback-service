package worker

import (
	"cashback-app/config"
	"cashback-app/enum"
	"cashback-app/models"
	"cashback-app/requests"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

var CashbackDecreaseQueue = make(chan requests.CashbackDecreaseQueue, 100)

func EnqueueCashbackDecrease(req requests.CashbackDecreaseQueue) {
	CashbackDecreaseQueue <- req
}

func StartDescreaseCashbackWorker() {
	for req := range CashbackDecreaseQueue {
		go processCashbackDecrease(req)
	}
}

func sendResult(traceCode string, err error) {
	if ch, ok := ResultMap.Load(traceCode); ok {
		ch.(chan Result) <- Result{TraceCode: traceCode, Error: err}
	}
}

func processCashbackDecrease(req requests.CashbackDecreaseQueue) {
	db := config.DB
	tx := db.Begin()

	var cashback models.Cashback
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("cinerama_user_id = ?", req.CineramaUserID).
		First(&cashback).Error

	if err != nil {
		tx.Rollback()
		logrus.WithField("trace_code", req.TraceCode).Error("Record not found")
		sendResult(req.TraceCode, fmt.Errorf("cashback record not found: %w", err))
		return
	}

	if cashback.CashbackAmount < req.DecreaseCashbackAmount {
		tx.Rollback()
		logrus.WithField("trace_code", req.TraceCode).Error("Insufficient cashback")
		sendResult(req.TraceCode, fmt.Errorf("not enough cashback"))
		return
	}

	cashback.CashbackAmount -= req.DecreaseCashbackAmount
	if err := tx.Save(&cashback).Error; err != nil {
		tx.Rollback()
		logrus.WithField("trace_code", req.TraceCode).Error("Update failed")
		sendResult(req.TraceCode, fmt.Errorf("failed to update cashback: %w", err))
		return
	}

	history := models.CashbackHistory{
		CashbackId:     cashback.ID,
		CashbackAmount: req.DecreaseCashbackAmount,
		HostIp:         req.HostIP,
		Device:         req.Device,
		Type:           enum.Decreased,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		logrus.WithField("trace_code", req.TraceCode).Error("History create failed")
		sendResult(req.TraceCode, fmt.Errorf("failed to create history: %w", err))
		return
	}

	if err := tx.Commit().Error; err != nil {
		logrus.WithField("trace_code", req.TraceCode).Error("Commit failed")
		sendResult(req.TraceCode, fmt.Errorf("commit error: %w", err))
		return
	}

	sendResult(req.TraceCode, nil)
}
