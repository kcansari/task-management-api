package utils

import (
	"fmt"
	"time"

	// JWT library for creating and validating tokens
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the data we store inside the JWT token
// This struct will be embedded in the token and can be extracted later
// jwt.RegisteredClaims provides standard JWT fields like expiration
type Claims struct {
	UserID uint   `json:"user_id"` // Custom field: which user this token belongs to
	Email  string `json:"email"`   // Custom field: user's email for convenience
	// Embedding jwt.RegisteredClaims gives us standard fields like exp, iat, etc.
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
// It takes userID, email, and secret key as parameters
// Returns the token string and any error that occurred
func GenerateToken(userID uint, email, secretKey string) (string, error) {
	// Create the claims (payload) for our token
	// This is the data that will be stored inside the JWT
	claims := Claims{
		UserID: userID,
		Email:  email,
		// RegisteredClaims contains standard JWT fields
		RegisteredClaims: jwt.RegisteredClaims{
			// Token expires in 24 hours from now
			// time.Now().Add() is Go's way to add duration to current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			// IssuedAt is when the token was created (now)
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// Issuer identifies who created the token (our app)
			Issuer: "task-management-api",
		},
	}

	// Create a new token with our claims
	// jwt.SigningMethodHS256 is HMAC-SHA256, a symmetric signing algorithm
	// This means the same secret key is used for both signing and verification
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key
	// []byte(secretKey) converts string to byte slice (required by the library)
	// This creates the final JWT string that can be sent to clients
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		// If signing fails, return empty string and the error
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	// Return the token string and nil error (success)
	return tokenString, nil
}

// ValidateToken takes a JWT token string and validates it
// Returns the claims if valid, or an error if invalid/expired
func ValidateToken(tokenString, secretKey string) (*Claims, error) {
	// Parse the token string and validate it
	// jwt.ParseWithClaims needs:
	// 1. The token string
	// 2. A struct to parse claims into (empty Claims struct)
	// 3. A function that returns the key for validation
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is what we expect (HMAC-SHA256)
		// This prevents attacks where someone changes the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return our secret key as bytes for validation
		return []byte(secretKey), nil
	})

	// Check if parsing failed
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract the claims from the parsed token
	// Type assertion: claims.(type) checks if claims is of type *Claims
	// The second return value (ok) tells us if the assertion succeeded
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if the token is valid (not expired, properly signed, etc.)
	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	// Return the claims - caller can access UserID, Email, etc.
	return claims, nil
}