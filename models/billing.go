package irminmodels

import "time"

// PlanTier represents the billing plan tier.
type PlanTier string

const (
	// PlanTierHobby represents the hobby plan tier.
	PlanTierHobby PlanTier = "hobby"
	// PlanTierPro represents the pro plan tier.
	PlanTierPro PlanTier = "pro"
	// PlanTierBusiness represents the business plan tier.
	PlanTierBusiness PlanTier = "business"
	// PlanTierEnterprise represents the enterprise plan tier.
	PlanTierEnterprise PlanTier = "enterprise"
)

// BillingInterval represents the billing interval.
type BillingInterval string

const (
	// BillingIntervalMonthly represents monthly billing.
	BillingIntervalMonthly BillingInterval = "monthly"
	// BillingIntervalAnnual represents annual billing.
	BillingIntervalAnnual BillingInterval = "annual"
)

// SubscriptionStatus represents the status of a subscription.
type SubscriptionStatus string

const (
	// SubscriptionStatusActive represents an active subscription.
	SubscriptionStatusActive SubscriptionStatus = "active"
	// SubscriptionStatusCancelled represents a cancelled subscription.
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	// SubscriptionStatusPastDue represents a past due subscription.
	SubscriptionStatusPastDue SubscriptionStatus = "past_due"
	// SubscriptionStatusTrialing represents a trialing subscription.
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	// SubscriptionStatusNone represents no active subscription.
	SubscriptionStatusNone SubscriptionStatus = "none"
)

// UsageDimension represents a usage tracking dimension.
type UsageDimension string

const (
	// UsageDimensionStorage represents storage usage.
	UsageDimensionStorage UsageDimension = "storage"
	// UsageDimensionWorkflowRuns represents workflow run usage.
	UsageDimensionWorkflowRuns UsageDimension = "workflow_runs"
	// UsageDimensionAIRequests represents AI request usage.
	UsageDimensionAIRequests UsageDimension = "ai_requests"
	// UsageDimensionAPIRequests represents API request usage.
	UsageDimensionAPIRequests UsageDimension = "api_requests"
	// UsageDimensionDataTransfer represents data transfer usage.
	UsageDimensionDataTransfer UsageDimension = "data_transfer"
)

// PlanInfo holds information about a workspace's current plan.
type PlanInfo struct {
	Tier            PlanTier           `json:"tier"`
	Status          SubscriptionStatus `json:"status"`
	BillingInterval BillingInterval    `json:"billing_interval"`
	PeriodStart     *time.Time         `json:"current_period_start"`
	PeriodEnd       *time.Time         `json:"current_period_end"`
	CancelledAt     *time.Time         `json:"cancelled_at"`
}

// UsageDimensionInfo holds usage info for a single dimension.
type UsageDimensionInfo struct {
	Dimension    UsageDimension `json:"dimension"`
	CurrentUsage int64          `json:"current_usage"`
	Limit        *int64         `json:"limit"`
	Unit         string         `json:"unit"`
	RatePerUnit  float64        `json:"rate_per_unit"`
}

// CheckoutRequest is the request body for creating a checkout session.
type CheckoutRequest struct {
	PlanTier  string `json:"plan_tier"  validate:"required"`
	ReturnURL string `json:"return_url" validate:"required"`
}

// CheckoutResponse is the response body for checkout/change-plan endpoints.
type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}

// PortalResponse is the response body for the portal endpoint.
type PortalResponse struct {
	PortalURL string `json:"portal_url"`
}
