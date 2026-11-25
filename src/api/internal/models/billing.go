package models

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionPlan represents a subscription plan
type SubscriptionPlan struct {
	ID                      uuid.UUID `json:"id" db:"id"`
	Name                    string    `json:"name" db:"name" validate:"required,min=2,max=255"`
	Description             *string   `json:"description,omitempty" db:"description"`
	PriceMonthly            *float64  `json:"price_monthly,omitempty" db:"price_monthly" validate:"omitempty,gt=0"`
	PriceYearly             *float64  `json:"price_yearly,omitempty" db:"price_yearly" validate:"omitempty,gt=0"`
	Currency                string    `json:"currency" db:"currency" validate:"required,len=3"`
	MaxGardens              *int      `json:"max_gardens,omitempty" db:"max_gardens" validate:"omitempty,gt=0"`
	MaxZonesPerGarden       *int      `json:"max_zones_per_garden,omitempty" db:"max_zones_per_garden" validate:"omitempty,gt=0"`
	MaxDevicesPerGarden     *int      `json:"max_devices_per_garden,omitempty" db:"max_devices_per_garden" validate:"omitempty,gt=0"`
	MaxAPICallsPerMonth     *int      `json:"max_api_calls_per_month,omitempty" db:"max_api_calls_per_month" validate:"omitempty,gt=0"`
	Features                Metadata  `json:"features,omitempty" db:"features"`
	IsActive                bool      `json:"is_active" db:"is_active"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

// Subscription represents a user's subscription
type Subscription struct {
	ID                    uuid.UUID        `json:"id" db:"id"`
	UserID                uuid.UUID        `json:"user_id" db:"user_id"`
	PlanID                uuid.UUID        `json:"plan_id" db:"plan_id"`
	StripeSubscriptionID  *string          `json:"stripe_subscription_id,omitempty" db:"stripe_subscription_id"`
	StripeCustomerID      *string          `json:"stripe_customer_id,omitempty" db:"stripe_customer_id"`
	Status                string           `json:"status" db:"status" validate:"oneof=active cancelled past_due unpaid"`
	BillingCycle          string           `json:"billing_cycle" db:"billing_cycle" validate:"oneof=monthly yearly"`
	StartedAt             time.Time        `json:"started_at" db:"started_at"`
	CurrentPeriodStart    time.Time        `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd      time.Time        `json:"current_period_end" db:"current_period_end"`
	CancelledAt           *time.Time       `json:"cancelled_at,omitempty" db:"cancelled_at"`
	EndedAt               *time.Time       `json:"ended_at,omitempty" db:"ended_at"`
	TrialStart            *time.Time       `json:"trial_start,omitempty" db:"trial_start"`
	TrialEnd              *time.Time       `json:"trial_end,omitempty" db:"trial_end"`
	UsageData             Metadata         `json:"usage_data,omitempty" db:"usage_data"`
	CreatedAt             time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at" db:"updated_at"`
	
	// Related data
	Plan     *SubscriptionPlan `json:"plan,omitempty"`
	Payments []Payment         `json:"payments,omitempty"`
}

// CreateSubscriptionRequest represents the request to create a subscription
type CreateSubscriptionRequest struct {
	PlanID       uuid.UUID `json:"plan_id" validate:"required"`
	BillingCycle string    `json:"billing_cycle" validate:"required,oneof=monthly yearly"`
	TrialDays    *int      `json:"trial_days,omitempty" validate:"omitempty,min=0,max=90"`
	PaymentMethod string   `json:"payment_method,omitempty"`
}

// UpdateSubscriptionRequest represents the request to update a subscription
type UpdateSubscriptionRequest struct {
	PlanID       *uuid.UUID `json:"plan_id,omitempty"`
	BillingCycle *string    `json:"billing_cycle,omitempty" validate:"omitempty,oneof=monthly yearly"`
	Status       *string    `json:"status,omitempty" validate:"omitempty,oneof=active cancelled past_due unpaid"`
}

