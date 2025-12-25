package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionStatus represents the status of a Stripe subscription.
type SubscriptionStatus string

const (
	SubscriptionStatusTrialing   SubscriptionStatus = "trialing"
	SubscriptionStatusActive     SubscriptionStatus = "active"
	SubscriptionStatusPastDue    SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled   SubscriptionStatus = "canceled"
	SubscriptionStatusUnpaid     SubscriptionStatus = "unpaid"
	SubscriptionStatusIncomplete SubscriptionStatus = "incomplete"
)

// PlanKey represents the internal plan identifier.
type PlanKey string

const (
	PlanKeyFree       PlanKey = "free"
	PlanKeyStarter    PlanKey = "starter"
	PlanKeyGrowth     PlanKey = "growth"
	PlanKeyEnterprise PlanKey = "enterprise"
)

// BillingSubscription represents a user's Stripe subscription.
type BillingSubscription struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID               primitive.ObjectID `bson:"user_id" json:"user_id"`
	StripeCustomerID     string             `bson:"stripe_customer_id" json:"stripe_customer_id"`
	StripeSubscriptionID string             `bson:"stripe_subscription_id" json:"stripe_subscription_id"`
	StripePriceID        string             `bson:"stripe_price_id" json:"stripe_price_id"`
	PlanKey              PlanKey            `bson:"plan_key" json:"plan_key"`
	Status               SubscriptionStatus `bson:"status" json:"status"`
	CancelAtPeriodEnd    bool               `bson:"cancel_at_period_end" json:"cancel_at_period_end"`
	CurrentPeriodEnd     time.Time          `bson:"current_period_end" json:"current_period_end"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}

// IsActive returns true if the subscription is in an active state.
func (s *BillingSubscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive || s.Status == SubscriptionStatusTrialing
}

// IsPremium returns true if the subscription has a paid plan.
func (s *BillingSubscription) IsPremium() bool {
	return s.IsActive() && s.PlanKey != PlanKeyFree
}

// WebhookEvent tracks processed Stripe webhook events for idempotency.
type WebhookEvent struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EventID     string             `bson:"event_id" json:"event_id"` // Stripe evt_... ID
	EventType   string             `bson:"event_type" json:"event_type"`
	ProcessedAt time.Time          `bson:"processed_at" json:"processed_at"`
}

// BillingStatusResponse is the API response for billing status.
type BillingStatusResponse struct {
	PlanKey              PlanKey            `json:"plan_key"`
	Status               SubscriptionStatus `json:"status"`
	CancelAtPeriodEnd    bool               `json:"cancel_at_period_end"`
	CurrentPeriodEnd     *time.Time         `json:"current_period_end,omitempty"`
	StripeSubscriptionID string             `json:"stripe_subscription_id,omitempty"`
	StripeCustomerID     string             `json:"stripe_customer_id,omitempty"`
}

// CheckoutSessionRequest is the request body for creating a checkout session.
type CheckoutSessionRequest struct {
	PlanKey string `json:"plan_key"` // "starter", "growth", "enterprise"
}

// CheckoutSessionResponse is the response for checkout session creation.
type CheckoutSessionResponse struct {
	URL string `json:"url"`
}

// PortalSessionResponse is the response for portal session creation.
type PortalSessionResponse struct {
	URL string `json:"url"`
}



