# Revalyze API

Go backend for the Revalyze SaaS application.

## Prerequisites

- Go 1.21+
- MongoDB (local or remote)

## Environment Separation

This project supports three environments: **local**, **staging**, and **production**.

| Environment | APP_ENV | Database | Stripe Keys |
|-------------|---------|----------|-------------|
| Local | `local` | `rev_saas` | `sk_test_*` |
| Staging | `staging` | `rev_saas_staging` | `sk_test_*` |
| Production | `production` | `rev_saas_prod` | `sk_live_*` |

For detailed documentation, see [Environment Separation Guide](../docs/ENVIRONMENTS.md).

### Quick Start

```bash
# Copy the template for local development
cp env-templates/local.env.template .env.local

# Edit with your values
nano .env.local

# Run the server (defaults to APP_ENV=local)
go run cmd/server/main.go
```

## Project Structure

```
api/
├── cmd/
│   └── server/
│       └── main.go          # Application entrypoint
├── internal/
│   ├── config/
│   │   └── config.go        # Environment-based configuration
│   ├── handler/
│   │   └── health_handler.go # HTTP handlers
│   ├── middleware/
│   │   └── doc.go           # Placeholder for future middleware
│   ├── model/
│   │   └── doc.go           # Placeholder for domain models
│   ├── repository/
│   │   └── mongo/
│   │       └── mongo.go     # MongoDB client wrapper
│   ├── router/
│   │   └── router.go        # HTTP router configuration
│   └── service/
│       └── doc.go           # Placeholder for business logic
├── go.mod
├── go.sum
└── README.md
```

## Configuration

The API uses environment variables for configuration:

### Core Settings

| Variable        | Default                       | Description           |
|-----------------|-------------------------------|-----------------------|
| `APP_ENV`       | `local`                       | Environment: `local`, `staging`, `production` |
| `APP_PORT`      | `8080`                        | HTTP server port      |
| `MONGO_URI`     | `mongodb://localhost:27017`   | MongoDB connection URI|
| `MONGO_DB_NAME` | `rev_saas`                    | Base MongoDB database name (suffixed by env) |
| `JWT_SECRET`    | `dev-secret-change-me`        | JWT signing secret    |
| `OPENAI_API_KEY`| -                             | OpenAI API key        |

### Stripe Connect Settings (Optional)

| Variable                     | Description                                      |
|------------------------------|--------------------------------------------------|
| `STRIPE_SECRET_KEY`          | Stripe secret key (`sk_test_...` or `sk_live_...`) |
| `STRIPE_CONNECT_CLIENT_ID`   | Stripe Connect client ID (`ca_...`)              |
| `STRIPE_CONNECT_REDIRECT_URL`| Full callback URL (e.g., `http://localhost:8080/api/stripe/connect/callback`) |
| `APP_PUBLIC_URL`             | Frontend URL for redirects (e.g., `http://localhost:5173`) |
| `ENCRYPTION_KEY`             | 32-byte hex string for token encryption (64 hex chars) |

---

## Stripe Connect Setup (Required for Stripe Integration)

Follow these steps to enable Stripe Connect OAuth:

### 1. Enable Stripe Connect in Dashboard

1. Go to [Stripe Dashboard](https://dashboard.stripe.com)
2. Navigate to **Settings → Connect → Get started**
3. Choose **Standard** account type
4. Complete the Connect onboarding

### 2. Get Your Client ID

1. Go to **Settings → Connect → Settings**
2. Find your **Client ID** (starts with `ca_`)
3. Copy it to `STRIPE_CONNECT_CLIENT_ID`

> **Note**: The Client ID is the same for test and live mode.

### 3. Configure Redirect URLs

1. Go to **Settings → Connect → Settings → Redirects**
2. Add your callback URLs to the whitelist:

**Local Development:**
```
http://localhost:8080/api/stripe/connect/callback
```

**Production:**
```
https://api.yourdomain.com/api/stripe/connect/callback
```

> **IMPORTANT**: The redirect URL must match `STRIPE_CONNECT_REDIRECT_URL` **EXACTLY** — including scheme (http/https), host, port, and path.

### 4. Get Your Secret Key

1. Go to **Developers → API keys**
2. Copy your **Secret key**:
   - Test mode: `sk_test_...`
   - Live mode: `sk_live_...`
3. Set it as `STRIPE_SECRET_KEY`

### 5. Generate Encryption Key

Generate a 32-byte (64 hex character) encryption key for secure token storage:

```bash
openssl rand -hex 32
```

Set the output as `ENCRYPTION_KEY`.

### 6. Example Configuration

**Local Development (.env):**
```bash
STRIPE_SECRET_KEY=sk_test_51ABC...
STRIPE_CONNECT_CLIENT_ID=ca_ABC123...
STRIPE_CONNECT_REDIRECT_URL=http://localhost:8080/api/stripe/connect/callback
APP_PUBLIC_URL=http://localhost:5173
ENCRYPTION_KEY=a1b2c3d4e5f6...  # 64 hex chars
```

**Production:**
```bash
STRIPE_SECRET_KEY=sk_live_51ABC...
STRIPE_CONNECT_CLIENT_ID=ca_ABC123...
STRIPE_CONNECT_REDIRECT_URL=https://api.yourdomain.com/api/stripe/connect/callback
APP_PUBLIC_URL=https://app.yourdomain.com
ENCRYPTION_KEY=a1b2c3d4e5f6...  # 64 hex chars
```

### Test vs Live Mode Consistency

- **Test mode**: Use `sk_test_...` secret key
- **Live mode**: Use `sk_live_...` secret key
- The Client ID (`ca_...`) works for both modes

### Troubleshooting

#### 404 or "Page not found" after clicking Connect

1. **Check redirect URL whitelist**: The URL in `STRIPE_CONNECT_REDIRECT_URL` must be **exactly** whitelisted in Stripe Dashboard → Connect → Redirects
2. **Check Connect is enabled**: Ensure Stripe Connect is enabled in your dashboard
3. **Check Client ID**: Must start with `ca_` (not `pk_` or `sk_`)

#### "invalid_request" error

1. Copy the generated connect URL from browser address bar
2. Check the URL parameters:
   - `client_id` should be `ca_...`
   - `redirect_uri` should match your whitelist exactly
3. Check Stripe error message in the URL

#### Callback returns error

Check server logs for debug output:
```
[stripe] connect URL generated user=xxx redirect=http://... client_id_prefix=ca_ABC123... livemode=false
```

#### Token encryption errors

Ensure `ENCRYPTION_KEY` is exactly 64 hex characters (32 bytes):
```bash
echo $ENCRYPTION_KEY | wc -c  # Should output 65 (64 chars + newline)
```

## Running the Server

### Local Development

```bash
cd api
go run ./cmd/server
```

### With Custom Configuration

```bash
APP_PORT=3001 MONGO_URI=mongodb://localhost:27017 go run ./cmd/server
```

## API Endpoints

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{"status":"ok"}
```

## Development

### Install Dependencies

```bash
cd api
go mod download
```

### Build

```bash
go build -o bin/server ./cmd/server
```

### Run Built Binary

```bash
./bin/server
```

## Future Endpoints

The following endpoints will be added:

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/plans` - List pricing plans
- `POST /api/v1/plans` - Create a pricing plan
- `GET /api/v1/competitors` - List competitors
- `POST /api/v1/competitors` - Add a competitor
- `GET /api/v1/analyses` - List pricing analyses
- `POST /api/v1/analyses/run` - Run a pricing analysis




