package models

// DatabaseInfo represents a SQLite database file.
type DatabaseInfo struct {
	Name string `json:"name"` // Database name (filename without .db extension)
	Path string `json:"path"` // Full path to the .db file
	Size int64  `json:"size"` // File size in bytes
}

// TableInfo represents a table within a database.
type TableInfo struct {
	Name    string       `json:"name"`    // Table name
	Schema  string       `json:"schema"`  // CREATE TABLE statement
	Columns []ColumnInfo `json:"columns"` // Column details
	RowCount int64       `json:"row_count"` // Approximate row count
}

// ColumnInfo represents a column in a table.
type ColumnInfo struct {
	CID          int    `json:"cid"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	NotNull      bool   `json:"not_null"`
	DefaultValue string `json:"default_value"`
	PrimaryKey   bool   `json:"primary_key"`
}

// QueryResult holds the result of a SQL query execution.
type QueryResult struct {
	Columns  []string        `json:"columns"`  // Column names
	Rows     [][]interface{} `json:"rows"`     // Data rows
	Affected int64           `json:"affected"` // Number of rows affected (for INSERT/UPDATE/DELETE)
	Message  string          `json:"message,omitempty"` // Status message
}

// QueryRequest is the body for executing a SQL query.
type QueryRequest struct {
	Query string `json:"query" binding:"required"`
}

// UploadResponse is returned after an Excel upload.
type UploadResponse struct {
	Database string      `json:"database"` // Database name created/updated
	Tables   []TableInfo `json:"tables"`   // Tables created/updated
	Message  string      `json:"message"`
}

// ExcelSheet represents a parsed Excel sheet.
type ExcelSheet struct {
	Name    string     // Sheet name
	Headers []string   // Column headers from first row
	Rows    [][]string // Data rows (excluding header)
}

// TableDataRequest holds pagination parameters for table browsing.
type TableDataRequest struct {
	Page    int `form:"page" json:"page"`
	PerPage int `form:"per_page" json:"per_page"`
}
