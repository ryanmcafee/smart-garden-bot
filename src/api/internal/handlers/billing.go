package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/services"
)

// BillingHandler handles billing-related requests
type BillingHandler struct {
	billingService *services.BillingService
	validator      *validator.Validate
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(billingService *services.BillingService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
		validator:      validator.New(),
	}
}

// ===== SUBSCRIPTION PLAN ENDPOINTS =====

// GetSubscriptionPlans retrieves all available subscription plans
func (h *BillingHandler) GetSubscriptionPlans(c *gin.Context) {
	plans, err := h.billingService.GetSubscriptionPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FETCH_FAILED",
				Message: "Failed to retrieve subscription plans",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      plans,
		Timestamp: time.Now(),
	})
}

// GetSubscriptionPlan retrieves a specific subscription plan
func (h *BillingHandler) GetSubscriptionPlan(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_PLAN_ID",
				Message: "Invalid plan ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	plan, err := h.billingService.GetSubscriptionPlanByID(c.Request.Context(), planID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "PLAN_NOT_FOUND",
				Message: "Subscription plan not found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      plan,
		Timestamp: time.Now(),
	})
}

// ===== SUBSCRIPTION ENDPOINTS =====

// GetCurrentSubscription retrieves the user's current subscription
func (h *BillingHandler) GetCurrentSubscription(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	subscription, err := h.billingService.GetUserSubscription(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SUBSCRIPTION_NOT_FOUND",
				Message: "No active subscription found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      subscription,
		Timestamp: time.Now(),
	})
}

// CreateSubscription creates a new subscription for the user
func (h *BillingHandler) CreateSubscription(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
				Details: getValidationErrors(err),
			},
			Timestamp: time.Now(),
		})
		return
	}

	subscription, err := h.billingService.CreateSubscription(c.Request.Context(), userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "CREATION_FAILED"
		
		if err.Error() == "user already has an active subscription" {
			statusCode = http.StatusConflict
			errorCode = "SUBSCRIPTION_EXISTS"
		} else if err.Error() == "invalid subscription plan: subscription plan not found" {
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_PLAN"
		}

		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    errorCode,
				Message: "Failed to create subscription",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success:   true,
		Data:      subscription,
		Timestamp: time.Now(),
	})
}

// UpdateSubscription updates the user's subscription
func (h *BillingHandler) UpdateSubscription(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
				Details: getValidationErrors(err),
			},
			Timestamp: time.Now(),
		})
		return
	}

	subscription, err := h.billingService.UpdateSubscription(c.Request.Context(), userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "UPDATE_FAILED"
		
		if err.Error() == "subscription not found: no active subscription found" {
			statusCode = http.StatusNotFound
			errorCode = "SUBSCRIPTION_NOT_FOUND"
		}

		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    errorCode,
				Message: "Failed to update subscription",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      subscription,
		Timestamp: time.Now(),
	})
}

// CancelSubscription cancels the user's subscription
func (h *BillingHandler) CancelSubscription(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Parse query parameter for cancel behavior
	cancelAtPeriodEndStr := c.DefaultQuery("cancel_at_period_end", "true")
	cancelAtPeriodEnd, err := strconv.ParseBool(cancelAtPeriodEndStr)
	if err != nil {
		cancelAtPeriodEnd = true // Default to canceling at period end
	}

	err = h.billingService.CancelSubscription(c.Request.Context(), userID, cancelAtPeriodEnd)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "CANCELLATION_FAILED"
		
		if err.Error() == "subscription not found: no active subscription found" {
			statusCode = http.StatusNotFound
			errorCode = "SUBSCRIPTION_NOT_FOUND"
		}

		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    errorCode,
				Message: "Failed to cancel subscription",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	message := "Subscription cancelled successfully"
	if cancelAtPeriodEnd {
		message = "Subscription will be cancelled at the end of current period"
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"message":                message,
			"cancel_at_period_end":   cancelAtPeriodEnd,
			"cancelled_at":          time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now(),
	})
}

// ===== PAYMENT METHOD ENDPOINTS =====

// GetPaymentMethods retrieves payment methods for the user
func (h *BillingHandler) GetPaymentMethods(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	paymentMethods, err := h.billingService.ListPaymentMethods(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FETCH_FAILED",
				Message: "Failed to retrieve payment methods",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      paymentMethods,
		Timestamp: time.Now(),
	})
}

// ===== INVOICE ENDPOINTS =====

// GetInvoices retrieves invoices for the user
func (h *BillingHandler) GetInvoices(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	startingAfter := c.Query("starting_after")

	invoices, err := h.billingService.ListInvoices(c.Request.Context(), userID, limit, startingAfter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FETCH_FAILED",
				Message: "Failed to retrieve invoices",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      invoices,
		Timestamp: time.Now(),
	})
}

// ===== USAGE ENDPOINTS =====

// GetUsage retrieves current usage statistics for the user's subscription
func (h *BillingHandler) GetUsage(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	usage, err := h.billingService.GetSubscriptionUsage(c.Request.Context(), userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "FETCH_FAILED"
		
		if err.Error() == "subscription not found: no active subscription found" {
			statusCode = http.StatusNotFound
			errorCode = "SUBSCRIPTION_NOT_FOUND"
		}

		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    errorCode,
				Message: "Failed to retrieve usage statistics",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      usage,
		Timestamp: time.Now(),
	})
}

