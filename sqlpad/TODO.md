# SQLPad - Excel to SQLite Manager with SQL Query Interface

## Overview
SQLPad is a full-stack application that bridges Excel data with SQLite databases, providing a web-based SQL query interface. It allows users to upload Excel files (each sheet becomes a table in a SQLite database), browse database schemas, and execute ad-hoc SQL queries with a rich editor.

---

## Architecture

### Layered Structure (Clean Architecture / DDD)

```
/app/sqlpad/
├── cmd/server/            # Application entry point
├── internal/
│   ├── core/models/       # Domain entities & value objects
│   ├── infrastructure/    # External implementations
│   │   ├── database/      # SQLite operations
│   │   └── excel/         # Excel file parsing
│   └── server/            # Transport layer
│       ├── api/           # HTTP handlers
│       └── middleware/     # CORS, etc.
├── web/                   # Vue.js SPA frontend
│   ├── src/components/    # Vue components
│   └── public/            # Static assets
└── data/                  # SQLite database files storage
```

### Data Flow

```
Excel Upload:
  [Client] POST /api/upload (multipart/form-data)
    → [Server] Parse Excel file
    → [Infrastructure/Excel] Read sheets & rows
    → [Infrastructure/Database] Create/Upsert tables
    → [Server] Return success with table list

Query Execution:
  [Client] POST /api/databases/:db/query (JSON: { query: "..." })
    → [Server] Validate & sanitize query
    → [Infrastructure/Database] Execute SQL
    → [Server] Return JSON array results

Schema Discovery:
  [Client] GET /api/databases
    → [Server] List .db files in data/
    → [Client] Display in header dropdown

  [Client] GET /api/databases/:db/tables
    → [Server] Query sqlite_master
    → [Client] Display in left panel
```

---

## Backend (Go + Gin)