// Payment represents a payment record
type Payment struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	SubscriptionID          uuid.UUID  `json:"subscription_id" db:"subscription_id"`
	StripePaymentIntentID   *string    `json:"stripe_payment_intent_id,omitempty" db:"stripe_payment_intent_id"`
	StripeInvoiceID         *string    `json:"stripe_invoice_id,omitempty" db:"stripe_invoice_id"`
	Amount                  float64    `json:"amount" db:"amount" validate:"required,gt=0"`
	Currency                string     `json:"currency" db:"currency" validate:"required,len=3"`
	Status                  string     `json:"status" db:"status" validate:"oneof=succeeded failed pending cancelled"`
	AttemptedAt             time.Time  `json:"attempted_at" db:"attempted_at"`
	SucceededAt             *time.Time `json:"succeeded_at,omitempty" db:"succeeded_at"`
	FailedAt                *time.Time `json:"failed_at,omitempty" db:"failed_at"`
	FailureCode             *string    `json:"failure_code,omitempty" db:"failure_code"`
	FailureMessage          *string    `json:"failure_message,omitempty" db:"failure_message"`
	Metadata                Metadata   `json:"metadata,omitempty" db:"metadata"`
}

// CreatePaymentRequest represents the request to create a payment
type CreatePaymentRequest struct {
	SubscriptionID uuid.UUID `json:"subscription_id" validate:"required"`
	Amount         float64   `json:"amount" validate:"required,gt=0"`
	Currency       string    `json:"currency" validate:"required,len=3"`
	PaymentMethod  string    `json:"payment_method" validate:"required"`
	Description    *string   `json:"description,omitempty"`
}

// SubscriptionUsage represents current usage statistics
type SubscriptionUsage struct {
	SubscriptionID     uuid.UUID `json:"subscription_id"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	APICallsUsed       int       `json:"api_calls_used"`
	APICallsLimit      *int      `json:"api_calls_limit"`
	GardensUsed        int       `json:"gardens_used"`
	GardensLimit       *int      `json:"gardens_limit"`
	DevicesUsed        int       `json:"devices_used"`
	DevicesLimit       *int      `json:"devices_limit"`
	StorageUsedMB      float64   `json:"storage_used_mb"`
	StorageLimitMB     *float64  `json:"storage_limit_mb"`
	OverageCharges     float64   `json:"overage_charges"`
	LastUpdated        time.Time `json:"last_updated"`
}

// BillingAddress represents a billing address
type BillingAddress struct {
	Line1      string  `json:"line1" validate:"required"`
	Line2      *string `json:"line2,omitempty"`
	City       string  `json:"city" validate:"required"`
	State      *string `json:"state,omitempty"`
	PostalCode string  `json:"postal_code" validate:"required"`
	Country    string  `json:"country" validate:"required,len=2"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID               string          `json:"id"`
	Type             string          `json:"type"` // card, bank_account, etc.
	Card             *CardDetails    `json:"card,omitempty"`
	BillingAddress   *BillingAddress `json:"billing_address,omitempty"`
	IsDefault        bool            `json:"is_default"`
	CreatedAt        time.Time       `json:"created_at"`
}

// CardDetails represents credit card details
type CardDetails struct {
	Brand    string `json:"brand"`
	Last4    string `json:"last4"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
	Country  string `json:"country"`
}

// Invoice represents an invoice
type Invoice struct {
	ID                string          `json:"id"`
	SubscriptionID    uuid.UUID       `json:"subscription_id"`
	Number            string          `json:"number"`
	Status            string          `json:"status"` // draft, open, paid, void, uncollectible
	Currency          string          `json:"currency"`
	AmountDue         float64         `json:"amount_due"`
	AmountPaid        float64         `json:"amount_paid"`
	AmountRemaining   float64         `json:"amount_remaining"`
	Subtotal          float64         `json:"subtotal"`
	Tax               float64         `json:"tax"`
	Total             float64         `json:"total"`
	PeriodStart       time.Time       `json:"period_start"`
	PeriodEnd         time.Time       `json:"period_end"`
	DueDate           *time.Time      `json:"due_date,omitempty"`
	PaidAt            *time.Time      `json:"paid_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	LineItems         []InvoiceLineItem `json:"line_items"`
	BillingAddress    *BillingAddress `json:"billing_address,omitempty"`
	HostedInvoiceURL  *string         `json:"hosted_invoice_url,omitempty"`
	InvoicePDF        *string         `json:"invoice_pdf,omitempty"`
}

// InvoiceLineItem represents a line item on an invoice
type InvoiceLineItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	UnitAmount  float64   `json:"unit_amount"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Metadata    Metadata  `json:"metadata,omitempty"`
}

// WebhookEvent represents a webhook event from Stripe
type WebhookEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Data      Metadata  `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	Processed bool      `json:"processed"`
}

