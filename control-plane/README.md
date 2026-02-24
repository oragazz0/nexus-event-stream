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
Base URL: `http://localhost:8000/api/v1/`

- `GET  /api/v1/signals/`: List all signals.
- `POST /api/v1/signals/`: Create a new signal.
- `PUT/PATCH /api/v1/signals/{id}/`: Update a signal.
- `DELETE /api/v1/signals/{id}/`: Delete a signal.
- `GET  /api/v1/users/`: List users.
- `GET  /api/v1/groups/`: List groups.

#### Interacting with the API

You can interact with the API either directly through the Django Rest Framework (DRF) web browsable API or via `curl`.

**1. Using the DRF Web Panel:**
Open your browser and navigate to [http://localhost:8000/api/v1/signals/](http://localhost:8000/api/v1/signals/).
- Here you can view the list of signals formatted beautifully.
- Scroll to the bottom to find an HTML form where you can **create** a new signal. (You might need to log in first using the superuser credentials set during `make setup`).
- Click on an individual signal's url (e.g., `http://localhost:8000/api/v1/signals/<uuid>/`) to view its specific details. On the detail page, you will find forms and buttons to **update** (PUT/PATCH) or **delete** it.

**2. Using cURL:**

> [!NOTE]
> The examples below use Basic Auth (`-u admin:admin`). If you configured different superuser credentials via `.env`, please replace `admin:admin` with your actual username and password.

**List Signals:**
```bash
curl -u admin:admin http://localhost:8000/api/v1/signals/
```

**Create a Signal:**
```bash
curl -X POST http://localhost:8000/api/v1/signals/ \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "title": "System Reboot Scheduled",
    "content": "A reboot will happen at midnight.",
    "priority": "High"
  }'
```
*(Note: If `author` is required, you may also need to provide `"author": 1` in the JSON payload).*

**Update a Signal (e.g., PATCH to change priority):**
```bash
curl -X PATCH http://localhost:8000/api/v1/signals/<uuid>/ \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{"priority": "Low"}'
```

**Delete a Signal:**
```bash
curl -X DELETE http://localhost:8000/api/v1/signals/<uuid>/ \
  -u admin:admin
```