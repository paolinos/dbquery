# SQLPad — Excel to SQLite Manager

Import Excel spreadsheets into SQLite databases and run ad‑hoc SQL queries from a clean web interface. No accounts, no setup — just upload and query.

## High‑Level Architecture

```
┌──────────────────────────────────────────────┐
│              Browser (Vue.js + PicoCSS)       │
│  ┌─────────────┐  ┌────────────────────────┐ │
│  │ DB Selector  │  │   SQL Editor w/        │ │
│  │ (Header)     │  │   Autocomplete+Linter  │ │
│  ├─────────────┤  │                        │ │
│  │ Table Panel  │  │   Run Button           │ │
│  │ (Left 20%)   │  │   (Ctrl+Enter)         │ │
│  │              │  ├────────────────────────┤ │
│  │ Click →      │  │   Results Table        │ │
│  │ Insert SELECT│  │   (Right 80%)          │ │
│  └─────────────┘  └────────────────────────┘ │
└──────────────────────┬───────────────────────┘
                       │  HTTP / JSON
                       ▼
┌──────────────────────────────────────────────┐
│          Go + Gin (REST API)                  │
│                                              │
│  /api/databases         → List .db files     │
│  /api/databases/:db/    → Tables, queries,   │
│    tables, query, schema   autocomplete      │
│  /api/upload            → Import Excel       │
└──────────────────────┬───────────────────────┘
                       │
              ┌────────┴────────┐
              ▼                 ▼
       ┌────────────┐   ┌──────────────┐
       │  SQLite DB  │   │  Excel Files │
       │  (.db file) │   │  (.xlsx)     │
       └─────────────┘   └──────────────┘
```

### Layers

| Layer | Directory | Responsibility |
|-------|-----------|----------------|
| **Domain** | `internal/core/models/` | Entities (Table, Column, QueryResult, etc.) |
| **Infrastructure** | `internal/infrastructure/` | SQLite driver, Excel parser |
| **Server** | `internal/server/` | HTTP handlers, middleware, routing |
| **Entrypoint** | `cmd/server/` | Flag parsing, server bootstrap |
| **Frontend** | `web/` | Vue 3 SPA, PicoCSS, Vite build |

## Quick Start

### Prerequisites

- **Go** ≥1.21
- **Node.js** ≥18 (for building the frontend)
- **Make** (optional, but recommended)

### 1. Install Dependencies

```bash
make install
```

### 2. Run in Development Mode

Starts the Vite dev server (`:5173`) and Go backend (`:8080`) concurrently with hot-reload:

```bash
make dev
```

Open **[http://localhost:5173](http://localhost:5173)** — Vite proxies `/api` requests to the backend.

### 3. Build for Production

```bash
make build
```

This produces a self-contained `build/` directory:

```
build/
  sqlpad-server    # Go binary
  web/             # Frontend SPA (index.html, assets/)
```

Run the production build:

```bash
cd build && ./sqlpad-server
```

Open **[http://localhost:8080](http://localhost:8080)**

## Usage

### Upload an Excel File

1. Click **📤 Upload Excel** in the header.
2. Select an `.xlsx` file. Each sheet becomes a SQLite table. Column types (INTEGER, REAL, TEXT) are inferred automatically.
3. The database is named after the file (e.g., `sales_report.xlsx` → `sales_report.db`).

### Browse Tables

- After selecting a database, the left panel shows all tables with row counts.
- **Click** a table to expand its schema (column names and types).
- **Double‑click** a table to insert `SELECT * FROM [table] LIMIT 100` into the editor.

### Run SQL Queries

1. Type (or paste) your SQL in the editor.
2. Press **Ctrl+Enter** (or click **▶ Run**).
3. Results appear immediately in the table below.
4. The divider between editor and results is draggable.

### Autocomplete

- Start typing any word — a dropdown suggests SQL keywords, table names, and column names filtered by prefix.
- Navigate with **↑/↓**, select with **Enter/Tab**, dismiss with **Esc**.

### Linter

Unmatched quotes and parentheses are flagged in real‑time below the editor.

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/databases` | List all databases (`.db` files in the data directory) |
| `GET` | `/api/databases/:db/tables` | List tables with schema and row counts |
| `GET` | `/api/databases/:db/tables/:table/schema` | Column definitions for a table |
| `GET` | `/api/databases/:db/tables/:table/data?page=1&per_page=100` | Paginated table data |
| `POST` | `/api/databases/:db/query` | Execute SQL `{ "query": "SELECT ..." }` |
| `GET` | `/api/databases/:db/autocomplete` | Table + column names for the editor |
| `POST` | `/api/upload` | Upload Excel file (multipart, field: `file`) |

### Example: Run a Query

```bash
curl -X POST http://localhost:8080/api/databases/my_data/query \
  -H "Content-Type: application/json" \
  -d '{"query": "SELECT name, email FROM users WHERE age > 25"}'
```

Response:

```json
{
  "data": {
    "columns": ["name", "email"],
    "rows": [
      ["Alice Johnson", "alice@example.com"],
      ["Charlie Brown", "charlie@example.com"]
    ],
    "affected": 0,
    "message": "2 row(s) returned"
  },
  "message": "ok"
}
```

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | HTTP server port |
| `--data` | `./data` | Directory for SQLite database files |
| `--frontend` | `./web` | Path to the built frontend directory |

## Project Layout

```
sqlpad/
├── cmd/server/              # Application entry point
├── internal/
│   ├── core/models/         # Domain entities
│   ├── infrastructure/      # SQLite + Excel implementations
│   └── server/              # HTTP transport layer
├── web/                     # Vue.js SPA
│   ├── src/components/      # Vue components
│   └── dist/                # Vite build output (intermediate)
├── build/                   # Production build output (gitignored)
│   ├── sqlpad-server        # Go binary
│   └── web/                 # Frontend SPA for serving
├── data/                    # SQLite database storage
├── Makefile                 # Build & dev automation
├── TODO.md                  # Full architecture document
└── README.md
```

## Makefile Reference

| Command | Description |
|---------|-------------|
| `make install` | Install all dependencies (FE + BE) |
| `make install-fe` | Install frontend dependencies only |
| `make install-be` | Install backend dependencies only |
| `make dev` | Run FE and BE concurrently (Ctrl+C kills both) |
| `make dev-fe` | Start Vite dev server only (`:5173`) |
| `make dev-be` | Start Go backend only (`:8080`, uses `air` if available) |
| `make build` | Build FE and BE into `build/` |
| `make build-fe` | Build frontend for production → `build/web/` |
| `make build-be` | Build backend binary → `build/sqlpad-server` |
| `make clean` | Remove all build artifacts |
| `make help` | Show available targets |

## Tech Stack

- **Backend**: Go, Gin, modernc.org/sqlite (pure Go, no CGO), excelize
- **Frontend**: Vue 3 (Composition API), Vite, PicoCSS
- **Database**: SQLite (one file per project)
- **No authentication**, no external database server required.
