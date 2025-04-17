package handler

import (
	"cashback-app/requests"
	"cashback-app/utils"
	"cashback-app/worker"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleCashback(c *gin.Context) {
	var req requests.CashbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{
				"status_code": 400,
				"message":     res.Error.Error(),
				"trace_code":  traceCode,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status_code": 200,
			"message":     "Cashback added successfully",
			"trace_code":  traceCode,
			"data": gin.H{
				"added_cashback": res.Data,
			},
		})
	case <-time.After(5 * time.Second):
		c.JSON(http.StatusRequestTimeout, gin.H{
			"status_code": 408,
			"message":     "Timeout",
			"trace_code":  traceCode,
		})
	}
}

func HandleCashbackDecrease(c *gin.Context) {
	var req requests.CashbackDecreaseQueue
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error(), "trace_code": traceCode})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Cashback decreased successfully", "trace_code": traceCode})
	case <-time.After(5 * time.Second):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Timeout", "trace_code": traceCode})
	}
}
