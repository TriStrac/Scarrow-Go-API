package controllers

import (
	"net/http"
	"time"

	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/gin-gonic/gin"
)

type ReportController struct {
	reportService service.ReportService
}

func NewReportController(reportService service.ReportService) *ReportController {
	return &ReportController{reportService: reportService}
}

func (c *ReportController) GetSummary(ctx *gin.Context) {
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// You might extract timeframe (e.g., 'last_7_days', 'last_30_days') from query params here
	timeframe := ctx.DefaultQuery("timeframe", "last_7_days")

	summary, err := c.reportService.GetSummary(callerID.(string), timeframe)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

func (c *ReportController) GetHubReport(ctx *gin.Context) {
	hubID := ctx.Param("hubId")

	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = &t
		}
	}
	if endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			t = t.Add(24*time.Hour - time.Nanosecond)
			endDate = &t
		}
	}

	report, err := c.reportService.GetHubReport(hubID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, report)
}
