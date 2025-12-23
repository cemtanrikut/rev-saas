package config

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	AppPort      string
	MongoURI     string
	MongoDB      string
	JWTSecret    string
	OpenAIAPIKey string

	// Stripe Connect OAuth (for analytics - connecting customer Stripe accounts)
	StripeSecretKey          string
	StripeConnectClientID    string
	StripeConnectRedirectURL string
	AppPublicURL             string
	StripeLivemode           bool // derived from secret key prefix

	// Stripe Billing (for Revalyze's own payments)
	StripeBillingSecretKey       string // Can be same as StripeSecretKey or separate
	StripeWebhookSecret          string
	StripeBillingSuccessURL      string
	StripeBillingCancelURL       string
	StripeCustomerPortalReturnURL string
	// Price IDs for subscription plans
	StripePriceStarterID    string
	StripePriceGrowthID     string
	StripePriceEnterpriseID string

	// Encryption key for sensitive data (32 bytes for AES-256)
	EncryptionKey []byte
}

// Load reads configuration from environment variables with sensible defaults.
// It first tries to load a .env file if present (for local development).
func Load() *Config {
	// Load .env file if it exists (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	} else {
		log.Println("Loaded configuration from .env file")
	}

	encKeyStr := getEnv("ENCRYPTION_KEY", "")
	var encKey []byte
	if encKeyStr != "" {
		var err error
		encKey, err = hex.DecodeString(encKeyStr)
		if err != nil || len(encKey) != 32 {
			log.Printf("Warning: ENCRYPTION_KEY invalid (should be 64 hex chars / 32 bytes). Stripe token encryption disabled.")
			encKey = nil
		}
	}

	stripeSecretKey := getEnv("STRIPE_SECRET_KEY", "")
	stripeLivemode := strings.HasPrefix(stripeSecretKey, "sk_live_")

	// Stripe Billing secret key - fallback to main secret key if not set separately
	stripeBillingSecretKey := getEnv("STRIPE_BILLING_SECRET_KEY", stripeSecretKey)

	cfg := &Config{
		AppPort:      getEnv("APP_PORT", "8080"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:      getEnv("MONGO_DB_NAME", "rev_saas"),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-change-me"),
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		// Stripe Connect (analytics)
		StripeSecretKey:          stripeSecretKey,
		StripeConnectClientID:    getEnv("STRIPE_CONNECT_CLIENT_ID", ""),
		StripeConnectRedirectURL: getEnv("STRIPE_CONNECT_REDIRECT_URL", ""),
		AppPublicURL:             getEnv("APP_PUBLIC_URL", "http://localhost:5173"),
		StripeLivemode:           stripeLivemode,

		// Stripe Billing (payments)
		StripeBillingSecretKey:        stripeBillingSecretKey,
		StripeWebhookSecret:           getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeBillingSuccessURL:       getEnv("STRIPE_BILLING_SUCCESS_URL", "http://localhost:5173/app/settings/billing?success=1"),
		StripeBillingCancelURL:        getEnv("STRIPE_BILLING_CANCEL_URL", "http://localhost:5173/app/settings/billing?canceled=1"),
		StripeCustomerPortalReturnURL: getEnv("STRIPE_CUSTOMER_PORTAL_RETURN_URL", "http://localhost:5173/app/settings/billing"),
		StripePriceStarterID:          getEnv("STRIPE_PRICE_STARTER_ID", ""),
		StripePriceGrowthID:           getEnv("STRIPE_PRICE_GROWTH_ID", ""),
		StripePriceEnterpriseID:       getEnv("STRIPE_PRICE_ENTERPRISE_ID", ""),

		EncryptionKey: encKey,
	}

	stripeEnabled := cfg.StripeSecretKey != "" && cfg.StripeConnectClientID != ""
	stripeMode := "disabled"
	if stripeEnabled {
		if stripeLivemode {
			stripeMode = "LIVE"
		} else {
			stripeMode = "test"
		}
	}

	log.Printf("Config loaded: port=%s, mongo_db=%s, openai_enabled=%v, stripe_mode=%s",
		cfg.AppPort, cfg.MongoDB, cfg.OpenAIAPIKey != "", stripeMode)

	return cfg
}

