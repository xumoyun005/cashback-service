package handler

import (
	"cashback-app/requests"
	"cashback-app/response"
	"cashback-app/utils"
	"cashback-app/worker"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleCashback(c *gin.Context) {
	var raw map[string]interface{}
	if err := c.BindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, response.CashbackResponse{
			StatusCode: 400,
			Message:    "Bad Request: " + err.Error(),
		})
		return
	}

	var req requests.CashbackRequest
	jsonBytes, _ := json.Marshal(raw)
	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		c.JSON(http.StatusInternalServerError, response.CashbackResponse{
			StatusCode: 500,
			Message:    "Failed to unmarshal JSON",
		})
		return
	}

	worker.EnqueueCashback(req)

	c.JSON(http.StatusOK, response.CashbackResponse{
		StatusCode: 200,
		Message:    "Success",
	})
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
