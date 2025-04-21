package handler

import (
	"cashback-app/config"
	"cashback-app/models"
	"cashback-app/requests"
	"cashback-app/response"
	"cashback-app/utils"
	"cashback-app/worker"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleCashback(c *gin.Context) {
	var req requests.CashbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.CashbackResponse{
			Code: -5000,
		})
		return
	}

	traceCode := utils.GenerateTraceCode()
	req.TraceCode = traceCode

	resultChan := make(chan worker.Result)
	worker.ResultMap.Store(traceCode, resultChan)
	defer worker.ResultMap.Delete(traceCode)

	worker.EnqueueCashback(req)

	select {
	case res := <-resultChan:
		if res.Error != nil {

			c.JSON(http.StatusBadRequest, response.CashbackResponse{
				Code: -5001,
			})
			return
		}
		c.JSON(http.StatusOK, response.CashbackResponse{
			Data: res.Data,
		})
	case <-time.After(5 * time.Second):
		c.JSON(http.StatusRequestTimeout, response.CashbackResponse{
			Code: -5002,
		})
	}
}

func HandleCashbackDecrease(c *gin.Context) {
	var req requests.CashbackDecreaseQueue
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.CashbackResponse{
			Code: -5003,
		})
		return
	}

	traceCode := utils.GenerateTraceCode()
	req.TraceCode = traceCode
	resultChan := make(chan worker.Result)
	worker.ResultMap.Store(traceCode, resultChan)
	defer worker.ResultMap.Delete(traceCode)

	worker.EnqueueCashbackDecrease(req)

	select {
	case res := <-resultChan:
		if res.Error != nil {
			c.JSON(http.StatusBadRequest, response.CashbackResponse{
				Code: -5004,
			})
			return
		}
		c.JSON(http.StatusOK, response.CashbackResponse{
			Data: res.Data,
		})
	case <-time.After(5 * time.Second):
		c.JSON(http.StatusRequestTimeout, response.CashbackResponse{
			Code: -5005,
		})
	}
}
func GetCashbackByCineramaId(c *gin.Context) {

	cineramaUserId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.CashbackResponse{
			Code: -5006,
		})
		return
	}

	var cashback models.Cashback
	err = config.DB.Where("cinerama_user_id = ?", cineramaUserId).First(&cashback).Error
	if err != nil {
		c.JSON(http.StatusNotFound, response.CashbackResponse{
			Code: -5007,
			Data: gin.H{
				"created_at":       nil,
				"updated_at":       nil,
				"deleted_at":       nil,
				"cinerama_user_id": cineramaUserId,
				"cashback_amount":  0,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.CashbackResponse{
		Code: 200,
		Data: gin.H{
			"created_at":       cashback.CreatedAt,
			"updated_at":       cashback.UpdatedAt,
			"deleted_at":       cashback.DeletedAt,
			"cinerama_user_id": cashback.CineramaUserId,
			"cashback_amount":  cashback.CashbackAmount,
		},
	})
}

func GetCashbackHistoryByCineramaId(c *gin.Context) {
	cineramaUserId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.CashbackResponse{
			Code: -5008,
		})
		return
	}
	var cashbackHistory []models.CashbackHistory
	err = config.DB.Joins("JOIN cashbacks on cashback_histories.cashback_id = cashbacks.id").
		Where("cashbacks.cinerama_user_id = ?", cineramaUserId).
		Find(&cashbackHistory).Error
	if err != nil {
		c.JSON(http.StatusNotFound, response.CashbackResponse{
			Code: -5009,
		})
		return
	}
	if len(cashbackHistory) == 0 {
		c.JSON(http.StatusNotFound, response.CashbackResponse{
			Code: -5009,
			Data: gin.H{
				"created_at":      nil,
				"updated_at":      nil,
				"deleted_at":      nil,
				"cashback_id":     0,
				"cashback_amount": 0,
				"host_ip":         nil,
				"device":          nil,
				"type":            nil,
			},
		})
		return
	}

	c.JSON(http.StatusOK, response.CashbackResponse{
		Code: 200,
		Data: gin.H{
			"cashback_histories": filterHistory,
			"cinerama_user_id":   cineramaUserId,
		},
	})
}

func filterHistory(cashbackHistory []models.CashbackHistory) []gin.H {
	filterHistory := make([]gin.H, 0, len(cashbackHistory))
	for _, history := range cashbackHistory {
		filterHistory = append(filterHistory, gin.H{
			"cashback_amount": history.CashbackAmount,
			"host_ip":         history.HostIp,
			"device":          history.Device,
			"type":            history.Type,
			"updated_at":      history.UpdatedAt,
			"created_at":      history.CreatedAt,
			"deleted_at":      history.DeletedAt,
		})
	}
	return filterHistory
}
