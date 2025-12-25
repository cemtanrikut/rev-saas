# Environment Separation Guide

This document explains how environment separation works in the Revalyze project, covering local development, staging, and production environments.

## Overview

| Environment | Branch | Database | Stripe Keys | Webhook Events |
|------------|--------|----------|-------------|----------------|
| **Local** | any | `rev_saas` | `sk_test_*` | Test only |
| **Staging** | `develop` | `rev_saas_staging` | `sk_test_*` | Test only |
| **Production** | `main` | `rev_saas_prod` | `sk_live_*` | Live only |

## Branch → Environment Mapping

```
main branch      → Production environment (LIVE Stripe keys)
develop branch   → Staging environment (TEST Stripe keys)
feature branches → No deployment (CI runs tests only)
```

## Backend Configuration

### APP_ENV Variable

Set `APP_ENV` to control the environment:

```bash
APP_ENV=local       # Default for development
APP_ENV=staging     # For staging deployment
APP_ENV=production  # For production deployment
```

### Environment File Loading

The backend loads configuration from environment-specific files:

| APP_ENV | Primary File | Fallback |
|---------|--------------|----------|
| `local` | `.env.local` | `.env` |
| `staging` | `.env.staging` | - |
| `production` | `.env.production` | - |

### MongoDB Database Names

Database names are automatically suffixed based on environment:

| APP_ENV | Base Name | Actual Database |
|---------|-----------|-----------------|
| `local` | `rev_saas` | `rev_saas` |
| `staging` | `rev_saas` | `rev_saas_staging` |
| `production` | `rev_saas` | `rev_saas_prod` |

### Stripe Key Validation

The application validates Stripe keys at startup:

| Environment | Required Keys | Will Reject |
|-------------|---------------|-------------|
| `local` | `sk_test_*` | `sk_live_*` |
| `staging` | `sk_test_*` | `sk_live_*` |
| `production` | `sk_live_*` | `sk_test_*` |

**The application will refuse to start** if there's a key mismatch.

### Webhook Livemode Validation

Stripe webhooks include a `livemode` field. The backend validates this:

| Environment | Expected `livemode` | Rejected Events |
|-------------|---------------------|-----------------|
| `local` | `false` | Live events |
| `staging` | `false` | Live events |
| `production` | `true` | Test events |

Events with mismatched `livemode` are rejected with a `400 Bad Request`.

## Frontend Configuration

### Vite Environment Files

| Mode | File | Usage |
|------|------|-------|
| Development | `.env.development` | `npm run dev` |
| Staging | `.env.staging` | `npm run build -- --mode staging` |
| Production | `.env.production` | `npm run build` |

### Frontend Environment Variables

```bash
# Only VITE_* variables are exposed to the frontend
VITE_API_URL=https://api.revalyze.com
VITE_STRIPE_PUBLISHABLE_KEY=pk_live_xxx  # Safe for frontend
VITE_APP_ENV=production
```

**Important:** Never put secret keys in frontend environment variables. Only publishable keys (`pk_*`) are safe for frontend use.

## GitHub Configuration

### Required Secrets per Environment

#### Staging Environment (GitHub Settings → Environments → staging)

| Secret | Description | Example |
|--------|-------------|---------|
| `STRIPE_SECRET_KEY` | Stripe Connect secret (TEST) | `sk_test_...` |
| `STRIPE_BILLING_SECRET_KEY` | Billing secret (TEST) | `sk_test_...` |
| `STRIPE_WEBHOOK_SECRET` | Webhook signing secret | `whsec_...` |
| `STRIPE_PUBLISHABLE_KEY` | Frontend publishable key | `pk_test_...` |
| `STRIPE_CONNECT_CLIENT_ID` | Connect OAuth client ID | `ca_...` |
| `STRIPE_PRICE_STARTER_ID` | Starter plan price ID | `price_...` |
| `STRIPE_PRICE_GROWTH_ID` | Growth plan price ID | `price_...` |
| `STRIPE_PRICE_ENTERPRISE_ID` | Enterprise plan price ID | `price_...` |
| `MONGO_URI` | MongoDB connection string | `mongodb+srv://...` |
| `JWT_SECRET` | JWT signing secret | Random 64+ chars |
| `ENCRYPTION_KEY` | AES-256 key (hex) | 64 hex characters |
| `OPENAI_API_KEY` | OpenAI API key | `sk-...` |

#### Production Environment (GitHub Settings → Environments → production)

Same secrets as staging, but with **LIVE** Stripe keys:

| Secret | Key Type |
|--------|----------|
| `STRIPE_SECRET_KEY` | `sk_live_...` |
| `STRIPE_BILLING_SECRET_KEY` | `sk_live_...` |
| `STRIPE_PUBLISHABLE_KEY` | `pk_live_...` |

### Setting Up GitHub Environments

1. Go to your repository → **Settings** → **Environments**
2. Create two environments: `staging` and `production`
3. For `production`:
   - Enable **Required reviewers** (optional but recommended)
   - Restrict to `main` branch only
4. Add secrets to each environment

### Branch Protection Rules

Recommended branch protection for `main`:

- ✅ Require pull request before merging
- ✅ Require status checks (backend-test, frontend-test)
- ✅ Require branches to be up to date
- ✅ Restrict who can push

## Stripe Test vs Live Separation

### Test Mode (Local & Staging)

- Use test API keys (`sk_test_*`, `pk_test_*`)
- Use test card numbers: `4242 4242 4242 4242`
- Create test products/prices in Stripe Dashboard (Test mode)
- Webhooks have `livemode: false`

### Live Mode (Production)

- Use live API keys (`sk_live_*`, `pk_live_*`)
- Real payments with real cards
- Create live products/prices in Stripe Dashboard (Live mode)
- Webhooks have `livemode: true`

### Webhook Endpoints

| Environment | Webhook URL | Stripe Dashboard |
|-------------|-------------|------------------|
| Local | `stripe listen --forward-to localhost:8080/api/billing/webhook` | N/A (CLI) |
| Staging | `https://staging-api.revalyze.com/api/billing/webhook` | Test mode |
| Production | `https://api.revalyze.com/api/billing/webhook` | Live mode |

## Quick Start

### Local Development

```bash
cd api
cp env-templates/local.env.template .env.local
# Edit .env.local with your values

# Start backend
go run cmd/server/main.go

# In another terminal, start Stripe webhook forwarding
stripe listen --forward-to localhost:8080/api/billing/webhook
```

### Deploying to Staging

```bash
git checkout develop
git pull origin develop
# Make changes
git push origin develop
# GitHub Actions will deploy to staging
```

### Deploying to Production

```bash
git checkout main
git merge develop
git push origin main
# GitHub Actions will deploy to production
```

## Troubleshooting

### "PRODUCTION environment requires sk_live_* keys"

You're trying to start the production environment with test keys. Use live keys.

### "staging environment cannot use sk_live_* key"

You're trying to use live keys in staging. Use test keys instead.

### "livemode mismatch" in webhook logs

A webhook event was rejected because its `livemode` doesn't match the environment. Check that your Stripe webhook endpoint is configured in the correct mode (Test vs Live).

### Database not found

Make sure your MongoDB URI is correct and the database exists:
- Local: `rev_saas`
- Staging: `rev_saas_staging`
- Production: `rev_saas_prod`

