package config_test

import (
	"os"
	"testing"

	"github.com/orimono/hari/internal/config"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{
			name:      "File should be parsed",
			path:      "testdata/config.json",
			expectErr: false,
		},
		{
			name:      "Should have error occurs",
			path:      "",
			expectErr: true,
		},
		{
			name:      "Non-existent file",
			path:      "not_found.json",
			expectErr: true,
		},
		{
			name:      "Malformed JSON content",
			path:      "testdata/bad_syntax.json",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var c config.Config
			err := config.ReadFromFile(test.path, &c)
			if (err != nil) != test.expectErr {
				t.Errorf("Unexpected error status: %v", err)
			}
		})
	}

	t.Run("Load default config", func(t *testing.T) {
		_, err := config.Load()
		if err == nil {
			t.Errorf("Should have been failed: %v", err)
		}
	})

	t.Run("Load default config with env", func(t *testing.T) {
		os.Setenv("HARI_CONFIG_PATH", "./testdata/config.json")
		_, err := config.Load()
		if err != nil {
			t.Errorf("Should have been no error occurs, got: %v", err)
		}
	})
}
