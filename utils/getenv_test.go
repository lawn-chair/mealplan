package utils

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test case 1: Environment variable exists
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	result := GetEnv("TEST_ENV_VAR", "fallback_value")
	if result != "test_value" {
		t.Errorf("GetEnv with existing variable failed, got: %s, want: %s", result, "test_value")
	}

	// Test case 2: Environment variable does not exist
	result = GetEnv("NON_EXISTENT_VAR", "fallback_value")
	if result != "fallback_value" {
		t.Errorf("GetEnv with non-existent variable failed, got: %s, want: %s", result, "fallback_value")
	}
}