// SubscriptionStats represents subscription statistics
type SubscriptionStats struct {
	TotalSubscriptions   int     `json:"total_subscriptions"`
	ActiveSubscriptions  int     `json:"active_subscriptions"`
	TrialSubscriptions   int     `json:"trial_subscriptions"`
	CancelledSubscriptions int   `json:"cancelled_subscriptions"`
	MonthlyRecurringRevenue float64 `json:"monthly_recurring_revenue"`
	AnnualRecurringRevenue  float64 `json:"annual_recurring_revenue"`
	ChurnRate              float64 `json:"churn_rate"`
	AverageRevenuePerUser  float64 `json:"average_revenue_per_user"`
}

// PlanFeature represents a feature available in a subscription plan
type PlanFeature struct {
	Key         string      `json:"key"`
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Type        string      `json:"type"` // boolean, numeric, text
	Value       interface{} `json:"value"`
	Limit       *int        `json:"limit,omitempty"`
}

// GetPlanFeatures returns the features for this plan
func (sp *SubscriptionPlan) GetPlanFeatures() []PlanFeature {
	features := []PlanFeature{}
	
	if sp.Features == nil {
		return features
	}
	
	for key, value := range sp.Features {
		feature := PlanFeature{
			Key:   key,
			Value: value,
		}
		features = append(features, feature)
	}
	
	return features
}

// IsFeatureEnabled checks if a feature is enabled in this plan
func (sp *SubscriptionPlan) IsFeatureEnabled(featureKey string) bool {
	if sp.Features == nil {
		return false
	}
	
	value, exists := sp.Features[featureKey]
	if !exists {
		return false
	}
	
	// Handle boolean features
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}
	
	// Handle numeric features (non-zero means enabled)
	if numValue, ok := value.(float64); ok {
		return numValue > 0
	}
	
	// Handle string features (non-empty means enabled)
	if strValue, ok := value.(string); ok {
		return strValue != ""
	}
	
	return false
}

// GetFeatureLimit returns the limit for a numeric feature
func (sp *SubscriptionPlan) GetFeatureLimit(featureKey string) *int {
	if sp.Features == nil {
		return nil
	}
	
	value, exists := sp.Features[featureKey]
	if !exists {
		return nil
	}
	
	if numValue, ok := value.(float64); ok {
		limit := int(numValue)
		return &limit
	}
	
	return nil
}

// IsTrialActive checks if the subscription is in an active trial period
func (s *Subscription) IsTrialActive() bool {
	if s.TrialStart == nil || s.TrialEnd == nil {
		return false
	}
	
	now := time.Now()
	return now.After(*s.TrialStart) && now.Before(*s.TrialEnd)
}

// DaysUntilTrialEnd returns days until trial ends (negative if ended)
func (s *Subscription) DaysUntilTrialEnd() *int {
	if s.TrialEnd == nil {
		return nil
	}
	
	days := int(time.Until(*s.TrialEnd).Hours() / 24)
	return &days
}

// IsActive checks if the subscription is currently active
func (s *Subscription) IsActive() bool {
	return s.Status == "active" && time.Now().Before(s.CurrentPeriodEnd)
}

// Constants for billing
const (
	// Subscription statuses
	SubscriptionStatusActive   = "active"
	SubscriptionStatusCancelled = "cancelled"
	SubscriptionStatusPastDue  = "past_due"
	SubscriptionStatusUnpaid   = "unpaid"
	
	// Billing cycles
	BillingCycleMonthly = "monthly"
	BillingCycleYearly  = "yearly"
	
	// Payment statuses
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusFailed    = "failed"
	PaymentStatusPending   = "pending"
	PaymentStatusCancelled = "cancelled"
	
	// Invoice statuses
	InvoiceStatusDraft         = "draft"
	InvoiceStatusOpen          = "open"
	InvoiceStatusPaid          = "paid"
	InvoiceStatusVoid          = "void"
	InvoiceStatusUncollectible = "uncollectible"
	
	// Webhook event types
	WebhookEventCustomerSubscriptionCreated = "customer.subscription.created"
	WebhookEventCustomerSubscriptionUpdated = "customer.subscription.updated"
	WebhookEventCustomerSubscriptionDeleted = "customer.subscription.deleted"
	WebhookEventInvoicePaymentSucceeded     = "invoice.payment_succeeded"
	WebhookEventInvoicePaymentFailed        = "invoice.payment_failed"
)