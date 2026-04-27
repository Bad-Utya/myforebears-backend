package config_test

import (
	"testing"

	cfg "github.com/Bad-Utya/myforebears-backend/services/auth/internal/config"
)

func TestConfigTypeExists(t *testing.T) {
	// Ensure the Config type compiles and is addressable.
	var _ cfg.Config
}
