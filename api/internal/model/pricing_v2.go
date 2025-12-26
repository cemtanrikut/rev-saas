package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PricingV2Plan represents a plan extracted from a website
type PricingV2Plan struct {
	ID                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID                  primitive.ObjectID `json:"user_id" bson:"user_id"`
	WebsiteURL              string             `json:"website_url" bson:"website_url"`
	SourceURL               string             `json:"source_url" bson:"source_url"`
	ExtractedAt             time.Time          `json:"extracted_at" bson:"extracted_at"`
	PlanName                string             `json:"plan_name" bson:"plan_name"`
	PriceAmount             float64            `json:"price_amount" bson:"price_amount"`
	PriceString             string             `json:"price_string" bson:"price_string"`
	Currency                string             `json:"currency" bson:"currency"`
	PriceFrequency          string             `json:"price_frequency" bson:"price_frequency"`                     // per_month, per_year
	BillingPeriod           string             `json:"billing_period" bson:"billing_period"`                       // monthly, yearly, unknown
	MonthlyEquivalentAmount float64            `json:"monthly_equivalent_amount" bson:"monthly_equivalent_amount"` // for yearly plans
	AnnualBilledAmount      float64            `json:"annual_billed_amount" bson:"annual_billed_amount"`           // total yearly amount
	IncludedUnits           []IncludedUnit     `json:"included_units" bson:"included_units"`
	Features                []string           `json:"features" bson:"features"`
	Evidence                PlanEvidence       `json:"evidence" bson:"evidence"`
}

// IncludedUnit represents quantifiable included items like "7,500 credits/seat/month"
type IncludedUnit struct {
	Name     string  `json:"name" bson:"name"`         // e.g., "credits", "seats", "users"
	Amount   float64 `json:"amount" bson:"amount"`     // e.g., 7500
	Unit     string  `json:"unit" bson:"unit"`         // e.g., "per seat", "per month"
	RawText  string  `json:"raw_text" bson:"raw_text"` // original text
}

// PlanEvidence contains evidence snippets from the source
type PlanEvidence struct {
	NameSnippet     string `json:"name_snippet" bson:"name_snippet"`
	PriceSnippet    string `json:"price_snippet" bson:"price_snippet"`
	UnitsSnippet    string `json:"units_snippet" bson:"units_snippet"`
	BillingEvidence string `json:"billing_evidence" bson:"billing_evidence"` // evidence proving billing period
}

// PricingDiscoverRequest is the request for discover endpoint
type PricingDiscoverRequest struct {
	WebsiteURL string `json:"website_url"`
}

// PricingDiscoverResponse is the response from discover endpoint
type PricingDiscoverResponse struct {
	PricingCandidates  []string `json:"pricing_candidates"`
	SelectedPricingURL *string  `json:"selected_pricing_url"`
	Error              string   `json:"error,omitempty"`
}

// PricingExtractRequest is the request for extract endpoint
type PricingExtractRequest struct {
	PricingURL string `json:"pricing_url"`
}

// ExtractedPlan represents a plan extracted by LLM
type ExtractedPlan struct {
	Name                    string         `json:"name"`
	PriceAmount             float64        `json:"price_amount"`
	PriceString             string         `json:"price_string"`
	Currency                string         `json:"currency"`
	PriceFrequency          string         `json:"price_frequency"`           // per_month, per_year
	BillingPeriod           string         `json:"billing_period"`            // monthly, yearly
	MonthlyEquivalentAmount float64        `json:"monthly_equivalent_amount"` // for yearly plans showing "equivalent monthly"
	AnnualBilledAmount      float64        `json:"annual_billed_amount"`      // total yearly amount
	IncludedUnits           []IncludedUnit `json:"included_units"`
	Features                []string       `json:"features"`
	Evidence                PlanEvidence   `json:"evidence"`
}

// PricingExtractResponse is the response from extract endpoint
type PricingExtractResponse struct {
	Plans           []ExtractedPlan `json:"plans"`
	SourceURL       string          `json:"source_url"`
	DetectedPeriods []string        `json:"detected_periods"` // e.g., ["monthly", "yearly"]
	NeedsRender     bool            `json:"needs_render"`     // true if browser render needed for toggle
	RenderUsed      bool            `json:"render_used"`      // true if browser was used
	Warnings        []string        `json:"warnings,omitempty"`
	Error           string          `json:"error,omitempty"`
}

// PricingSaveRequest is the request for save endpoint
type PricingSaveRequest struct {
	Plans     []ExtractedPlan `json:"plans"`
	SourceURL string          `json:"source_url"`
	WebsiteURL string         `json:"website_url"`
}

// PricingSaveResponse is the response from save endpoint
type PricingSaveResponse struct {
	SavedCount int    `json:"saved_count"`
	Error      string `json:"error,omitempty"`
}

// SavedPricingV2Response is the response for listing saved v2 plans
type SavedPricingV2Response struct {
	Plans []PricingV2Plan `json:"plans"`
	Count int             `json:"count"`
}

// PricingExtractFromTextRequest is the request for extract-from-text endpoint (paste mode)
type PricingExtractFromTextRequest struct {
	MonthlyText string `json:"monthly_text"`
	YearlyText  string `json:"yearly_text"`
	WebsiteURL  string `json:"website_url,omitempty"` // optional context
}

// PricingExtractFromTextResponse is the response from extract-from-text endpoint
type PricingExtractFromTextResponse struct {
	Plans           []ExtractedPlan `json:"plans"`
	DetectedPeriods []string        `json:"detected_periods"`
	Warnings        []string        `json:"warnings,omitempty"`
	Error           string          `json:"error,omitempty"`
}
