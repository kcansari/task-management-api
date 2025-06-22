package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/kcansari/task-management-api/config"
	"github.com/kcansari/task-management-api/database"
	"github.com/kcansari/task-management-api/models"
	"github.com/kcansari/task-management-api/utils"
)

// RegisterRequest represents the data needed to register a new user
// JSON tags tell Go how to map JSON fields to struct fields
// The `binding:"required"` tag would be used by frameworks like Gin for validation
type RegisterRequest struct {
	Email    string `json:"email"`    // User's email address
	Password string `json:"password"` // Plain text password (will be hashed)
}

// LoginRequest represents the data needed to log in
type LoginRequest struct {
	Email    string `json:"email"`    // User's email address
	Password string `json:"password"` // Plain text password to verify
}

// AuthResponse represents what we send back after successful authentication
type AuthResponse struct {
	Token string      `json:"token"` // JWT token for future requests
	User  models.User `json:"user"`  // User information (without password)
}

// ErrorResponse represents an error message we send to clients
type ErrorResponse struct {
	Error string `json:"error"` // Human-readable error message
}

// Register handles user registration (POST /api/auth/register)
// http.ResponseWriter is used to write the HTTP response
// *http.Request contains the incoming HTTP request data
func Register(w http.ResponseWriter, r *http.Request) {
	// Set response content type to JSON
	// This tells the client what format the response will be in
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST method for registration
	// HTTP methods have specific meanings: POST = create new resource
	if r.Method != "POST" {
		// http.StatusMethodNotAllowed = 405
		w.WriteHeader(http.StatusMethodNotAllowed)
		// json.NewEncoder(w).Encode() converts a Go struct to JSON and writes to response
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Parse the JSON request body into our RegisterRequest struct
	var req RegisterRequest
	// json.NewDecoder(r.Body).Decode() reads JSON from request and converts to Go struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If JSON is malformed, return 400 Bad Request
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
		return
	}

	// Basic validation - check if required fields are provided
	// strings.TrimSpace() removes leading/trailing whitespace
	if strings.TrimSpace(req.Email) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Email is required"})
		return
	}

	if strings.TrimSpace(req.Password) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Password is required"})
		return
	}

	// Check if user with this email already exists
	// database.GetDB() returns our GORM database instance
	db := database.GetDB()
	var existingUser models.User
	// GORM's Where().First() tries to find one record matching the condition
	// If no record found, it returns an error
	result := db.Where("email = ?", req.Email).First(&existingUser)
	
	// Check if we found a user (no error means user exists)
	if result.Error == nil {
		// User already exists - return conflict error
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User with this email already exists"})
		return
	}

	// Hash the password before storing it
	// NEVER store plain text passwords in the database!
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		// If hashing fails, return internal server error
		log.Printf("Failed to hash password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to process password"})
		return
	}

	// Create a new user struct with the provided data
	user := models.User{
		Email:    req.Email,
		Password: hashedPassword, // Store the hashed password, not the plain text
	}

	// Save the user to the database
	// GORM's Create() inserts a new record and updates the struct with the generated ID
	if err := db.Create(&user).Error; err != nil {
		log.Printf("Failed to create user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create user"})
		return
	}

	// Generate a JWT token for the new user
	// Load config to get the JWT secret key
	cfg := config.Load()
	token, err := utils.GenerateToken(user.ID, user.Email, cfg.JWTSecret)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Clear the password field before sending user data to client
	// The `json:"-"` tag in the model already excludes it, but this is extra safety
	user.Password = ""

	// Return success response with token and user data
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user authentication (POST /api/auth/login)
func Login(w http.ResponseWriter, r *http.Request) {
	// Set JSON content type
	w.Header().Set("Content-Type", "application/json")

	// Only allow POST method
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Parse login request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Email and password are required"})
		return
	}

	// Find user by email
	db := database.GetDB()
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// User not found - return generic error for security
		// Don't reveal whether email exists or not to prevent email enumeration attacks
		w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// Check if the provided password matches the stored hash
	if !utils.CheckPassword(req.Password, user.Password) {
		// Password doesn't match - return same generic error
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// Generate JWT token for successful login
	cfg := config.Load()
	token, err := utils.GenerateToken(user.ID, user.Email, cfg.JWTSecret)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Clear password before sending response
	user.Password = ""

	// Return success response
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}