package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/database"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
)

// BillingService handles billing operations
type BillingService struct {
	db           *database.DB
	stripeAPIKey string
}

// NewBillingService creates a new billing service
func NewBillingService(db *database.DB, stripeAPIKey string) *BillingService {
	return &BillingService{
		db:           db,
		stripeAPIKey: stripeAPIKey,
	}
}

// ===== SUBSCRIPTION PLAN METHODS =====

// GetSubscriptionPlans retrieves all active subscription plans
func (s *BillingService) GetSubscriptionPlans(ctx context.Context) ([]models.SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, price_monthly, price_yearly, currency,
		       max_gardens, max_zones_per_garden, max_devices_per_garden,
		       max_api_calls_per_month, features, is_active, created_at, updated_at
		FROM billing.subscription_plans
		WHERE is_active = true
		ORDER BY price_monthly ASC NULLS LAST, price_yearly ASC NULLS LAST
	`
	
	var plans []models.SubscriptionPlan
	err := s.db.SelectContext(ctx, &plans, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plans: %w", err)
	}
	
	return plans, nil
}

// GetSubscriptionPlanByID retrieves a subscription plan by ID
func (s *BillingService) GetSubscriptionPlanByID(ctx context.Context, planID uuid.UUID) (*models.SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, price_monthly, price_yearly, currency,
		       max_gardens, max_zones_per_garden, max_devices_per_garden,
		       max_api_calls_per_month, features, is_active, created_at, updated_at
		FROM billing.subscription_plans
		WHERE id = $1 AND is_active = true
	`
	
	var plan models.SubscriptionPlan
	err := s.db.GetContext(ctx, &plan, query, planID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription plan not found")
		}
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}
	
	return &plan, nil
}

// ===== SUBSCRIPTION METHODS =====

