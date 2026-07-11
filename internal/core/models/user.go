package models

import "time"

// User represents an authenticated user.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // never serialized
	CreatedAt    time.Time `json:"created_at"`
}

// JWTToken represents a stored JWT token.
type JWTToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

// LoginRequest is the body for a login attempt.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest is the body for first-time registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is returned after successful login or registration.
type AuthResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

// MeResponse is returned by the /me endpoint.
type MeResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}
