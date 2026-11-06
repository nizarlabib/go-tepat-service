package infrastructure

import (
	"testing"

	"github.com/spf13/viper"
)

func TestValidateConfigMissingKeys(t *testing.T) {
	// Clear all viper settings
	viper.Reset()

	for _, key := range requiredConfigKeys {
		viper.Set(key, "")
	}

	err := ValidateConfig()
	if err == nil {
		t.Error("expected error when required config keys are missing")
	}
}

func TestValidateConfigAllKeysPresent(t *testing.T) {
	viper.Reset()

	for _, key := range requiredConfigKeys {
		viper.Set(key, "test-value")
	}

	err := ValidateConfig()
	if err != nil {
		t.Errorf("expected no error when all config keys are present, got: %v", err)
	}
}

func TestValidateConfigSomeKeysMissing(t *testing.T) {
	viper.Reset()

	for _, key := range requiredConfigKeys {
		viper.Set(key, "test-value")
	}
	// Remove JWT_KEY specifically
	viper.Set("JWT_KEY", "")

	err := ValidateConfig()
	if err == nil {
		t.Error("expected error when JWT_KEY is missing")
	}
}
