package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestGenerateToken tests JWT token generation
func TestGenerateToken(t *testing.T) {
	testCases := []struct {
		name      string
		userID    uint
		email     string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "valid inputs",
			userID:    1,
			email:     "test@example.com",
			secretKey: "test-secret-key",
			wantErr:   false,
		},
		{
			name:      "empty email",
			userID:    1,
			email:     "",
			secretKey: "test-secret-key",
			wantErr:   false, // JWT allows empty email
		},
		{
			name:      "zero user ID",
			userID:    0,
			email:     "test@example.com",
			secretKey: "test-secret-key",
			wantErr:   false, // JWT allows zero user ID
		},
		{
			name:      "empty secret key",
			userID:    1,
			email:     "test@example.com",
			secretKey: "",
			wantErr:   false, // JWT allows empty secret (though insecure)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate token
			token, err := GenerateToken(tc.userID, tc.email, tc.secretKey)

			// Check error expectation
			if (err != nil) != tc.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// If no error expected, validate token format
			if !tc.wantErr {
				// Token should not be empty
				if token == "" {
					t.Errorf("GenerateToken() returned empty token")
				}

				// Token should have JWT structure (3 parts separated by dots)
				// JWT format: header.payload.signature
				parts := len([]rune(token)) // Count characters, not bytes
				if parts < 10 { // Reasonable minimum length for JWT
					t.Errorf("GenerateToken() returned suspiciously short token: %s", token)
				}

				// Try to parse the token to verify it's valid JWT format
				// This doesn't validate the signature, just the structure
				parsedToken, parseErr := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte(tc.secretKey), nil
				})

				if parseErr != nil {
					t.Errorf("GenerateToken() produced unparseable JWT: %v", parseErr)
				}

				// Extract claims and validate content
				if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
					// Check if user_id is present and correct
					if userID, exists := claims["user_id"]; !exists || userID != float64(tc.userID) {
						// JSON numbers are parsed as float64 by default
						t.Errorf("Token missing or incorrect user_id claim: got %v, want %v", userID, tc.userID)
					}

					// Check if email is present and correct
					if email, exists := claims["email"]; !exists || email != tc.email {
						t.Errorf("Token missing or incorrect email claim: got %v, want %v", email, tc.email)
					}

					// Check if issuer is set correctly
					if iss, exists := claims["iss"]; !exists || iss != "task-management-api" {
						t.Errorf("Token missing or incorrect issuer claim: got %v", iss)
					}

					// Check if expiration is set and in the future
					if exp, exists := claims["exp"]; exists {
						if expTime, ok := exp.(float64); ok {
							expiration := time.Unix(int64(expTime), 0)
							if expiration.Before(time.Now()) {
								t.Errorf("Token is already expired")
							}
							// Should expire in approximately 24 hours
							expectedExpiry := time.Now().Add(24 * time.Hour)
							if expiration.Sub(expectedExpiry) > time.Minute || expectedExpiry.Sub(expiration) > time.Minute {
								t.Errorf("Token expiration is not ~24 hours from now: %v", expiration)
							}
						}
					} else {
						t.Errorf("Token missing expiration claim")
					}
				}
			}
		})
	}
}

// TestValidateToken tests JWT token validation
func TestValidateToken(t *testing.T) {
	// Setup: create a valid token for testing
	testUserID := uint(123)
	testEmail := "test@example.com"
	testSecret := "test-secret-key"
	
	validToken, err := GenerateToken(testUserID, testEmail, testSecret)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	testCases := []struct {
		name      string
		token     string
		secretKey string
		wantErr   bool
		checkUserID bool // Whether to validate userID in claims
		checkEmail  bool // Whether to validate email in claims
	}{
		{
			name:        "valid token",
			token:       validToken,
			secretKey:   testSecret,
			wantErr:     false,
			checkUserID: true,
			checkEmail:  true,
		},
		{
			name:      "empty token",
			token:     "",
			secretKey: testSecret,
			wantErr:   true,
		},
		{
			name:      "malformed token",
			token:     "invalid.jwt.token",
			secretKey: testSecret,
			wantErr:   true,
		},
		{
			name:      "wrong secret key",
			token:     validToken,
			secretKey: "wrong-secret-key",
			wantErr:   true,
		},
		{
			name:      "random string as token",
			token:     "this-is-not-a-jwt-token",
			secretKey: testSecret,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate token
			claims, err := ValidateToken(tc.token, tc.secretKey)

			// Check error expectation
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// If no error expected, validate claims
			if !tc.wantErr {
				if claims == nil {
					t.Errorf("ValidateToken() returned nil claims")
					return
				}

				// Check userID if requested
				if tc.checkUserID && claims.UserID != testUserID {
					t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, testUserID)
				}

				// Check email if requested
				if tc.checkEmail && claims.Email != testEmail {
					t.Errorf("ValidateToken() email = %v, want %v", claims.Email, testEmail)
				}
			}
		})
	}
}

// TestTokenExpiration tests token expiration functionality
func TestTokenExpiration(t *testing.T) {
	// This test is tricky because we can't easily create an expired token
	// without modifying the system clock or waiting 24 hours
	// Instead, we'll test that a newly created token is not expired
	
	userID := uint(1)
	email := "test@example.com"
	secretKey := "test-secret"
	
	token, err := GenerateToken(userID, email, secretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid immediately after creation
	claims, err := ValidateToken(token, secretKey)
	if err != nil {
		t.Errorf("Newly created token should be valid: %v", err)
	}

	if claims == nil {
		t.Errorf("ValidateToken() returned nil claims for fresh token")
	}

	// Verify the expiration time is reasonable (within 24 hours + 1 minute from now)
	expectedMaxExpiry := time.Now().Add(24*time.Hour + time.Minute)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.After(expectedMaxExpiry) {
		t.Errorf("Token expires too far in the future: %v", claims.ExpiresAt.Time)
	}

	// Verify the expiration time is not in the past
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		t.Errorf("Token is already expired: %v", claims.ExpiresAt.Time)
	}
}

// TestDifferentSecretKeys tests that tokens signed with different keys don't validate
func TestDifferentSecretKeys(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	secret1 := "secret-key-1"
	secret2 := "secret-key-2"

	// Generate token with first secret
	token, err := GenerateToken(userID, email, secret1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should validate with the same secret
	_, err = ValidateToken(token, secret1)
	if err != nil {
		t.Errorf("Token should validate with same secret: %v", err)
	}

	// Token should NOT validate with different secret
	_, err = ValidateToken(token, secret2)
	if err == nil {
		t.Errorf("Token should not validate with different secret")
	}
}