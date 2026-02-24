# Nexus Event Stream

This repository demonstrates concepts and the implementation of **Event-Driven Architecture (EDA)**, **Polyglot Microservices (Python + Go)**, and high-performance **Message Streaming**.

The project implements the **CQRS (Command-Query Responsibility Segregation)** pattern:
- **Django App** is the "Source of Truth" (Management Plane / Command Side) 
- **Go service** is the "Data Plane" (High-performance execution / Query Side)
- The goal is **Eventual Consistency** with low latency using an event broker and a fast key-value store.

## Architecture

![Architecture](docs/architecture.mmd)

The architecture splits responsibilities:
- **Operations (Commands):** Validated and persisted in the relational database by Django. A domain event is then published to the stream.
- **Data Distribution:** Redpanda (Kafka) acts as the event broker forwarding immutable events across microservices.
- **Reads (Queries):** The Go consumer projects the changed state into a Redis materialized view, making reads extremely fast without touching the core system database.

## Components

The system is split into three main parts:

### 1. Control Plane (Python/Django)
*Location: `./control-plane/`*
- The authoritative command center.
- Exposes a REST API for **Signal Management** (prioritized directives) and **Identity Management**.
- Uses **PostgreSQL** for durable, normalized state.
- Publishes change events (created, updated, deleted) to Redpanda.

### 2. Data Plane (Go)
*Location: `./data-plane/`*
- High-performance, low-latency read service.
- **Kafka Consumer**: Listens to the `nexus.signals` topic and projects changes into a **Redis Materialized View**.
- **HTTP Server**: Exposes a REST API resolving reads directly from Redis.
- **CLI Client**: Terminal interface to interact with the read API (`nexus-cli list`, `nexus-cli get`).

### 3. Infrastructure
*Location: `./infrastructure/`*
- Locally managed via Docker Compose.
- **PostgreSQL**: Used by the Control Plane.
- **Redpanda**: High-performance Kafka-compatible platform.
- **Redis**: Used by the Data Plane as the materialized view.
- **Redpanda Console**: Web UI for inspecting topics (accessible at `http://localhost:8080`).

## Getting Started

### 1. Start the Infrastructure
```bash
# Start PostgreSQL, Redpanda, and Redis
docker compose -f infrastructure/docker-compose.yml up -d
```

### 2. Run the Control Plane
```bash
cd control-plane
cp .env.example .env

# Full setup (install dependencies, migrate DB, create admin)
make setup

# Run the Django server
make run
```

### 3. Run the Data Plane
```bash
cd data-plane
go mod tidy

# Run the Go consumer and HTTP server
make run_server
```

In another terminal, you can interact with the CLI:
```bash
cd data-plane
make run_cli ARGS="list"
```

## Documentation
- [Control Plane Details](./control-plane/README.md)
- [Data Plane Details](./data-plane/README.md)
- [Infrastructure Details](./infrastructure/README.md)
