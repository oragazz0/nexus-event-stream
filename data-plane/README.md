# Nexus Data Plane

This directory contains the high-performance read service for the Nexus Event Stream system, written in Go.

## Purpose

The Data Plane consumes domain events from Redpanda and builds a **materialized view** in Redis, providing a fast, denormalized read path. It implements the query side of a CQRS architecture:

- **Event Consumption**: Listens to `nexus.signals` topic and projects every create, update, and delete into Redis.
- **Read API**: Exposes a low-latency REST API that serves signal data directly from the materialized view, without touching PostgreSQL.

## Architecture

The service runs two concurrent workloads in a single binary:

1. **Kafka Consumer** — reads events, applies them to the Redis projection, commits offsets only after a successful write.
2. **HTTP Server** — serves read queries from the Redis projection.

```
Redpanda ──▶ Consumer ──▶ Redis ◀── HTTP Server ──▶ Clients
```

### Modules

#### `internal/domain`
Domain types shared across the service.
- **`SignalEvent`**: Represents an event received from the `nexus.signals` topic. Carries an `Action` field (`created`, `updated`, `deleted`) plus the signal payload.
- **`Signal`**: The read model struct served by the API.
- **`ParseSignalEvent`**: Deserializes a raw Kafka message into a `SignalEvent`.
- **`SignalFromMap`**: Builds a `Signal` from a Redis hash result.

#### `internal/projection`
Owns the entire Redis data model — both writes and reads.
- **`Apply`**: Routes an event to `upsert` or `evict` based on its action.
- **`upsert`**: Stores/updates the signal hash and its sorted set indices (by creation time and by priority) in a single atomic transaction.
- **`evict`**: Removes the signal hash and all index entries atomically.
- **`ListByCreatedAt`**: Returns signals ordered by newest first, using a pipelined batch fetch.
- **`ListByPriority`**: Returns signals filtered by a specific priority level.
- **`FindByID`**: Returns a single signal by its UUID.
- **`Health`**: Pings Redis for liveness checks.

#### `internal/consumer`
Kafka consumer loop with manual offset management.
- **`Start`**: Blocks and processes messages until the context is cancelled.
- **`processNext`**: Fetches a message, parses it, applies the projection, and commits the offset. Malformed messages are skipped; projection failures trigger retry with backoff.
- **`applyWithRetry`**: Retries the Redis write indefinitely (1s interval) until success or context cancellation.

#### `internal/handler`
HTTP read API using Go's stdlib `net/http` with 1.22+ method routing.
- **`Register`**: Mounts all routes on a `ServeMux`.
- **`listSignals`**: Lists signals, optionally filtered by `?priority=`.
- **`getSignal`**: Returns a single signal by ID.
- **`health`**: Returns Redis liveness status.

#### `cmd/server`
Application entry point.
- Initializes a signal-aware context for graceful shutdown.
- Connects to Redis and validates the connection.
- Starts the Kafka consumer in a background goroutine.
- Blocks on the HTTP server until shutdown.

## Development

### Prerequisites
- Go 1.25+
- Docker (for Redis and Redpanda)

### Quick Start

```bash
# Start infrastructure (Redis, Redpanda, PostgreSQL)
docker compose -f ../infrastructure/docker-compose.yml up -d

# Install dependencies
go mod tidy

# Run the service
go run ./cmd/server
```

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `REDIS_ADDR` | `localhost:6379` | Redis connection address |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated Kafka broker addresses |
| `HTTP_ADDR` | `:8081` | HTTP server listen address |

### API Endpoints

Base URL: `http://localhost:8081`

| Method | Path | Description |
|---|---|---|
| `GET` | `/signals` | List all signals (newest first, max 50) |
| `GET` | `/signals?priority=High` | List signals filtered by priority (`Low`, `Medium`, `High`) |
| `GET` | `/signals/{id}` | Get a single signal by UUID |
| `GET` | `/health` | Redis liveness check |

### Redis Data Model

Each signal is stored as a Redis Hash with two sorted set indices:

```
signal:{uuid}              → Hash   (id, title, content, priority, author, timestamps)
signals:by_created_at      → ZSet   (score = unix timestamp, member = uuid)
signals:by_priority         → ZSet   (score = 1|2|3, member = uuid)
```

## Edge Cases (TODO)

> To be tested and implemented in future iterations.

| Scenario | Current Behavior | Planned Strategy |
|---|---|---|
| **Redis is down during consumption** | Consumer retries indefinitely with 1s backoff; offset is not committed, so no data loss. | Add exponential backoff and a circuit breaker to avoid log flooding. |
| **Cold start (empty Redis, existing events)** | Consumer group starts from `earliest`, replaying the full topic to rebuild the view. | Validate with integration tests; consider a `/rebuild` admin endpoint to trigger manual replay. |
| **Out-of-order events** | Not an issue today — single partition guarantees ordering per key. | If partitions scale, ensure signal ID is the partition key (already the Kafka message key) and add last-write-wins timestamp checks. |
| **Event schema evolution** | `json.Unmarshal` ignores unknown fields; missing fields get Go zero values. | Add explicit schema versioning to the event payload and handle migration in the consumer. |
