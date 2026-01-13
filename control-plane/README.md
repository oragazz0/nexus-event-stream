# Nexus Control Plane

This directory contains the backend configuration and core logic for the Nexus control plane, part of the Nexus Event Stream system.

## Purpose

Nexus serves as the authoritative source of truth and command center. It exposes a REST API for:

- **Signal Management**: Broadcasting prioritized directives (Low, Medium, High) to distributed agents.
- **Identity Management**: Handling Users and Groups authentication and authorization.

## Architecture

The project is built with **Django** and **Django Rest Framework (DRF)**.

### Core App (`nexus.core`)
The main application logic resides in `core/`.
- **Models**:
  - `Signal`: Represents a directive with a title, content, priority, and author.
- **API**: Exposes endpoints via DRF ViewSets.

## Development

### Prerequisites
- Python 3.14
- Docker (for PostgreSQL)

### Quick Start

```bash
cp .env.example .env
make setup
make run
```

### Makefile Shortcuts

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make db-up` | Start PostgreSQL container |
| `make db-down` | Stop PostgreSQL container |
| `make install` | Install Python dependencies |
| `make migrate` | Apply database migrations |
| `make superuser` | Create superuser from env vars (Remember to set the .env vars first) |
| `make run` | Run development server |
| `make setup` | Full setup (db, deps, migrate, superuser) |

### Manual Setup

If you prefer not to use Make:

```bash
# Start database
docker compose -f ../infrastructure/docker-compose.yml up -d

# Configure environment
cp .env.example .env

# Install dependencies
pip install -r requirements.txt

# Apply migrations
python manage.py migrate

# Create superuser (uses env vars from .env)
python manage.py createsuperuser --noinput

# Run the development server
python manage.py runserver
```

### API Endpoints
Base URL: `/api/v1/`

- `GET  /api/v1/signals/`: List all signals.
- `POST /api/v1/signals/`: Create a new signal.
- `GET  /api/v1/users/`: List users.
- `GET  /api/v1/groups/`: List groups.