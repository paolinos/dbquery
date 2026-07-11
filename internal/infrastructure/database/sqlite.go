package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/dbquery/dbquery/internal/core/models"
)

// SQLiteDB wraps a SQLite database connection.
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB opens (or creates) a SQLite database at the given path.
func NewSQLiteDB(path string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s: %w", path, err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &SQLiteDB{db: db}, nil
}

// Close closes the database connection.
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// GetDB returns the underlying *sql.DB.
func (s *SQLiteDB) GetDB() *sql.DB {
	return s.db
}

// ListDatabases returns all .db files in the given directory that belong to the specified user.
// Files are prefixed with "{userID}_" to isolate per-user data.
// The _auth.db file is always excluded.
func ListDatabases(dataDir string, userID int64) ([]models.DatabaseInfo, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.DatabaseInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read data directory %s: %w", dataDir, err)
	}

	// Build the user prefix: "1_" for userID 1
	userPrefix := fmt.Sprintf("%d_", userID)

	var databases = []models.DatabaseInfo{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".db") {
			continue
		}
		// Always exclude the auth database
		if name == "_auth.db" {
			continue
		}
		// Only include databases belonging to this user
		if !strings.HasPrefix(name, userPrefix) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		// Strip the user prefix from the display name
		dbName := strings.TrimPrefix(name, userPrefix)
		dbName = strings.TrimSuffix(dbName, ".db")
		databases = append(databases, models.DatabaseInfo{
			Name: dbName,
			Path: filepath.Join(dataDir, name),
			Size: info.Size(),
		})
	}

	return databases, nil
}

// ListDatabasesRaw returns all .db files in the given directory (no user filtering).
// Used internally for operations that need the real filename on disk.
func ListDatabasesRaw(dataDir string) ([]models.DatabaseInfo, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.DatabaseInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read data directory %s: %w", dataDir, err)
	}

	var databases = []models.DatabaseInfo{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".db") && name != "_auth.db" {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			databases = append(databases, models.DatabaseInfo{
				Name: strings.TrimSuffix(name, ".db"),
				Path: filepath.Join(dataDir, name),
				Size: info.Size(),
			})
		}
	}

	return databases, nil
}

// GetUserDBPath returns the file path for a user-scoped database.
// The filename is prefixed with "{userID}_" to isolate per-user data.
func GetUserDBPath(dataDir string, userID int64, dbName string) string {
	dbName = filepath.Clean(dbName)
	dbName = strings.ReplaceAll(dbName, "..", "")
	dbName = strings.ReplaceAll(dbName, "/", "")
	dbName = strings.ReplaceAll(dbName, "\\", "")
	prefixed := fmt.Sprintf("%d_%s", userID, dbName)
	return filepath.Join(dataDir, prefixed+".db")
}

// ResolveUserDBPath resolves a user-facing database name to the actual file path on disk.
// It searches for "{userID}_{dbName}.db" in the data directory.
func ResolveUserDBPath(dataDir string, userID int64, dbName string) (string, error) {
	prefixed := fmt.Sprintf("%d_%s", userID, dbName)
	path := filepath.Join(dataDir, prefixed+".db")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("database not found: %s", dbName)
	}
	return path, nil
}

