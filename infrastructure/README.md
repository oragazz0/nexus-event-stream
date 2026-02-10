# Infrastructure

Local development infrastructure managed via Docker Compose.

## Services

### PostgreSQL

| Property  | Value              |
|-----------|--------------------|
| Image     | `postgres:18-alpine` |
| Container | `nexus-postgres`   |
| Port      | `5432`             |
| Restart   | `unless-stopped`   |

Default credentials:

| Variable            | Value   |
|---------------------|---------|
| `POSTGRES_DB`       | `nexus` |
| `POSTGRES_USER`     | `nexus` |
| `POSTGRES_PASSWORD` | `nexus` |

Data is persisted in the `nexus_postgres_data` named volume.

### Redpanda

Kafka-compatible streaming platform running in development mode.

| Property  | Value                                                      |
|-----------|------------------------------------------------------------|
| Image     | `docker.redpanda.com/redpandadata/redpanda:latest`         |
| Container | `redpanda`                                                 |
| Mode      | `dev-container`                                            |
| Resources | 1 CPU, 512M memory                                        |

**Ports:**

| Port    | Listener    | Usage                                      |
|---------|-------------|--------------------------------------------|
| `9092`  | `OUTSIDE`   | Host access (`localhost:9092`)              |
| `29092` | `PLAINTEXT` | Internal access between containers (`redpanda:29092`) |

Data is persisted in the `redpanda_data` named volume.

**Health check:** `rpk cluster health` — runs every 10s, 5s timeout, 5 retries.

### Redpanda Init

One-shot init container that creates default Kafka topics after Redpanda is healthy. Reuses the same Redpanda image to leverage `rpk`. Idempotent — safe to run on repeated `docker compose up` calls.

**Topics created:**

| Topic              | Partitions | Description                        |
|--------------------|------------|------------------------------------|
| `signals.created`  | 1          | Emitted when a Signal is created   |
| `signals.updated`  | 1          | Emitted when a Signal is updated   |
| `signals.deleted`  | 1          | Emitted when a Signal is deleted   |

### Redpanda Console

Web UI for inspecting topics, messages, and consumer groups.

| Property  | Value                                                      |
|-----------|------------------------------------------------------------|
| Image     | `docker.redpanda.com/redpandadata/console:latest`          |
| Container | `redpanda-console`                                         |
| Port      | `8080`                                                     |
| Broker    | `redpanda:29092`                                           |

Accessible at [http://localhost:8080](http://localhost:8080).

## Volumes

| Volume                | Used by    | Purpose                  |
|-----------------------|------------|--------------------------|
| `nexus_postgres_data` | PostgreSQL | Database file persistence |
| `redpanda_data`       | Redpanda   | Stream data persistence   |

## Usage

```bash
# Start all services
docker compose -f infrastructure/docker-compose.yml up -d

# Stop all services
docker compose -f infrastructure/docker-compose.yml down

# Stop and remove volumes (full reset)
docker compose -f infrastructure/docker-compose.yml down -v
```
