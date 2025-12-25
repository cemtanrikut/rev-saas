package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Environment represents the application environment.
type Environment string

const (
	EnvLocal      Environment = "local"
	EnvStaging    Environment = "staging"
	EnvProduction Environment = "production"
)

// String returns the string representation of the environment.
func (e Environment) String() string {
	return string(e)
}

// IsProduction returns true if the environment is production.
func (e Environment) IsProduction() bool {
	return e == EnvProduction
}

// IsStaging returns true if the environment is staging.
func (e Environment) IsStaging() bool {
	return e == EnvStaging
}

// IsLocal returns true if the environment is local.
func (e Environment) IsLocal() bool {
	return e == EnvLocal
}

// IsDevelopment returns true if the environment is local or staging (non-production).
func (e Environment) IsDevelopment() bool {
	return e == EnvLocal || e == EnvStaging
}

// GetEnvironment returns the current environment from APP_ENV.
// Defaults to "local" if not set.
func GetEnvironment() Environment {
	env := os.Getenv("APP_ENV")
	switch strings.ToLower(env) {
	case "production", "prod":
		return EnvProduction
	case "staging", "stg":
		return EnvStaging
	case "local", "development", "dev", "":
		return EnvLocal
	default:
		log.Printf("Warning: Unknown APP_ENV '%s', defaulting to 'local'", env)
		return EnvLocal
	}
}

// LoadEnvFile loads the appropriate .env file based on APP_ENV.
// Priority:
//  1. .env.[environment] (e.g., .env.local, .env.staging, .env.production)
//  2. .env (fallback for backwards compatibility)
//
// Environment variables already set take precedence over .env file values.
func LoadEnvFile() Environment {
	env := GetEnvironment()

	// Try environment-specific file first
	envFile := fmt.Sprintf(".env.%s", env)
	if err := godotenv.Load(envFile); err == nil {
		log.Printf("Loaded configuration from %s", envFile)
		return env
	}

	// Fallback to .env
	if err := godotenv.Load(); err == nil {
		log.Printf("Loaded configuration from .env (APP_ENV=%s)", env)
		return env
	}

	log.Printf("No .env file found, using environment variables (APP_ENV=%s)", env)
	return env
}

// GetMongoDBName returns the MongoDB database name for the given environment.
func GetMongoDBName(env Environment, baseDBName string) string {
	switch env {
	case EnvProduction:
		return baseDBName + "_prod"
	case EnvStaging:
		return baseDBName + "_staging"
	default:
		// Local uses base name (e.g., "rev_saas")
		return baseDBName
	}
}

// ValidateEnvironmentKeys validates that Stripe keys match the expected environment.
// Returns an error if there's a mismatch.
func ValidateEnvironmentKeys(env Environment, stripeSecretKey, webhookSecret string) error {
	if stripeSecretKey == "" {
		// Stripe not configured, skip validation
		return nil
	}

	isLiveKey := strings.HasPrefix(stripeSecretKey, "sk_live_")
	isTestKey := strings.HasPrefix(stripeSecretKey, "sk_test_")
	isLiveWebhook := strings.HasPrefix(webhookSecret, "whsec_") && !strings.Contains(webhookSecret, "test")
	isTestWebhook := webhookSecret == "" || strings.HasPrefix(webhookSecret, "whsec_")

	// Note: Stripe CLI webhook secrets all start with whsec_ regardless of mode
	// We can't reliably distinguish test vs live webhook secrets by prefix alone
	// The livemode check in webhooks is the authoritative validation

	switch env {
	case EnvProduction:
		if !isLiveKey {
			return fmt.Errorf("PRODUCTION environment requires sk_live_* key, got test key")
		}
		if isTestKey {
			return fmt.Errorf("PRODUCTION environment cannot use sk_test_* key")
		}
		log.Println("✓ Production environment: Using LIVE Stripe keys")

	case EnvStaging, EnvLocal:
		if isLiveKey {
			return fmt.Errorf("%s environment cannot use sk_live_* key (use sk_test_* instead)", env)
		}
		if !isTestKey && stripeSecretKey != "" {
			return fmt.Errorf("%s environment requires sk_test_* key", env)
		}
		if isTestKey {
			log.Printf("✓ %s environment: Using TEST Stripe keys", env)
		}
	}

	// Log webhook secret status (without revealing the secret)
	if webhookSecret != "" {
		log.Printf("✓ Webhook secret configured (%d chars)", len(webhookSecret))
	}

	// Suppress unused variable warning
	_ = isLiveWebhook
	_ = isTestWebhook

	return nil
}

