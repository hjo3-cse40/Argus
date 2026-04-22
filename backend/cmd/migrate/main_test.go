package main

import (
	"os"
	"testing"

	"argus-backend/internal/config"
)

// TestConfigurationLoading verifies the migration command can load configuration
func TestConfigurationLoading(t *testing.T) {
	// Set up test environment variables
	_ = os.Setenv("DB_HOST", "localhost")
	_ = os.Setenv("DB_PORT", "5432")
	_ = os.Setenv("DB_USER", "argus")
	_ = os.Setenv("DB_PASSWORD", "argus")
	_ = os.Setenv("DB_NAME", "argus")
	_ = os.Setenv("ENV", "dev")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify database configuration is loaded
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected DB_HOST 'localhost', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("Expected DB_PORT '5432', got '%s'", cfg.Database.Port)
	}
	if cfg.Database.User != "argus" {
		t.Errorf("Expected DB_USER 'argus', got '%s'", cfg.Database.User)
	}
	if cfg.Database.DBName != "argus" {
		t.Errorf("Expected DB_NAME 'argus', got '%s'", cfg.Database.DBName)
	}

	// Verify connection string is properly formatted
	connStr := cfg.Database.ConnectionString()
	if connStr == "" {
		t.Error("Connection string is empty")
	}

	expectedConnStr := "host=localhost port=5432 user=argus password=argus dbname=argus sslmode=disable"
	if connStr != expectedConnStr {
		t.Errorf("Expected connection string '%s', got '%s'", expectedConnStr, connStr)
	}
}

// TestConfigurationValidation verifies proper error handling for invalid configuration
func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func()
		expectedToFail bool
	}{
		{
			name: "valid configuration",
			setupEnv: func() {
				_ = os.Setenv("DB_HOST", "localhost")
				_ = os.Setenv("DB_PORT", "5432")
				_ = os.Setenv("DB_USER", "argus")
				_ = os.Setenv("DB_PASSWORD", "argus")
				_ = os.Setenv("DB_NAME", "argus")
				_ = os.Setenv("ENV", "dev")
			},
			expectedToFail: false,
		},
		{
			name: "invalid environment",
			setupEnv: func() {
				_ = os.Setenv("ENV", "invalid")
				_ = os.Setenv("DB_HOST", "localhost")
				_ = os.Setenv("DB_PORT", "5432")
				_ = os.Setenv("DB_USER", "argus")
				_ = os.Setenv("DB_PASSWORD", "argus")
				_ = os.Setenv("DB_NAME", "argus")
			},
			expectedToFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Setup test environment
			tt.setupEnv()

			// Load configuration
			cfg, err := config.Load()
			if tt.expectedToFail {
				if err == nil {
					t.Error("Expected configuration load to fail, but it succeeded")
				}
				return
			}

			if err != nil {
				t.Errorf("Configuration load failed: %v", err)
				return
			}

			// Verify configuration is valid
			if cfg.Database.Host == "" {
				t.Error("Database host not configured")
			}
		})
	}
}

// TestConnectionStringFormat verifies the connection string format
func TestConnectionStringFormat(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		user     string
		password string
		dbname   string
		expected string
	}{
		{
			name:     "standard configuration",
			host:     "localhost",
			port:     "5432",
			user:     "argus",
			password: "argus",
			dbname:   "argus",
			expected: "host=localhost port=5432 user=argus password=argus dbname=argus sslmode=disable",
		},
		{
			name:     "production configuration",
			host:     "db.example.com",
			port:     "5432",
			user:     "prod_user",
			password: "prod_pass",
			dbname:   "argus_prod",
			expected: "host=db.example.com port=5432 user=prod_user password=prod_pass dbname=argus_prod sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			_ = os.Setenv("DB_HOST", tt.host)
			_ = os.Setenv("DB_PORT", tt.port)
			_ = os.Setenv("DB_USER", tt.user)
			_ = os.Setenv("DB_PASSWORD", tt.password)
			_ = os.Setenv("DB_NAME", tt.dbname)
			_ = os.Setenv("ENV", "dev")

			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("Failed to load configuration: %v", err)
			}

			connStr := cfg.Database.ConnectionString()
			if connStr != tt.expected {
				t.Errorf("Expected connection string '%s', got '%s'", tt.expected, connStr)
			}
		})
	}
}