// ===== WEBHOOK ENDPOINTS =====

// ProcessStripeWebhook processes Stripe webhook events
func (h *BillingHandler) ProcessStripeWebhook(c *gin.Context) {
	// In a real implementation, you would:
	// 1. Verify the webhook signature using Stripe's webhook secret
	// 2. Parse the webhook payload
	// 3. Process the event based on its type
	
	var webhookEvent models.WebhookEvent
	if err := c.ShouldBindJSON(&webhookEvent); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_WEBHOOK",
				Message: "Invalid webhook payload",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	err := h.billingService.ProcessWebhook(c.Request.Context(), &webhookEvent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "WEBHOOK_PROCESSING_FAILED",
				Message: "Failed to process webhook",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"message":   "Webhook processed successfully",
			"event_id":  webhookEvent.ID,
			"event_type": webhookEvent.Type,
		},
		Timestamp: time.Now(),
	})
}

// ===== ADMIN ENDPOINTS =====

// GetSubscriptionStats returns subscription statistics (admin only)
func (h *BillingHandler) GetSubscriptionStats(c *gin.Context) {
	// Check if user has admin role
	userRole := c.GetString("user_role")
	if userRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Insufficient permissions",
			},
			Timestamp: time.Now(),
		})
		return
	}

	stats, err := h.billingService.GetSubscriptionStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "STATS_FETCH_FAILED",
				Message: "Failed to retrieve subscription statistics",
				Details: map[string]string{"error": err.Error()},
			},
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      stats,
		Timestamp: time.Now(),
	})
}

// ===== PORTAL ENDPOINTS =====

// CreateBillingPortalSession creates a Stripe Customer Portal session
func (h *BillingHandler) CreateBillingPortalSession(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Get user's subscription to get Stripe customer ID
	subscription, err := h.billingService.GetUserSubscription(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SUBSCRIPTION_NOT_FOUND",
				Message: "No active subscription found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	if subscription.StripeCustomerID == nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NO_CUSTOMER_ID",
				Message: "No Stripe customer ID found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// In a real implementation, you would create a billing portal session:
	// portalSession, err := stripe.BillingPortalSession.New(&stripe.BillingPortalSessionParams{
	//     Customer:  stripe.String(*subscription.StripeCustomerID),
	//     ReturnURL: stripe.String(returnURL),
	// })

	// For now, return a mock response
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"url": "https://billing.stripe.com/session/mock_session_id",
			"message": "Billing portal session created (mock)",
		},
		Timestamp: time.Now(),
	})
}

// ===== PREVIEW ENDPOINTS =====

// PreviewSubscriptionChange previews the cost of changing subscription plan
func (h *BillingHandler) PreviewSubscriptionChange(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "User ID not found in token",
			},
			Timestamp: time.Now(),
		})
		return
	}

	newPlanIDStr := c.Query("plan_id")
	if newPlanIDStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_PLAN_ID",
				Message: "plan_id query parameter is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	newPlanID, err := uuid.Parse(newPlanIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_PLAN_ID",
				Message: "Invalid plan ID format",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Get current subscription
	currentSubscription, err := h.billingService.GetUserSubscription(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SUBSCRIPTION_NOT_FOUND",
				Message: "No active subscription found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Get new plan details
	newPlan, err := h.billingService.GetSubscriptionPlanByID(c.Request.Context(), newPlanID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "PLAN_NOT_FOUND",
				Message: "Subscription plan not found",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Calculate preview (simplified logic)
	var currentPrice, newPrice float64
	if currentSubscription.BillingCycle == models.BillingCycleMonthly {
		if currentSubscription.Plan != nil && currentSubscription.Plan.PriceMonthly != nil {
			currentPrice = *currentSubscription.Plan.PriceMonthly
		}
		if newPlan.PriceMonthly != nil {
			newPrice = *newPlan.PriceMonthly
		}
	} else {
		if currentSubscription.Plan != nil && currentSubscription.Plan.PriceYearly != nil {
			currentPrice = *currentSubscription.Plan.PriceYearly
		}
		if newPlan.PriceYearly != nil {
			newPrice = *newPlan.PriceYearly
		}
	}

	priceDifference := newPrice - currentPrice
	
	// In a real implementation, you would use Stripe's subscription preview API
	preview := gin.H{
		"current_plan":     currentSubscription.Plan,
		"new_plan":         newPlan,
		"current_price":    currentPrice,
		"new_price":        newPrice,
		"price_difference": priceDifference,
		"currency":         newPlan.Currency,
		"billing_cycle":    currentSubscription.BillingCycle,
		"next_billing_date": currentSubscription.CurrentPeriodEnd.Format(time.RFC3339),
		"proration_credit": 0, // Would be calculated based on remaining time
		"immediate_charge": priceDifference > 0,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      preview,
		Timestamp: time.Now(),
	})
}