// ValidateStripeConfig validates Stripe configuration and returns an error if invalid.
// Call this during startup to fail fast with clear error messages.
func (c *Config) ValidateStripeConfig() error {
	// If no Stripe keys are set, Stripe is simply disabled (not an error)
	if c.StripeSecretKey == "" && c.StripeConnectClientID == "" {
		log.Println("Stripe Connect: DISABLED (no credentials configured)")
		return nil
	}

	var errors []string

	// Validate STRIPE_SECRET_KEY
	if c.StripeSecretKey == "" {
		errors = append(errors, "STRIPE_SECRET_KEY is required when Stripe is enabled")
	} else if !strings.HasPrefix(c.StripeSecretKey, "sk_test_") && !strings.HasPrefix(c.StripeSecretKey, "sk_live_") {
		errors = append(errors, "STRIPE_SECRET_KEY must start with 'sk_test_' or 'sk_live_'")
	}

	// Validate STRIPE_CONNECT_CLIENT_ID
	if c.StripeConnectClientID == "" {
		errors = append(errors, "STRIPE_CONNECT_CLIENT_ID is required when Stripe is enabled")
	} else if !strings.HasPrefix(c.StripeConnectClientID, "ca_") {
		errors = append(errors, "STRIPE_CONNECT_CLIENT_ID must start with 'ca_' (get this from Stripe Dashboard > Connect > Settings)")
	}

	// Validate STRIPE_CONNECT_REDIRECT_URL
	if c.StripeConnectRedirectURL == "" {
		errors = append(errors, "STRIPE_CONNECT_REDIRECT_URL is required when Stripe is enabled")
	} else if !strings.HasPrefix(c.StripeConnectRedirectURL, "http://") && !strings.HasPrefix(c.StripeConnectRedirectURL, "https://") {
		errors = append(errors, "STRIPE_CONNECT_REDIRECT_URL must be a full URL (e.g., http://localhost:8080/api/stripe/connect/callback)")
	}

	// Validate APP_PUBLIC_URL
	if c.AppPublicURL == "" {
		errors = append(errors, "APP_PUBLIC_URL is required when Stripe is enabled")
	}

	// Validate ENCRYPTION_KEY for token storage
	if c.EncryptionKey == nil || len(c.EncryptionKey) != 32 {
		errors = append(errors, "ENCRYPTION_KEY is required for Stripe token encryption (must be 64 hex chars / 32 bytes)")
	}

	if len(errors) > 0 {
		return fmt.Errorf("stripe configuration errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	// Log successful validation (without secrets)
	mode := "TEST"
	if strings.HasPrefix(c.StripeSecretKey, "sk_live_") {
		mode = "LIVE"
	}
	clientIDPrefix := c.StripeConnectClientID
	if len(clientIDPrefix) > 10 {
		clientIDPrefix = clientIDPrefix[:10] + "..."
	}

	log.Printf("Stripe Connect: ENABLED mode=%s client_id_prefix=%s redirect_url=%s",
		mode, clientIDPrefix, c.StripeConnectRedirectURL)

	return nil
}

// ValidateStripeBillingConfig validates Stripe Billing configuration.
// Returns error if critical config is missing when billing is expected to be enabled.
func (c *Config) ValidateStripeBillingConfig() error {
	// If no billing secret key, billing is disabled
	if c.StripeBillingSecretKey == "" {
		log.Println("Stripe Billing: DISABLED (no STRIPE_BILLING_SECRET_KEY)")
		return nil
	}

	var errors []string

	// Validate secret key format
	if !strings.HasPrefix(c.StripeBillingSecretKey, "sk_test_") && !strings.HasPrefix(c.StripeBillingSecretKey, "sk_live_") {
		errors = append(errors, "STRIPE_BILLING_SECRET_KEY must start with 'sk_test_' or 'sk_live_'")
	}

	// Webhook secret is critical for billing
	if c.StripeWebhookSecret == "" {
		errors = append(errors, "STRIPE_WEBHOOK_SECRET is required for Stripe Billing (get from Stripe Dashboard > Webhooks)")
	} else if !strings.HasPrefix(c.StripeWebhookSecret, "whsec_") {
		errors = append(errors, "STRIPE_WEBHOOK_SECRET must start with 'whsec_'")
	}

	// Price IDs - warn if not set but don't fail
	if c.StripePriceStarterID == "" || c.StripePriceGrowthID == "" || c.StripePriceEnterpriseID == "" {
		log.Println("Warning: Some STRIPE_PRICE_*_ID env vars not set. Checkout for those plans will fail.")
	}

	if len(errors) > 0 {
		return fmt.Errorf("stripe billing configuration errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	mode := "TEST"
	if strings.HasPrefix(c.StripeBillingSecretKey, "sk_live_") {
		mode = "LIVE"
	}

	log.Printf("Stripe Billing: ENABLED mode=%s webhook_configured=%v",
		mode, c.StripeWebhookSecret != "")

	return nil
}

// IsBillingEnabled returns true if Stripe Billing is configured.
func (c *Config) IsBillingEnabled() bool {
	return c.StripeBillingSecretKey != "" && c.StripeWebhookSecret != ""
}

// GetPriceIDForPlan returns the Stripe price ID for a given plan key.
func (c *Config) GetPriceIDForPlan(planKey string) string {
	switch planKey {
	case "starter":
		return c.StripePriceStarterID
	case "growth":
		return c.StripePriceGrowthID
	case "enterprise":
		return c.StripePriceEnterpriseID
	default:
		return ""
	}
}

// GetPlanKeyForPriceID returns the plan key for a given Stripe price ID.
func (c *Config) GetPlanKeyForPriceID(priceID string) string {
	switch priceID {
	case c.StripePriceStarterID:
		return "starter"
	case c.StripePriceGrowthID:
		return "growth"
	case c.StripePriceEnterpriseID:
		return "enterprise"
	default:
		return "free"
	}
}

// getEnv retrieves an environment variable or returns a fallback value.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

