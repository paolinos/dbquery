package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dbquery/dbquery/internal/infrastructure/database"
)

const jwtExpiration = 2 * time.Hour

// LoginHandler authenticates a user and returns a JWT token.
func LoginHandler(authDB *database.AuthDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request",
				"details": "Username and password are required",
			})
			return
		}

		user, err := authDB.AuthenticateUser(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"details": err.Error(),
			})
			return
		}

		// Generate JWT token
		token, err := generateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate token",
			})
			return
		}

		expiresAt := time.Now().Add(jwtExpiration)
		if err := authDB.StoreJWT(user.ID, token, expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to store token",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"token":    token,
				"username": user.Username,
			},
			"message": "ok",
		})
	}
}

// RegisterHandler creates the first user. Only works when no users exist.
func RegisterHandler(authDB *database.AuthDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if any users already exist
		hasUsers, err := authDB.HasUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check users",
			})
			return
		}

		if hasUsers {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Registration disabled",
				"details": "A user already exists. Please log in.",
			})
			return
		}

		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request",
				"details": "Username and password are required",
			})
			return
		}

		if len(req.Username) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid username",
				"details": "Username must be at least 3 characters",
			})
			return
		}

		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid password",
				"details": "Password must be at least 6 characters",
			})
			return
		}

		user, err := authDB.CreateUser(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Failed to create user",
				"details": err.Error(),
			})
			return
		}

		// Auto-login: generate token
		token, err := generateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate token",
			})
			return
		}

		expiresAt := time.Now().Add(jwtExpiration)
		if err := authDB.StoreJWT(user.ID, token, expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to store token",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"data": gin.H{
				"token":    token,
				"username": user.Username,
			},
			"message": "User created successfully",
		})
	}
}

// LogoutHandler revokes the current JWT token.
func LogoutHandler(authDB *database.AuthDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, exists := c.Get("token")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No token found in request",
			})
			return
		}

		if err := authDB.RevokeToken(token.(string)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to revoke token",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out successfully",
		})
	}
}

// MeHandler returns the current authenticated user's info.
func MeHandler(authDB *database.AuthDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Not authenticated",
			})
			return
		}

		user, err := authDB.GetUserByID(userID.(int64))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"user_id":  user.ID,
				"username": user.Username,
			},
			"message": "ok",
		})
	}
}

// HasUsersHandler returns whether any users exist (for first-time setup detection).
func HasUsersHandler(authDB *database.AuthDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		hasUsers, err := authDB.HasUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check users",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"has_users": hasUsers,
			},
			"message": "ok",
		})
	}
}

// generateToken creates a cryptographically secure random token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
