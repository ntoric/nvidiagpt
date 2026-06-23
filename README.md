# NvidiaGPT — ChatGPT Clone

A ChatGPT-style chat application powered by the NVIDIA API (Llama 4 Maverick).

## Architecture

- **Frontend:** React + Vite + TailwindCSS (runs locally, not dockerized)
- **Backend:** Go (HTTP server, SSE streaming)
- **Database:** PostgreSQL (conversations & messages)
- **Cache:** Redis
- **LLM:** NVIDIA API (`meta/llama-4-maverick-17b-128e-instruct`)

## Project Structure

```
NvidiaGPT/
├── docker-compose.yml      # PostgreSQL + Redis + Go backend
├── backend/
│   ├── Dockerfile
│   ├── .env.example
│   ├── go.mod / go.sum
│   ├── main.go             # Entry point, routes, CORS
│   ├── config.go           # Env config loading
│   ├── db/                 # PostgreSQL connection & migrations
│   ├── cache/              # Redis connection
│   ├── models/             # DB queries (conversations, messages)
│   ├── nvidia/             # NVIDIA API streaming client
│   └── handlers/           # HTTP handlers (CRUD + chat SSE)
└── frontend/
    ├── package.json
    ├── vite.config.js       # Dev proxy to backend
    └── src/
        ├── main.jsx
        ├── App.jsx          # Main chat UI
        ├── api.js           # API client (fetch + SSE streaming)
        └── index.css        # Tailwind + custom styles
```

## Quick Start

### 1. Set up the NVIDIA API key

```bash
cp backend/.env.example backend/.env
# Edit backend/.env and set your NVIDIA_API_KEY
```

### 2. Start backend services (PostgreSQL, Redis, Go backend)

```bash
docker compose up --build
```

This starts:
- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`
- Go backend on `localhost:8089`

### 3. Start the frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend runs on `http://localhost:5173` and proxies `/api` requests to the Go backend.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/conversations` | List all conversations |
| POST | `/api/conversations` | Create a new conversation |
| GET | `/api/conversations/:id` | Get conversation with messages |
| DELETE | `/api/conversations/:id` | Delete a conversation |
| POST | `/api/conversations/:id/chat` | Send message & stream response (SSE) |

## How Streaming Works

1. Frontend sends a POST to `/api/conversations/:id/chat` with `{"message": "..."}`
2. Backend saves the user message to PostgreSQL
3. Backend calls NVIDIA API with `stream: true`
4. Backend reads the SSE stream from NVIDIA and forwards each token to the frontend via SSE
5. Frontend reads the stream with `ReadableStream` and renders tokens in real-time
6. When the stream completes, the full assistant response is saved to PostgreSQL

## Configuration

### Backend (.env)

| Variable | Default | Description |
|----------|---------|-------------|
| `NVIDIA_API_KEY` | (required) | Your NVIDIA API key |
| `NVIDIA_MODEL` | `meta/llama-4-maverick-17b-128e-instruct` | Model to use |
| `PORT` | `8089` | Backend server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `nvidiagpt` | PostgreSQL user |
| `DB_PASSWORD` | `nvidiagpt` | PostgreSQL password |
| `DB_NAME` | `nvidiagpt` | PostgreSQL database name |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |

> When running with docker compose, DB and Redis hosts are set automatically to `postgres` and `redis` respectively.
