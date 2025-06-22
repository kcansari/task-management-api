package utils

import (
	// golang.org/x/crypto/bcrypt provides the bcrypt hashing algorithm
	// bcrypt is a password hashing function designed to be slow to prevent brute force attacks
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain text password and returns a bcrypt hash
// The cost parameter determines how slow the hashing will be (higher = more secure but slower)
// bcrypt.DefaultCost (10) is a good balance between security and performance
func HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword() does the actual hashing
	// []byte(password) converts the string to a byte slice (bcrypt works with bytes)
	// The function returns ([]byte, error) - a common Go pattern
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	// Always check for errors in Go - this is the idiomatic way
	if err != nil {
		// Return empty string and the error - Go supports multiple return values
		return "", err
	}
	
	// Convert the byte slice back to string and return with nil error
	// In Go, returning nil for error means "no error occurred"
	return string(hashedBytes), nil
}

// CheckPassword compares a plain text password with a hash to see if they match
// This is used during login to verify the user's password
func CheckPassword(password, hash string) bool {
	// bcrypt.CompareHashAndPassword compares the hash with the plain password
	// It returns an error if they don't match, nil if they do match
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	
	// Return true if no error (passwords match), false if error (passwords don't match)
	// This is a concise way to convert an error to a boolean
	return err == nil
}