// GetUserSubscription retrieves the active subscription for a user
func (s *BillingService) GetUserSubscription(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT s.id, s.user_id, s.plan_id, s.stripe_subscription_id, s.stripe_customer_id,
		       s.status, s.billing_cycle, s.started_at, s.current_period_start,
		       s.current_period_end, s.cancelled_at, s.ended_at, s.trial_start,
		       s.trial_end, s.usage_data, s.created_at, s.updated_at,
		       p.id as plan_id, p.name as plan_name, p.description as plan_description,
		       p.price_monthly as plan_price_monthly, p.price_yearly as plan_price_yearly,
		       p.currency as plan_currency, p.max_gardens as plan_max_gardens,
		       p.max_zones_per_garden as plan_max_zones_per_garden,
		       p.max_devices_per_garden as plan_max_devices_per_garden,
		       p.max_api_calls_per_month as plan_max_api_calls_per_month,
		       p.features as plan_features, p.is_active as plan_is_active,
		       p.created_at as plan_created_at, p.updated_at as plan_updated_at
		FROM billing.subscriptions s
		LEFT JOIN billing.subscription_plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status IN ('active', 'trialing', 'past_due')
		ORDER BY s.created_at DESC
		LIMIT 1
	`
	
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user subscription: %w", err)
	}
	defer rows.Close()
	
	if !rows.Next() {
		return nil, fmt.Errorf("no active subscription found")
	}
	
	var subscription models.Subscription
	var plan models.SubscriptionPlan
	var planID uuid.UUID
	
	err = rows.Scan(
		&subscription.ID, &subscription.UserID, &subscription.PlanID,
		&subscription.StripeSubscriptionID, &subscription.StripeCustomerID,
		&subscription.Status, &subscription.BillingCycle, &subscription.StartedAt,
		&subscription.CurrentPeriodStart, &subscription.CurrentPeriodEnd,
		&subscription.CancelledAt, &subscription.EndedAt, &subscription.TrialStart,
		&subscription.TrialEnd, &subscription.UsageData, &subscription.CreatedAt,
		&subscription.UpdatedAt,
		&planID, &plan.Name, &plan.Description, &plan.PriceMonthly,
		&plan.PriceYearly, &plan.Currency, &plan.MaxGardens,
		&plan.MaxZonesPerGarden, &plan.MaxDevicesPerGarden,
		&plan.MaxAPICallsPerMonth, &plan.Features, &plan.IsActive,
		&plan.CreatedAt, &plan.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan subscription: %w", err)
	}
	
	plan.ID = planID
	subscription.Plan = &plan
	
	return &subscription, nil
}

// CreateSubscription creates a new subscription
func (s *BillingService) CreateSubscription(ctx context.Context, userID uuid.UUID, req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	// Verify the subscription plan exists
	plan, err := s.GetSubscriptionPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription plan: %w", err)
	}
	
	// Check if user already has an active subscription
	existingSubscription, err := s.GetUserSubscription(ctx, userID)
	if err == nil && existingSubscription != nil {
		return nil, fmt.Errorf("user already has an active subscription")
	}
	
	// Create Stripe customer if needed
	stripeCustomerID, err := s.createOrGetStripeCustomer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
	}
	
	subscription := models.Subscription{
		ID:                 uuid.New(),
		UserID:             userID,
		PlanID:             req.PlanID,
		StripeCustomerID:   &stripeCustomerID,
		Status:             models.SubscriptionStatusActive,
		BillingCycle:       req.BillingCycle,
		StartedAt:          time.Now(),
		CurrentPeriodStart: time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	
	// Calculate period end based on billing cycle
	if req.BillingCycle == models.BillingCycleMonthly {
		subscription.CurrentPeriodEnd = subscription.CurrentPeriodStart.AddDate(0, 1, 0)
	} else {
		subscription.CurrentPeriodEnd = subscription.CurrentPeriodStart.AddDate(1, 0, 0)
	}
	
	// Handle trial period
	if req.TrialDays != nil && *req.TrialDays > 0 {
		trialStart := time.Now()
		trialEnd := trialStart.AddDate(0, 0, *req.TrialDays)
		subscription.TrialStart = &trialStart
		subscription.TrialEnd = &trialEnd
	}
	
	// Create Stripe subscription
	stripeSubscriptionID, err := s.createStripeSubscription(ctx, stripeCustomerID, plan, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe subscription: %w", err)
	}
	subscription.StripeSubscriptionID = &stripeSubscriptionID
	
	// Insert subscription into database
	query := `
		INSERT INTO billing.subscriptions (
			id, user_id, plan_id, stripe_subscription_id, stripe_customer_id,
			status, billing_cycle, started_at, current_period_start,
			current_period_end, trial_start, trial_end, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`
	
	_, err = s.db.ExecContext(ctx, query,
		subscription.ID, subscription.UserID, subscription.PlanID,
		subscription.StripeSubscriptionID, subscription.StripeCustomerID,
		subscription.Status, subscription.BillingCycle, subscription.StartedAt,
		subscription.CurrentPeriodStart, subscription.CurrentPeriodEnd,
		subscription.TrialStart, subscription.TrialEnd,
		subscription.CreatedAt, subscription.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	
	subscription.Plan = plan
	return &subscription, nil
}

// UpdateSubscription updates an existing subscription
func (s *BillingService) UpdateSubscription(ctx context.Context, userID uuid.UUID, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	// Get current subscription
	subscription, err := s.GetUserSubscription(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}
	
	// Build update query dynamically
	setParts := []string{}
	args := []interface{}{}
	argCount := 1
	
	if req.PlanID != nil {
		// Verify new plan exists
		_, err = s.GetSubscriptionPlanByID(ctx, *req.PlanID)
		if err != nil {
			return nil, fmt.Errorf("invalid subscription plan: %w", err)
		}
		
		setParts = append(setParts, fmt.Sprintf("plan_id = $%d", argCount))
		args = append(args, *req.PlanID)
		argCount++
	}
	
	if req.BillingCycle != nil {
		setParts = append(setParts, fmt.Sprintf("billing_cycle = $%d", argCount))
		args = append(args, *req.BillingCycle)
		argCount++
	}
	
	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *req.Status)
		argCount++
		
		// Handle cancellation
		if *req.Status == models.SubscriptionStatusCancelled {
			now := time.Now()
			setParts = append(setParts, fmt.Sprintf("cancelled_at = $%d", argCount))
			args = append(args, now)
			argCount++
		}
	}
	
	if len(setParts) == 0 {
		return subscription, nil // No updates needed
	}
	
	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())
	argCount++
	
	// Add WHERE clause
	args = append(args, subscription.ID)
	whereClause := fmt.Sprintf("id = $%d", argCount)
	
	query := fmt.Sprintf("UPDATE billing.subscriptions SET %s WHERE %s",
		fmt.Sprintf("%s", setParts), whereClause)
	
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}
	
	// Update Stripe subscription if needed
	if subscription.StripeSubscriptionID != nil {
		err = s.updateStripeSubscription(ctx, *subscription.StripeSubscriptionID, req)
		if err != nil {
			return nil, fmt.Errorf("failed to update Stripe subscription: %w", err)
		}
	}
	
	// Return updated subscription
	return s.GetUserSubscription(ctx, userID)
}

// CancelSubscription cancels a user's subscription
func (s *BillingService) CancelSubscription(ctx context.Context, userID uuid.UUID, cancelAtPeriodEnd bool) error {
	subscription, err := s.GetUserSubscription(ctx, userID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}
	
	// Cancel Stripe subscription
	if subscription.StripeSubscriptionID != nil {
		err = s.cancelStripeSubscription(ctx, *subscription.StripeSubscriptionID, cancelAtPeriodEnd)
		if err != nil {
			return fmt.Errorf("failed to cancel Stripe subscription: %w", err)
		}
	}
	
	// Update local subscription
	now := time.Now()
	var endedAt *time.Time
	status := models.SubscriptionStatusCancelled
	
	if !cancelAtPeriodEnd {
		endedAt = &now
	}
	
	query := `
		UPDATE billing.subscriptions
		SET status = $1, cancelled_at = $2, ended_at = $3, updated_at = $4
		WHERE id = $5
	`
	
	_, err = s.db.ExecContext(ctx, query, status, now, endedAt, now, subscription.ID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}
	
	return nil
}

// ===== PAYMENT METHODS =====

// ListPaymentMethods retrieves payment methods for a user
func (s *BillingService) ListPaymentMethods(ctx context.Context, userID uuid.UUID) ([]models.PaymentMethod, error) {
	subscription, err := s.GetUserSubscription(ctx, userID)
	if err != nil {
		return []models.PaymentMethod{}, nil // No subscription, no payment methods
	}
	
	if subscription.StripeCustomerID == nil {
		return []models.PaymentMethod{}, nil
	}
	
	// Get payment methods from Stripe
	return s.getStripePaymentMethods(ctx, *subscription.StripeCustomerID)
}

// ===== INVOICES =====

// ListInvoices retrieves invoices for a user
func (s *BillingService) ListInvoices(ctx context.Context, userID uuid.UUID, limit int, startingAfter string) ([]models.Invoice, error) {
	subscription, err := s.GetUserSubscription(ctx, userID)
	if err != nil {
		return []models.Invoice{}, nil // No subscription, no invoices
	}
	
	if subscription.StripeCustomerID == nil {
		return []models.Invoice{}, nil
	}
	
	// Get invoices from Stripe
	return s.getStripeInvoices(ctx, *subscription.StripeCustomerID, limit, startingAfter)
}

// ===== USAGE TRACKING =====

// GetSubscriptionUsage retrieves current usage for a subscription
func (s *BillingService) GetSubscriptionUsage(ctx context.Context, userID uuid.UUID) (*models.SubscriptionUsage, error) {
	subscription, err := s.GetUserSubscription(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}
	
	usage := &models.SubscriptionUsage{
		SubscriptionID: subscription.ID,
		PeriodStart:    subscription.CurrentPeriodStart,
		PeriodEnd:      subscription.CurrentPeriodEnd,
		LastUpdated:    time.Now(),
	}
	
	// Get gardens count
	gardensQuery := "SELECT COUNT(*) FROM gardens.gardens WHERE user_id = $1 AND deleted_at IS NULL"
	err = s.db.GetContext(ctx, &usage.GardensUsed, gardensQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gardens count: %w", err)
	}
	
	// Get devices count
	devicesQuery := `
		SELECT COUNT(d.*)
		FROM gardens.devices d
		JOIN gardens.gardens g ON d.garden_id = g.id
		WHERE g.user_id = $1 AND g.deleted_at IS NULL
	`
	err = s.db.GetContext(ctx, &usage.DevicesUsed, devicesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices count: %w", err)
	}
	
	// Get API calls count for current period
	apiCallsQuery := `
		SELECT COUNT(*)
		FROM analytics.api_usage
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
	`
	err = s.db.GetContext(ctx, &usage.APICallsUsed, apiCallsQuery, userID, usage.PeriodStart, usage.PeriodEnd)
	if err != nil {
		// API usage table might not exist yet, default to 0
		usage.APICallsUsed = 0
	}
	
	// Set limits from plan
	if subscription.Plan != nil {
		usage.GardensLimit = subscription.Plan.MaxGardens
		usage.DevicesLimit = subscription.Plan.MaxDevicesPerGarden
		usage.APICallsLimit = subscription.Plan.MaxAPICallsPerMonth
	}
	
	return usage, nil
}

// ===== STRIPE INTEGRATION HELPERS =====

// createOrGetStripeCustomer creates or retrieves a Stripe customer
func (s *BillingService) createOrGetStripeCustomer(ctx context.Context, userID uuid.UUID) (string, error) {
	// This is a placeholder - in a real implementation, you would:
	// 1. Check if user already has a Stripe customer ID
	// 2. If not, create one using Stripe API
	// 3. Store the customer ID in the users table
	
	// For now, return a mock customer ID
	return fmt.Sprintf("cus_%s", generateRandomString(14)), nil
}

// createStripeSubscription creates a subscription in Stripe
func (s *BillingService) createStripeSubscription(ctx context.Context, customerID string, plan *models.SubscriptionPlan, req *models.CreateSubscriptionRequest) (string, error) {
	// This is a placeholder - in a real implementation, you would:
	// 1. Create a Stripe subscription using their API
	// 2. Handle trial periods, billing cycles, etc.
	
	// For now, return a mock subscription ID
	return fmt.Sprintf("sub_%s", generateRandomString(14)), nil
}

// updateStripeSubscription updates a subscription in Stripe
func (s *BillingService) updateStripeSubscription(ctx context.Context, subscriptionID string, req *models.UpdateSubscriptionRequest) error {
	// This is a placeholder - in a real implementation, you would:
	// 1. Update the Stripe subscription using their API
	// 2. Handle plan changes, billing cycle changes, etc.
	
	return nil
}

// cancelStripeSubscription cancels a subscription in Stripe
func (s *BillingService) cancelStripeSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool) error {
	// This is a placeholder - in a real implementation, you would:
	// 1. Cancel the Stripe subscription using their API
	// 2. Handle immediate vs. period-end cancellation
	
	return nil
}

// getStripePaymentMethods retrieves payment methods from Stripe
func (s *BillingService) getStripePaymentMethods(ctx context.Context, customerID string) ([]models.PaymentMethod, error) {
	// This is a placeholder - in a real implementation, you would:
	// 1. Retrieve payment methods from Stripe API
	// 2. Convert to our internal format
	
	return []models.PaymentMethod{}, nil
}

// getStripeInvoices retrieves invoices from Stripe
func (s *BillingService) getStripeInvoices(ctx context.Context, customerID string, limit int, startingAfter string) ([]models.Invoice, error) {
	// This is a placeholder - in a real implementation, you would:
	// 1. Retrieve invoices from Stripe API
	// 2. Convert to our internal format
	
	return []models.Invoice{}, nil
}

// ===== WEBHOOK HANDLING =====

// ProcessWebhook processes a Stripe webhook event
func (s *BillingService) ProcessWebhook(ctx context.Context, event *models.WebhookEvent) error {
	switch event.Type {
	case models.WebhookEventCustomerSubscriptionCreated:
		return s.handleSubscriptionCreated(ctx, event)
	case models.WebhookEventCustomerSubscriptionUpdated:
		return s.handleSubscriptionUpdated(ctx, event)
	case models.WebhookEventCustomerSubscriptionDeleted:
		return s.handleSubscriptionDeleted(ctx, event)
	case models.WebhookEventInvoicePaymentSucceeded:
		return s.handleInvoicePaymentSucceeded(ctx, event)
	case models.WebhookEventInvoicePaymentFailed:
		return s.handleInvoicePaymentFailed(ctx, event)
	default:
		// Unknown event type, just mark as processed
		return nil
	}
}

// handleSubscriptionCreated handles subscription.created webhook
func (s *BillingService) handleSubscriptionCreated(ctx context.Context, event *models.WebhookEvent) error {
	// Extract subscription data from webhook event
	// Update local subscription record
	return nil
}

// handleSubscriptionUpdated handles subscription.updated webhook
func (s *BillingService) handleSubscriptionUpdated(ctx context.Context, event *models.WebhookEvent) error {
	// Extract subscription data from webhook event
	// Update local subscription record
	return nil
}

// handleSubscriptionDeleted handles subscription.deleted webhook
func (s *BillingService) handleSubscriptionDeleted(ctx context.Context, event *models.WebhookEvent) error {
	// Extract subscription data from webhook event
	// Mark local subscription as cancelled/ended
	return nil
}

// handleInvoicePaymentSucceeded handles invoice.payment_succeeded webhook
func (s *BillingService) handleInvoicePaymentSucceeded(ctx context.Context, event *models.WebhookEvent) error {
	// Create payment record
	// Update subscription status if needed
	return nil
}

// handleInvoicePaymentFailed handles invoice.payment_failed webhook
func (s *BillingService) handleInvoicePaymentFailed(ctx context.Context, event *models.WebhookEvent) error {
	// Create failed payment record
	// Update subscription status to past_due
	return nil
}

// ===== ADMIN METHODS =====

// GetSubscriptionStats retrieves subscription statistics (admin only)
func (s *BillingService) GetSubscriptionStats(ctx context.Context) (*models.SubscriptionStats, error) {
	stats := &models.SubscriptionStats{}
	
	// Get total subscriptions
	err := s.db.GetContext(ctx, &stats.TotalSubscriptions,
		"SELECT COUNT(*) FROM billing.subscriptions")
	if err != nil {
		return nil, fmt.Errorf("failed to get total subscriptions: %w", err)
	}
	
	// Get active subscriptions
	err = s.db.GetContext(ctx, &stats.ActiveSubscriptions,
		"SELECT COUNT(*) FROM billing.subscriptions WHERE status = 'active'")
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}
	
	// Get trial subscriptions
	err = s.db.GetContext(ctx, &stats.TrialSubscriptions,
		`SELECT COUNT(*) FROM billing.subscriptions 
		 WHERE status = 'active' AND trial_end IS NOT NULL AND trial_end > NOW()`)
	if err != nil {
		return nil, fmt.Errorf("failed to get trial subscriptions: %w", err)
	}
	
	// Get cancelled subscriptions
	err = s.db.GetContext(ctx, &stats.CancelledSubscriptions,
		"SELECT COUNT(*) FROM billing.subscriptions WHERE status = 'cancelled'")
	if err != nil {
		return nil, fmt.Errorf("failed to get cancelled subscriptions: %w", err)
	}
	
	// Calculate MRR and ARR (simplified calculation)
	mrrQuery := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN s.billing_cycle = 'monthly' AND p.price_monthly IS NOT NULL THEN p.price_monthly
				WHEN s.billing_cycle = 'yearly' AND p.price_yearly IS NOT NULL THEN p.price_yearly / 12
				ELSE 0
			END
		), 0)
		FROM billing.subscriptions s
		JOIN billing.subscription_plans p ON s.plan_id = p.id
		WHERE s.status = 'active'
	`
	err = s.db.GetContext(ctx, &stats.MonthlyRecurringRevenue, mrrQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate MRR: %w", err)
	}
	
	stats.AnnualRecurringRevenue = stats.MonthlyRecurringRevenue * 12
	
	// Calculate churn rate (simplified - last 30 days)
	churnQuery := `
		SELECT 
			COALESCE(
				(SELECT COUNT(*) FROM billing.subscriptions WHERE cancelled_at >= NOW() - INTERVAL '30 days')::float /
				NULLIF((SELECT COUNT(*) FROM billing.subscriptions WHERE created_at <= NOW() - INTERVAL '30 days'), 0),
				0
			) * 100
	`
	err = s.db.GetContext(ctx, &stats.ChurnRate, churnQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate churn rate: %w", err)
	}
	
	// Calculate ARPU
	if stats.ActiveSubscriptions > 0 {
		stats.AverageRevenuePerUser = stats.MonthlyRecurringRevenue / float64(stats.ActiveSubscriptions)
	}
	
	return stats, nil
}

// ===== HELPER FUNCTIONS =====

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}