// GetTables returns all tables in the database.
func (s *SQLiteDB) GetTables() ([]models.TableInfo, error) {
	query := `SELECT name, sql FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables = []models.TableInfo{}
	for rows.Next() {
		var t models.TableInfo
		var schema sql.NullString
		if err := rows.Scan(&t.Name, &schema); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}
		if schema.Valid {
			t.Schema = schema.String
		}

		// Get columns
		columns, err := s.GetTableColumns(t.Name)
		if err != nil {
			// Non-fatal: continue with empty columns
			columns = []models.ColumnInfo{}
		}
		t.Columns = columns

		// Get row count estimate
		var count sql.NullInt64
		err = s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM [%s]", t.Name)).Scan(&count)
		if err == nil && count.Valid {
			t.RowCount = count.Int64
		}

		tables = append(tables, t)
	}

	if tables == nil {
		tables = []models.TableInfo{}
	}

	return tables, rows.Err()
}

// GetTableColumns returns column information for a given table.
func (s *SQLiteDB) GetTableColumns(tableName string) ([]models.ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info([%s])", tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table info for %s: %w", tableName, err)
	}
	defer rows.Close()

	var columns []models.ColumnInfo
	for rows.Next() {
		var c models.ColumnInfo
		var notNull, pk int
		var defaultVal sql.NullString

		if err := rows.Scan(&c.CID, &c.Name, &c.Type, &notNull, &defaultVal, &pk); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		c.NotNull = notNull == 1
		c.PrimaryKey = pk == 1
		if defaultVal.Valid {
			c.DefaultValue = defaultVal.String
		}
		columns = append(columns, c)
	}

	if columns == nil {
		columns = []models.ColumnInfo{}
	}

	return columns, rows.Err()
}

// ExecuteQuery executes a SQL query and returns the result.
// For SELECT-like queries, it returns the rows data.
// For INSERT/UPDATE/DELETE, it returns the affected row count.
func (s *SQLiteDB) ExecuteQuery(query string) (*models.QueryResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	upper := strings.ToUpper(query)
	isReadOnly := strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "PRAGMA") ||
		strings.HasPrefix(upper, "EXPLAIN") ||
		strings.HasPrefix(upper, "WITH")

	if isReadOnly {
		return s.executeReadQuery(query)
	}

	return s.executeWriteQuery(query)
}

func (s *SQLiteDB) executeReadQuery(query string) (*models.QueryResult, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	result := &models.QueryResult{
		Columns: columns,
		Rows:    make([][]interface{}, 0),
	}

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert byte slices to strings for JSON serialization
		row := make([]interface{}, len(columns))
		for i, v := range values {
			switch val := v.(type) {
			case []byte:
				row[i] = string(val)
			default:
				row[i] = val
			}
		}
		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	result.Message = fmt.Sprintf("%d row(s) returned", len(result.Rows))
	return result, nil
}

func (s *SQLiteDB) executeWriteQuery(query string) (*models.QueryResult, error) {
	result, err := s.db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	affected, _ := result.RowsAffected()

	return &models.QueryResult{
		Columns:  []string{},
		Rows:     [][]interface{}{},
		Affected: affected,
		Message:  fmt.Sprintf("%d row(s) affected", affected),
	}, nil
}

// CreateTable creates a table from Excel sheet data.
func (s *SQLiteDB) CreateTable(tableName string, headers []string, rows [][]string) error {
	// Build column definitions from headers and sample data
	colDefs := make([]string, len(headers))
	for i, h := range headers {
		colType := inferColumnType(rows, i)
		colName := normalizeColumnName(h)
		colDefs[i] = fmt.Sprintf("[%s] %s", colName, colType)
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS [%s] (\n  %s\n)", tableName, strings.Join(colDefs, ",\n  "))
	if _, err := s.db.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	// Check if table already had data - if so, we might need to add columns
	existingCols, _ := s.GetTableColumns(tableName)
	existingNames := make(map[string]bool)
	for _, c := range existingCols {
		existingNames[strings.ToLower(c.Name)] = true
	}

	// Add any missing columns
	for _, h := range headers {
		colName := normalizeColumnName(h)
		if !existingNames[strings.ToLower(colName)] {
			colType := "TEXT" // Default type for new columns
			alterSQL := fmt.Sprintf("ALTER TABLE [%s] ADD COLUMN [%s] %s", tableName, colName, colType)
			if _, err := s.db.Exec(alterSQL); err != nil {
				return fmt.Errorf("failed to add column %s to table %s: %w", colName, tableName, err)
			}
		}
	}

	return nil
}

// InsertRows inserts data rows into a table. Uses a transaction for performance.
func (s *SQLiteDB) InsertRows(tableName string, headers []string, rows [][]string) error {
	if len(rows) == 0 {
		return nil
	}

	colNames := make([]string, len(headers))
	placeholders := make([]string, len(headers))
	for i, h := range headers {
		colNames[i] = fmt.Sprintf("[%s]", normalizeColumnName(h))
		placeholders[i] = "?"
	}

	insertSQL := fmt.Sprintf(
		"INSERT OR REPLACE INTO [%s] (%s) VALUES (%s)",
		tableName,
		strings.Join(colNames, ", "),
		strings.Join(placeholders, ", "),
	)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		values := make([]interface{}, len(headers))
		for i, v := range row {
			values[i] = v
			if i >= len(headers) {
				break
			}
		}
		// Pad with empty strings if row is shorter than headers
		for i := len(row); i < len(headers); i++ {
			values[i] = ""
		}
		if _, err := stmt.Exec(values...); err != nil {
			return fmt.Errorf("failed to insert row: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// inferColumnType tries to determine the SQLite column type from sample data.
func inferColumnType(rows [][]string, colIndex int) string {
	hasFloat := false
	hasInt := false
	hasText := false

	for _, row := range rows {
		if colIndex >= len(row) {
			continue
		}
		val := strings.TrimSpace(row[colIndex])
		if val == "" {
			continue
		}
		// Try integer
		if isInteger(val) {
			hasInt = true
			continue
		}
		// Try float
		if isFloat(val) {
			hasFloat = true
			continue
		}
		// Otherwise it's text
		hasText = true
	}

	if hasText {
		return "TEXT"
	}
	if hasFloat {
		return "REAL"
	}
	if hasInt {
		return "INTEGER"
	}
	return "TEXT"
}

func isInteger(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

func isFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// normalizeColumnName converts a header to a safe SQL column name.
func normalizeColumnName(name string) string {
	name = strings.TrimSpace(name)
	// Replace spaces and special chars with underscores
	var result strings.Builder
	for i, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '/' || r == '\\' || r == '.' || r == ',' || r == '(' || r == ')' {
			if i > 0 && result.String()[result.Len()-1] != '_' {
				result.WriteRune('_')
			}
		} else if r == '_' {
			result.WriteRune('_')
		}
	}
	colName := result.String()
	// Remove leading underscores
	colName = strings.TrimLeft(colName, "_")
	// Remove trailing underscores
	colName = strings.TrimRight(colName, "_")
	if colName == "" {
		colName = "column"
	}
	// Ensure it doesn't start with a digit
	if len(colName) > 0 && colName[0] >= '0' && colName[0] <= '9' {
		colName = "c_" + colName
	}
	return strings.ToLower(colName)
}

// GetDBPath returns the path for a database by name in the data directory.
func GetDBPath(dataDir, dbName string) string {
	// Sanitize db name
	dbName = filepath.Clean(dbName)
	dbName = strings.ReplaceAll(dbName, "..", "")
	dbName = strings.ReplaceAll(dbName, "/", "")
	dbName = strings.ReplaceAll(dbName, "\\", "")
	return filepath.Join(dataDir, dbName+".db")
}

// TableExists checks if a table exists in the database.
func (s *SQLiteDB) TableExists(tableName string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
		tableName,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// OpenExisting opens an existing SQLite database, returning an error if it doesn't exist.
func OpenExisting(path string) (*SQLiteDB, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file not found: %s", path)
	}
	return NewSQLiteDB(path)
}


