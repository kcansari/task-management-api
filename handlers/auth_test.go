package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kcansari/task-management-api/config"
	"github.com/kcansari/task-management-api/database"
)

// setupTestDB initializes a test database connection
// In a real application, you'd use a separate test database
func setupTestDB(t *testing.T) {
	// Load test configuration
	cfg := config.Load()
	
	// Initialize database
	if err := database.Initialize(cfg); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	
	// Clean up any existing test data
	// In production, you'd use database transactions that rollback
	db := database.GetDB()
	db.Exec("DELETE FROM tasks WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%test%')")
	db.Exec("DELETE FROM users WHERE email LIKE '%test%'")
}

// TestRegisterHandler tests the user registration endpoint
func TestRegisterHandler(t *testing.T) {
	// Setup test database
	setupTestDB(t)

	testCases := []struct {
		name           string
		requestBody    interface{} // What JSON to send
		expectedStatus int         // Expected HTTP status code
		checkResponse  bool        // Whether to validate response body
	}{
		{
			name: "valid registration",
			requestBody: RegisterRequest{
				Email:    "test-register@example.com",
				Password: "testpassword123",
			},
			expectedStatus: http.StatusCreated, // 201
			checkResponse:  true,
		},
		{
			name: "missing email",
			requestBody: RegisterRequest{
				Email:    "",
				Password: "testpassword123",
			},
			expectedStatus: http.StatusBadRequest, // 400
			checkResponse:  false,
		},
		{
			name: "missing password",
			requestBody: RegisterRequest{
				Email:    "test-nopass@example.com",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest, // 400
			checkResponse:  false,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid-json-string",
			expectedStatus: http.StatusBadRequest, // 400
			checkResponse:  false,
		},
		{
			name: "duplicate email",
			requestBody: RegisterRequest{
				Email:    "test-register@example.com", // Same as first test
				Password: "anotherpassword",
			},
			expectedStatus: http.StatusConflict, // 409
			checkResponse:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert request body to JSON
			var requestBody []byte
			var err error
			
			if str, ok := tc.requestBody.(string); ok {
				// For invalid JSON test case
				requestBody = []byte(str)
			} else {
				// For normal struct cases
				requestBody, err = json.Marshal(tc.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			// Create HTTP request
			// httptest.NewRequest creates a test HTTP request
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder to capture handler output
			// httptest.NewRecorder implements http.ResponseWriter for testing
			rr := httptest.NewRecorder()

			// Call the handler
			Register(rr, req)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			// Check response body if needed
			if tc.checkResponse && tc.expectedStatus == http.StatusCreated {
				var response AuthResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				// Validate response structure
				if response.Token == "" {
					t.Errorf("Expected token in response, got empty string")
				}

				if response.User.Email == "" {
					t.Errorf("Expected user email in response, got empty string")
				}

				if response.User.ID == 0 {
					t.Errorf("Expected user ID in response, got 0")
				}
			}

			// Validate Content-Type header
			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}

// TestLoginHandler tests the user login endpoint
func TestLoginHandler(t *testing.T) {
	// Setup test database
	setupTestDB(t)

	// Create a test user first
	testEmail := "test-login@example.com"
	testPassword := "testpassword123"
	
	// Register user using the handler to ensure consistent setup
	registerReq := RegisterRequest{
		Email:    testEmail,
		Password: testPassword,
	}
	registerBody, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	Register(rr, req)
	
	if rr.Code != http.StatusCreated {
		t.Fatalf("Failed to create test user: status %d", rr.Code)
	}

	testCases := []struct {
		name           string
		requestBody    LoginRequest
		expectedStatus int
		checkResponse  bool
	}{
		{
			name: "valid login",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: testPassword,
			},
			expectedStatus: http.StatusOK, // 200
			checkResponse:  true,
		},
		{
			name: "wrong password",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized, // 401
			checkResponse:  false,
		},
		{
			name: "non-existent user",
			requestBody: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: testPassword,
			},
			expectedStatus: http.StatusUnauthorized, // 401
			checkResponse:  false,
		},
		{
			name: "missing email",
			requestBody: LoginRequest{
				Email:    "",
				Password: testPassword,
			},
			expectedStatus: http.StatusBadRequest, // 400
			checkResponse:  false,
		},
		{
			name: "missing password",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: "",
			},
			expectedStatus: http.StatusBadRequest, // 400
			checkResponse:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			requestBody, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Call handler
			Login(rr, req)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			// Check response body for successful login
			if tc.checkResponse && tc.expectedStatus == http.StatusOK {
				var response AuthResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				// Validate response structure
				if response.Token == "" {
					t.Errorf("Expected token in response, got empty string")
				}

				if response.User.Email != testEmail {
					t.Errorf("Expected user email %s, got %s", testEmail, response.User.Email)
				}
			}
		})
	}
}

// TestMethodNotAllowed tests that auth endpoints reject non-POST methods
func TestMethodNotAllowed(t *testing.T) {
	setupTestDB(t)

	// Test different HTTP methods on register endpoint
	methods := []string{"GET", "PUT", "DELETE", "PATCH"}
	
	for _, method := range methods {
		t.Run("register_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/auth/register", nil)
			rr := httptest.NewRecorder()
			
			Register(rr, req)
			
			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for %s method, got %d", 
					http.StatusMethodNotAllowed, method, rr.Code)
			}
		})

		t.Run("login_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/auth/login", nil)
			rr := httptest.NewRecorder()
			
			Login(rr, req)
			
			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for %s method, got %d", 
					http.StatusMethodNotAllowed, method, rr.Code)
			}
		})
	}
}