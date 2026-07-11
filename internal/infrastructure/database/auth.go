package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dbquery/dbquery/internal/core/models"
)

// AuthDB wraps the authentication SQLite database.
type AuthDB struct {
	db *sql.DB
}

// InitAuthDB opens or creates the _auth.db file in the data directory.
// If the file does not exist, it is created along with the required schema.
func InitAuthDB(dataDir string) (*AuthDB, error) {
	dbPath := filepath.Join(dataDir, "_auth.db")

	// Open or create the database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open auth database: %w", err)
	}

	// Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set WAL mode on auth db: %w", err)
	}

	// Create tables if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS jwt_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		revoked BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create auth schema: %w", err)
	}

	return &AuthDB{db: db}, nil
}

// Close closes the auth database connection.
func (a *AuthDB) Close() error {
	return a.db.Close()
}

// HasUsers returns true if at least one user exists.
func (a *AuthDB) HasUsers() (bool, error) {
	var count int
	if err := a.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check users: %w", err)
	}
	return count > 0, nil
}

// CreateUser creates a new user with a bcrypt-hashed password.
func (a *AuthDB) CreateUser(username, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	result, err := a.db.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		username, string(hash),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, _ := result.LastInsertId()

	return &models.User{
		ID:           id,
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}, nil
}

// AuthenticateUser verifies credentials and returns the user if valid.
func (a *AuthDB) AuthenticateUser(username, password string) (*models.User, error) {
	var user models.User
	err := a.db.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid username or password")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	return &user, nil
}

// StoreJWT saves a JWT token in the database.
func (a *AuthDB) StoreJWT(userID int64, token string, expiresAt time.Time) error {
	_, err := a.db.Exec(
		"INSERT INTO jwt_tokens (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store JWT: %w", err)
	}
	return nil
}

// IsTokenValid checks if a token is valid (exists, not revoked, not expired).
func (a *AuthDB) IsTokenValid(token string) (bool, int64, error) {
	var userID int64
	var revoked bool
	var expiresAt time.Time

	err := a.db.QueryRow(
		"SELECT user_id, revoked, expires_at FROM jwt_tokens WHERE token = ?",
		token,
	).Scan(&userID, &revoked, &expiresAt)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("failed to validate token: %w", err)
	}

	if revoked || time.Now().After(expiresAt) {
		return false, 0, nil
	}

	return true, userID, nil
}

// RevokeToken marks a token as revoked.
func (a *AuthDB) RevokeToken(token string) error {
	_, err := a.db.Exec(
		"UPDATE jwt_tokens SET revoked = TRUE WHERE token = ?",
		token,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

// GetUserByID returns a user by their ID.
func (a *AuthDB) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	err := a.db.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// EnsureAuthDB checks if the auth DB file exists, creates it if not, and returns the handle.
func EnsureAuthDB(dataDir string) (*AuthDB, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return InitAuthDB(dataDir)
}
