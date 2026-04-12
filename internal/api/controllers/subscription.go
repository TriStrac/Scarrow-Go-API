package controllers

import (
	"net/http"

	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/gin-gonic/gin"
)

type SubscriptionController struct {
	subscriptionService service.SubscriptionService
}

func NewSubscriptionController(subscriptionService service.SubscriptionService) *SubscriptionController {
	return &SubscriptionController{subscriptionService: subscriptionService}
}

type CheckoutReq struct {
	PlanID string `json:"plan_id" binding:"required"`
}

type VerifyPaymentReq struct {
	ReferenceID string `json:"reference_id" binding:"required"`
}

func (c *SubscriptionController) GetAvailablePlans(ctx *gin.Context) {
	plans, err := c.subscriptionService.GetAvailablePlans()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (c *SubscriptionController) GetMySubscription(ctx *gin.Context) {
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sub, err := c.subscriptionService.GetMySubscription(callerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if sub == nil {
		ctx.JSON(http.StatusOK, gin.H{"subscription": nil, "message": "No active subscription"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"subscription": sub})
}

func (c *SubscriptionController) CreateCheckoutSession(ctx *gin.Context) {
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CheckoutReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.subscriptionService.CreateCheckoutSession(callerID.(string), req.PlanID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *SubscriptionController) VerifyPayment(ctx *gin.Context) {
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req VerifyPaymentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.subscriptionService.VerifyPayment(callerID.(string), req.ReferenceID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Payment verified and subscription activated."})
}
