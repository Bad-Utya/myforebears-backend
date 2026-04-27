package config_test

import (
	"testing"

	cfg "github.com/Bad-Utya/myforebears-backend/services/emailprovider/internal/config"
)

func TestConfigTypeExists(t *testing.T) {
	var _ cfg.Config
}
