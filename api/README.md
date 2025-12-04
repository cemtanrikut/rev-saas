# Revalyze API

Go backend for the Revalyze SaaS application.

## Prerequisites

- Go 1.21+
- MongoDB (local or remote)

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

| Variable        | Default                       | Description           |
|-----------------|-------------------------------|-----------------------|
| `APP_PORT`      | `8080`                        | HTTP server port      |
| `MONGO_URI`     | `mongodb://localhost:27017`   | MongoDB connection URI|
| `MONGO_DB_NAME` | `rev_saas`                    | MongoDB database name |

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

