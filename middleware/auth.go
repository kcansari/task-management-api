package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kcansari/task-management-api/config"
	"github.com/kcansari/task-management-api/utils"
)

// ContextKey is a custom type for context keys to avoid collisions
// In Go, context keys should be their own type to prevent key conflicts
// This is a best practice when storing values in request context
type ContextKey string

// UserContextKey is the key used to store user information in request context
// This allows us to access user data in protected route handlers
const UserContextKey ContextKey = "user"

// UserContext represents the user data we store in request context
// This is what protected handlers will have access to
type UserContext struct {
	UserID uint   `json:"user_id"` // ID of the authenticated user
	Email  string `json:"email"`   // Email of the authenticated user
}

// ErrorResponse represents an error message for middleware responses
type ErrorResponse struct {
	Error string `json:"error"` // Human-readable error message
}

// AuthMiddleware is a higher-order function that returns HTTP middleware
// Middleware in Go is a function that wraps another HTTP handler
// This pattern allows us to add authentication to any route by wrapping it
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	// Return a new handler function that includes authentication logic
	// This is a closure - it "closes over" the 'next' parameter
	return func(w http.ResponseWriter, r *http.Request) {
		// Set JSON content type for all middleware responses
		w.Header().Set("Content-Type", "application/json")

		// Extract the Authorization header from the request
		// HTTP Authorization header format: "Bearer <token>"
		authHeader := r.Header.Get("Authorization")
		
		// Check if Authorization header is present
		if authHeader == "" {
			// No authorization header provided
			w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Authorization header required"})
			return // Stop processing, don't call next handler
		}

		// Parse the Authorization header
		// Expected format: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
		// strings.SplitN splits into at most N parts (here, 2 parts)
		parts := strings.SplitN(authHeader, " ", 2)
		
		// Validate Authorization header format
		if len(parts) != 2 {
			// Header doesn't have exactly 2 parts (scheme and token)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid authorization header format"})
			return
		}

		// Extract scheme and token
		scheme := parts[0]  // Should be "Bearer"
		token := parts[1]   // The actual JWT token

		// Verify the authentication scheme is Bearer
		// Bearer token is the standard for JWT authentication
		if scheme != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid authorization scheme. Use Bearer"})
			return
		}

		// Validate the JWT token using our utility function
		// Load configuration to get the JWT secret key
		cfg := config.Load()
		claims, err := utils.ValidateToken(token, cfg.JWTSecret)
		if err != nil {
			// Token validation failed (expired, invalid signature, malformed, etc.)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid or expired token"})
			return
		}

		// Token is valid! Create user context from the claims
		userCtx := UserContext{
			UserID: claims.UserID,
			Email:  claims.Email,
		}

		// Add user information to the request context
		// context.WithValue creates a new context with the user data
		// This allows the next handler to access the authenticated user's info
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		
		// Create a new request with the updated context
		// In Go, context is immutable, so we need to create a new request
		r = r.WithContext(ctx)

		// Authentication successful! Call the next handler in the chain
		// This is where the actual route handler (like GetTasks) will execute
		next(w, r)
	}
}

// GetUserFromContext extracts user information from request context
// This is a helper function that protected handlers can use
// It returns the user context and a boolean indicating if user was found
func GetUserFromContext(r *http.Request) (UserContext, bool) {
	// Extract the user value from context using our key
	// r.Context().Value() returns interface{}, so we need type assertion
	user, ok := r.Context().Value(UserContextKey).(UserContext)
	
	// Return the user context and whether the extraction was successful
	// If ok is false, it means no user was found in context (not authenticated)
	return user, ok
}