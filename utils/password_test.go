package utils

import (
	"testing"
)

// TestHashPassword tests the password hashing functionality
// Go test functions must start with "Test" and take *testing.T parameter
func TestHashPassword(t *testing.T) {
	// Test cases - a slice of anonymous structs
	// This pattern is common in Go testing for multiple test scenarios
	testCases := []struct {
		name     string // Test case name for better error reporting
		password string // Input password to hash
		wantErr  bool   // Whether we expect an error
	}{
		{
			name:     "valid password",
			password: "testpassword123",
			wantErr:  false, // Should not error
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords (though not recommended)
		},
		{
			name:     "long password (within bcrypt limit)",
			password: "this-is-a-long-password-but-under-72-bytes-limit-for-bcrypt",
			wantErr:  false,
		},
		{
			name:     "too long password (over 72 bytes)",
			password: "this-is-a-very-very-very-long-password-that-exceeds-the-72-byte-limit-that-bcrypt-has-for-password-length-and-should-cause-an-error",
			wantErr:  true, // bcrypt will error on passwords over 72 bytes
		},
	}

	// Iterate through test cases using range
	// The range keyword is Go's way to iterate over slices, maps, arrays
	for _, tc := range testCases {
		// t.Run creates a subtest - each test case runs independently
		// This allows better isolation and reporting of individual test failures
		t.Run(tc.name, func(t *testing.T) {
			// Call the function we're testing
			hash, err := HashPassword(tc.password)

			// Check if error expectation matches reality
			if (err != nil) != tc.wantErr {
				// t.Errorf reports an error but continues the test
				// %v is Go's general verb for printing any value
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// If we don't expect an error, validate the hash
			if !tc.wantErr {
				// Hash should not be empty
				if hash == "" {
					t.Errorf("HashPassword() returned empty hash")
				}

				// Hash should be different from original password
				if hash == tc.password {
					t.Errorf("HashPassword() returned same as input password")
				}

				// Hash should have bcrypt prefix ($2a$ or $2b$)
				// bcrypt hashes start with version identifier
				if len(hash) < 4 || (hash[:4] != "$2a$" && hash[:4] != "$2b$") {
					t.Errorf("HashPassword() returned invalid bcrypt hash format: %s", hash)
				}
			}
		})
	}
}

// TestCheckPassword tests password verification functionality
func TestCheckPassword(t *testing.T) {
	// First, create a known hash for testing
	testPassword := "testpassword123"
	hash, err := HashPassword(testPassword)
	if err != nil {
		// t.Fatalf stops the test immediately on fatal error
		// Use this when the test cannot continue without this setup
		t.Fatalf("Failed to create test hash: %v", err)
	}

	testCases := []struct {
		name     string
		password string // Password to check
		hash     string // Hash to check against
		want     bool   // Expected result (true = passwords match)
	}{
		{
			name:     "correct password",
			password: testPassword,
			hash:     hash,
			want:     true, // Should match
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     hash,
			want:     false, // Should not match
		},
		{
			name:     "empty password against real hash",
			password: "",
			hash:     hash,
			want:     false, // Should not match
		},
		{
			name:     "password against empty hash",
			password: testPassword,
			hash:     "",
			want:     false, // Should not match (invalid hash)
		},
		{
			name:     "password against invalid hash",
			password: testPassword,
			hash:     "invalid-hash-format",
			want:     false, // Should not match (malformed hash)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function we're testing
			got := CheckPassword(tc.password, tc.hash)

			// Check if result matches expectation
			if got != tc.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestHashPasswordConsistency tests that the same password produces different hashes
// This is important for security - bcrypt should use random salts
func TestHashPasswordConsistency(t *testing.T) {
	password := "testpassword123"

	// Generate two hashes for the same password
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	// Both should succeed
	if err1 != nil {
		t.Fatalf("First HashPassword() failed: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("Second HashPassword() failed: %v", err2)
	}

	// Hashes should be different (due to random salt)
	// This proves that bcrypt is using proper salt generation
	if hash1 == hash2 {
		t.Errorf("HashPassword() produced identical hashes for same password - salt randomization may be broken")
	}

	// But both hashes should validate against the original password
	if !CheckPassword(password, hash1) {
		t.Errorf("First hash does not validate against original password")
	}
	if !CheckPassword(password, hash2) {
		t.Errorf("Second hash does not validate against original password")
	}
}