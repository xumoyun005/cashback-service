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
		sendResult(req.TraceCode, fmt.Errorf("cashback not found: %w", err))
		return
	}

	if cashback.CashbackAmount < req.DecreaseCashbackAmount {
		tx.Rollback()
		logrus.WithField("trace_code", req.TraceCode).Error("Insufficient cashback")
		sendResult(req.TraceCode, fmt.Errorf("not enough cashback"))
		return
	}

	cashback.CashbackAmount -= req.DecreaseCashbackAmount
	if cashback.CashbackAmount < 0 {
		cashback.CashbackAmount = 0
	}

	if err := tx.Save(&cashback).Error; err != nil {
		tx.Rollback()
		logrus.WithField("trace_code", req.TraceCode).Error("Failed processing cashback")
		sendResult(req.TraceCode, fmt.Errorf("failed to process cashback: %w", err))
		return
	}

	history := models.CashbackHistory{
		CashbackId:     cashback.ID,
		CashbackAmount: req.DecreaseCashbackAmount,
		HostIp:         req.HostIP,
		Device:         req.Device,
		Type:           enum.Decreased,
	}

	if req.DecreaseCashbackAmount <= 0 {
		tx.Rollback()
		logrus.WithField("trace_code", req.TraceCode).Error("Attempted to decrease by 0")
		sendResult(req.TraceCode, fmt.Errorf("cannot decrease cashback by 0"))
		return
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		logrus.WithField("trace_code", req.TraceCode).Error("Failed to create cashback history")
		sendResult(req.TraceCode, fmt.Errorf("failed to create cashback history: %w", err))
		return
	}

	if err := tx.Commit().Error; err != nil {
		logrus.WithField("trace_code", req.TraceCode).Error("Commit failed")
		sendResult(req.TraceCode, fmt.Errorf("failed to commit transaction: %w", err))
		return
	}

	sendResult(req.TraceCode, nil, req.DecreaseCashbackAmount)
}

func sendResult(traceCode string, err error, data ...interface{}) {
	if ch, ok := ResultMap.Load(traceCode); ok {
		var payload interface{}
		if len(data) > 0 {
			payload = data[0]
		}
		ch.(chan Result) <- Result{TraceCode: traceCode, Error: err, Data: payload}
	}
}
