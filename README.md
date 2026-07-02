# user-service

A Go/Gin microservice that manages user profiles, preferences, and activity history. Consumes Kafka events to auto-create profiles on registration and log subscription changes.

## Description

`user-service` is an internal microservice accessible only through the api-gateway. It provides CRUD operations for user profiles, a generic key/value preference system, and a paginated activity log. It also acts as a Kafka consumer — automatically creating a profile when a user registers and logging subscription changes.

Authentication is handled by the gateway — this service trusts the `X-User-ID` header injected by the gateway and rejects requests without it.

### Key design decisions
- **No authentication**: trusts `X-User-ID` from the api-gateway
- **DB migrations on startup**: runs automatically, no manual migration step needed
- **Kafka consumer groups**: uses `user-service` group ID for exactly-once delivery per instance
- **Activity logging middleware**: every API request is logged to the activity table
- **Kafka producer**: publishes `user.profile_updated` and `user.preferences_updated` events
- **Graceful shutdown**: handles SIGINT/SIGTERM, drains in-flight requests, closes Kafka connections

## Project Structure

```
user-service/
├── cmd/
│   └── main.go                 # Entry point — DB, Kafka, router, server, shutdown
├── internal/
│   ├── config/
│   │   └── config.go           # Environment variable loading with defaults
│   ├── database/
│   │   └── database.go         # MySQL connection + migration runner
│   ├── handler/
│   │   ├── profile.go          # Profile CRUD handler (Create, Get, Update)
│   │   ├── preference.go       # Preferences handler (GetAll, Upsert, Delete)
│   │   └── activity.go         # Activity log handler (GetByUserID with pagination)
│   ├── kafka/
│   │   ├── consumer.go         # Consumes user.registered + subscription.changed
│   │   └── producer.go         # Publishes events to any topic
│   ├── middleware/
│   │   └── middleware.go       # RequireUserID, RequestLogger, ActivityLogger
│   ├── model/
│   │   └── model.go            # Profile, Preference, ActivityLog structs
│   ├── repository/
│   │   ├── profile.go          # Profile DB operations
│   │   ├── preference.go       # Preference DB operations
│   │   └── activity.go         # Activity DB operations
│   └── service/
│       ├── profile.go          # Profile business logic + Kafka publish
│       ├── preference.go       # Preference business logic + Kafka publish
│       └── activity.go         # Activity query logic
├── migrations/                 # SQL migration files (auto-run on startup)
├── k8s/local/                  # Local Kubernetes manifests
│   ├── deployment.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── mysql.yaml
│   └── network-policy.yaml
├── .github/workflows/
│   ├── ci.yml                  # Lint (golangci-lint) + Test (MySQL) + Docker build/push + Trivy CVE scan
│   └── cd.yml                  # Deploy to local k8s, AWS EKS, Azure AKS
├── Dockerfile                  # Multi-stage: Go 1.22 builder → Alpine runtime
└── go.mod
```

## API Endpoints

All endpoints require `X-User-ID` header (injected by api-gateway).

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Kubernetes liveness/readiness probe |
| `POST` | `/api/profiles` | Create a new profile |
| `GET` | `/api/profiles/:user_id` | Get profile by user ID |
| `PUT` | `/api/profiles/:user_id` | Update profile (display_name, avatar_url, bio, language, timezone) |
| `GET` | `/api/profiles/:user_id/preferences` | Get all user preferences |
| `PUT` | `/api/profiles/:user_id/preferences` | Upsert a preference (key/value) |
| `DELETE` | `/api/profiles/:user_id/preferences/:key` | Delete a preference |
| `GET` | `/api/profiles/:user_id/activity` | Get activity history (supports `?limit=20&offset=0`) |

### Consumers
The following clients call this service via the api-gateway proxy:
- **frontend** (Vue 3 SPA) — `GET /api/profiles/{id}`, `PUT /api/profiles/{id}`, `GET /api/profiles/{id}/activity`
- **api-gateway** — forwards requests with `X-User-ID`, `X-User-Email`, `X-User-Name`, `X-User-Role` headers

## Models

### Profile
| Field | Type | Description |
|---|---|---|
| `user_id` | string (UUID) | Unique user identifier from gateway |
| `email` | string | User's email |
| `display_name` | string (nullable) | Display name |
| `avatar_url` | string (nullable) | Avatar URL |
| `bio` | string (nullable) | Short biography |
| `language` | string | Preferred language (default: `en`) |
| `timezone` | string | Timezone (default: `UTC`) |

### Preference
| Field | Type | Description |
|---|---|---|
| `user_id` | string | Owner |
| `key` | string | Preference name |
| `value` | string | Preference value |

### ActivityLog
| Field | Type | Description |
|---|---|---|
| `user_id` | string | Who performed the action |
| `type` | string | `profile_updated`, `preferences_updated`, `subscription_changed`, `api_request` |
| `metadata` | JSON string | Additional context |

## Kafka Integration

### Events Consumed

| Topic | Event | Action |
|---|---|---|
| `user.registered` | user.registered | Auto-create profile (if not exists) |
| `subscription.changed` | subscription.changed | Log activity with plan + status |

### Events Produced

| Topic | When |
|---|---|
| `user.profile_updated` | Profile updated via PUT |
| `user.preferences_updated` | Preference upserted or deleted |

## Getting Started

### Requirements
- Go 1.22+
- MySQL 8.0
- Kafka broker

### Installation

```bash
cp .env.example .env
# Fill in your database and Kafka credentials

go mod download
```

### Run locally

```bash
go run cmd/main.go
# → http://localhost:8080
```

### Tests

```bash
go test ./... -v
```

### Lint

```bash
golangci-lint run
```

## Local Kubernetes Deployment

```bash
# Build
docker build -t user-service:latest .
docker save user-service:latest | docker exec -i $(docker ps -qf "name=desktop-control-plane") ctr -n k8s.io images import -

# Deploy
kubectl apply -f k8s/local/

# Verify
kubectl get pods -l app=user-service
kubectl logs -l app=user-service
```

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `APP_PORT` | Server port | `8080` |
| `APP_ENV` | Environment (`local`, `production`) | `production` |
| `DB_HOST` | MySQL host | `127.0.0.1` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_DATABASE` | Database name | `user_service` |
| `DB_USERNAME` | MySQL user | `root` |
| `DB_PASSWORD` | MySQL password | — |
| `KAFKA_BROKER` | Kafka broker address | `kafka:9092` |

## GitHub Actions Secrets Required

| Secret | Description |
|---|---|
| `DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN` | DockerHub credentials |
| `DB_PASSWORD` | user-service MySQL root password |
| `KUBECONFIG_LOCAL` | Local cluster kubeconfig (self-hosted runner) |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_REGION` | AWS credentials |
| `EKS_CLUSTER_NAME` | EKS cluster name |
| `AZURE_CREDENTIALS` | Azure service principal JSON |
| `AKS_CLUSTER_NAME` / `AKS_RESOURCE_GROUP` | AKS cluster info |