### Dependencies
- `github.com/gin-gonic/gin` — HTTP framework
- `modernc.org/sqlite` — Pure Go SQLite driver (no CGO)
- `github.com/xuri/excelize/v2` — Excel file reading
- `github.com/gin-contrib/cors` — CORS middleware

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/databases` | List all .db files in the data directory |
| GET | `/api/databases/:db/tables` | List all tables in the specified database |
| GET | `/api/databases/:db/tables/:table/schema` | Get column info for a table |
| POST | `/api/databases/:db/query` | Execute a SQL query and return results |
| POST | `/api/upload` | Upload Excel file, create/update database |
| GET | `/api/databases/:db/tables/:table/data` | Get paginated table data |
| GET | `/*any` | Serve static frontend files |

### Upload Flow Details

1. Client sends Excel file via multipart form
2. The endpoint uses the original filename (without extension) as the database name
3. Each Excel sheet is read — the sheet name becomes the table name
4. First row of each sheet is treated as column headers (normalized to `snake_case`)
5. Subsequent rows become data rows
6. If the table already exists, the schema is compared — new columns are added via `ALTER TABLE`
7. Data is upserted using rowid matching (if a primary key / unique constraint is detected)

### Query Execution Details

1. Client sends `{ "query": "SELECT ..." }` as JSON body
2. Server validates that the query is a read-only operation (SELECT, PRAGMA, EXPLAIN) — writes are allowed but logged
3. Query is executed against the specified database
4. Results are returned as `{ "columns": [...], "rows": [[...], ...], "affected": 0 }`
5. If an error occurs, `{ "error": "..." }` is returned

---

## Frontend (Vue.js + PicoCSS)

### Structure

```
web/
├── index.html              # Entry HTML with PicoCSS CDN
├── package.json            # Vite + Vue deps
├── vite.config.js          # Vite configuration
├── src/
│   ├── main.js             # Vue app bootstrap
│   ├── App.vue             # Root component (layout)
│   ├── api.js              # Axios/fetch API client
│   └── components/
│       ├── DatabaseHeader.vue   # DB selector in header
│       ├── TablePanel.vue       # Left sidebar table list
│       ├── SqlEditor.vue        # SQL editor with autocomplete
│       └── ResultsTable.vue     # Query results grid
```

### Component Responsibilities

#### App.vue
- Main layout: header (full width), body with left panel (20%) + right panel (80%)
- State management: current database, current table list, query results
- Loading states and error display

#### DatabaseHeader.vue
- Dropdown/select listing all `.db` files
- On change, emits event to load tables for selected database
- Shows current database name prominently

#### TablePanel.vue
- Lists all tables in the current database
- Clicking a table inserts `SELECT * FROM table_name LIMIT 100` into the editor
- Shows table schema on hover or click
- Scrollable list

#### SqlEditor.vue
- Textarea-based SQL editor with monospace font
- **Autocomplete**: Suggests SQL keywords (SELECT, FROM, WHERE, JOIN, etc.) and table names
- **Linter**: Basic validation — checks unmatched quotes, parentheses
- "Run" button (Ctrl+Enter shortcut)
- Dispatches query execution event
- Handles loading state

#### ResultsTable.vue
- Renders query results as an HTML table
- Handles empty results gracefully
- Displays column headers from query result
- Shows row count
- Handles errors with clear messaging

### SQL Autocomplete Implementation

Simple prefix-based autocomplete:
- Maintain a list of SQLite keywords: `SELECT, FROM, WHERE, INSERT, UPDATE, DELETE, CREATE, ALTER, DROP, JOIN, LEFT, RIGHT, INNER, OUTER, ON, AND, OR, NOT, IN, EXISTS, BETWEEN, LIKE, ORDER, BY, GROUP, HAVING, LIMIT, OFFSET, AS, DISTINCT, COUNT, SUM, AVG, MIN, MAX, UNION, ALL, NULL, TRUE, FALSE, CASE, WHEN, THEN, ELSE, END, SET, VALUES`
- Fetch table names from API when database is selected
- When user types, show dropdown suggestions based on current word prefix
- On selection, replace current word with suggestion

### Styling with PicoCSS

- Use PicoCSS classes for clean, responsive design
- Custom CSS for the split-pane layout (left panel 20%, right panel 80%)
- Dark theme preferred for SQL editor
- Monospace font for SQL editor and results

---

## Implementation Steps

### Phase 1: Backend Core
1. Initialize Go module with dependencies
2. Create domain models (Database, Table, Query)
3. Implement SQLite infrastructure layer
4. Implement Excel parsing infrastructure
5. Create API handlers
6. Set up Gin router and middleware
7. Create main entry point

### Phase 2: Frontend
1. Initialize Vite + Vue 3 project
2. Create PicoCSS-styled layout
3. Implement API client module
4. Build DatabaseHeader component
5. Build TablePanel component
6. Build SqlEditor with autocomplete
7. Build ResultsTable component
8. Wire everything in App.vue

### Phase 3: Integration & Polish
1. Configure Go to serve frontend static files
2. Test end-to-end: upload Excel → query
3. Error handling and edge cases
4. SQL autocomplete refinement
5. Documentation

---

## Database Schema (Internal)

The application uses SQLite databases stored in `/app/sqlpad/data/`. Each `.db` file represents one "project" derived from an Excel upload.

### Excel-to-Table Mapping
- Excel file: `sales_report.xlsx` → Database: `sales_report.db`
- Sheet "Q1_2024" → Table `q1_2024`
- Sheet "Q2_2024" → Table `q2_2024`

### Column Type Inference
- String values → `TEXT`
- Numeric values (int) → `INTEGER`
- Numeric values (float) → `REAL`
- Boolean-like values → `INTEGER` (0/1)
- Empty values → `TEXT` (NULL allowed)

---

## Error Handling Strategy

### Backend
- All API responses use consistent JSON structure
- Success: `{ "data": ..., "message": "ok" }`
- Error: `{ "error": "description", "details": "..." }`
- HTTP status codes follow REST conventions
- Query errors return 400 with SQLite error message

### Frontend
- Loading states for all async operations
- Error display using PicoCSS `[role="alert"]` components
- Empty states with helpful messages
- Network errors caught and displayed gracefully

---

## Security Considerations

- No authentication required (as specified)
- SQL queries are executed directly — user is responsible
- Read-only queries are encouraged but not enforced (user can write)
- Uploaded files are validated for Excel format
- Maximum file size limit: 50MB
- Database files are isolated per project

---

## Future Enhancements (Out of Scope)

- Query history
- Saved queries / scripts
- Export results to CSV/JSON
- Chart/visualization support
- Multi-user with authentication
- Database import (CSV, JSON, Parquet)
- Dark/light theme toggle
