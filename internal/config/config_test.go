package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.WeekStart != "monday" {
		t.Errorf("expected week_start 'monday', got %s", cfg.WeekStart)
	}

	if cfg.AuthorAliases == nil {
		t.Error("expected author_aliases to be initialized (empty slice)")
	}

	if cfg.Timezone == "" {
		t.Error("expected timezone to be auto-detected, got empty string")
	}

	// Timezone should be detected (could be "UTC" or system timezone)
	// We just verify it's not empty
	if cfg.Timezone == "" {
		t.Error("timezone should not be empty after auto-detection")
	}
}

func TestLoad_WithAuthorAliases(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("IKNO_HOME", tmpDir)

	// Create config file with author_aliases
	configContent := `week_start: monday
author_email: primary@example.com
author_aliases:
  - work@company.com
  - personal@example.com
timezone: Europe/Athens
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AuthorEmail != "primary@example.com" {
		t.Errorf("expected author_email 'primary@example.com', got %s", cfg.AuthorEmail)
	}

	if len(cfg.AuthorAliases) != 2 {
		t.Fatalf("expected 2 author_aliases, got %d", len(cfg.AuthorAliases))
	}

	if cfg.AuthorAliases[0] != "work@company.com" {
		t.Errorf("expected first alias 'work@company.com', got %s", cfg.AuthorAliases[0])
	}

	if cfg.AuthorAliases[1] != "personal@example.com" {
		t.Errorf("expected second alias 'personal@example.com', got %s", cfg.AuthorAliases[1])
	}

	if cfg.Timezone != "Europe/Athens" {
		t.Errorf("expected timezone 'Europe/Athens', got %s", cfg.Timezone)
	}
}

func TestLoad_AutoDetectTimezone(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("IKNO_HOME", tmpDir)

	// Create config file without timezone
	configContent := `week_start: monday
author_email: test@example.com
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Timezone should be auto-detected
	if cfg.Timezone == "" {
		t.Error("expected timezone to be auto-detected when not set in config")
	}
}

func TestLoad_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("IKNO_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed when config doesn't exist: %v", err)
	}

	// Should return defaults
	if cfg.WeekStart != "monday" {
		t.Errorf("expected default week_start 'monday', got %s", cfg.WeekStart)
	}

	if cfg.AuthorAliases == nil {
		t.Error("expected author_aliases to be initialized")
	}

	if cfg.Timezone == "" {
		t.Error("expected timezone to be auto-detected")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("IKNO_HOME", tmpDir)

	cfg := &Config{
		WeekStart:     "sunday",
		AuthorEmail:   "test@example.com",
		AuthorAliases: []string{"work@company.com", "personal@example.com"},
		Timezone:      "America/New_York",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load it back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.WeekStart != "sunday" {
		t.Errorf("expected week_start 'sunday', got %s", loaded.WeekStart)
	}

	if loaded.AuthorEmail != "test@example.com" {
		t.Errorf("expected author_email 'test@example.com', got %s", loaded.AuthorEmail)
	}

	if len(loaded.AuthorAliases) != 2 {
		t.Fatalf("expected 2 author_aliases, got %d", len(loaded.AuthorAliases))
	}

	if loaded.Timezone != "America/New_York" {
		t.Errorf("expected timezone 'America/New_York', got %s", loaded.Timezone)
	}
}

func TestEnsureConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("IKNO_HOME", tmpDir)

	configPath, err := EnsureConfigFile()
	if err != nil {
		t.Fatalf("EnsureConfigFile failed: %v", err)
	}

	// File should exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Should contain the template
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("config file is empty")
	}

	// Verify it contains the new fields in comments
	if !containsString(contentStr, "author_aliases") {
		t.Error("config template should mention author_aliases")
	}

	if !containsString(contentStr, "timezone") {
		t.Error("config template should mention timezone")
	}
}

func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("IKNO_HOME", tmpDir)

	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "config.yaml")
	if configPath != expected {
		t.Errorf("expected config path %s, got %s", expected, configPath)
	}
}

func TestDetectTimezone(t *testing.T) {
	tz := detectTimezone()

	if tz == "" {
		t.Error("detectTimezone returned empty string")
	}

	// Should not be "Local" (the raw Location.String() for local timezone)
	if tz == "Local" {
		t.Error("detectTimezone should convert 'Local' to 'UTC'")
	}
}

func TestLoad_WithAIHTTPTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("IKNO_HOME", tmpDir)

	tests := []struct {
		name     string
		yaml     string
		expected Duration
	}{
		{
			name:     "seconds",
			yaml:     "ai_http_timeout: 30s",
			expected: Duration(30 * time.Second),
		},
		{
			name:     "minutes",
			yaml:     "ai_http_timeout: 2m",
			expected: Duration(2 * time.Minute),
		},
		{
			name:     "mixed",
			yaml:     "ai_http_timeout: 1m30s",
			expected: Duration(90 * time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configContent := "week_start: monday\n" + tt.yaml + "\n"
			configPath := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatal(err)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}

			if cfg.AIHTTPTimeout != tt.expected {
				t.Errorf("expected ai_http_timeout %v, got %v", tt.expected, cfg.AIHTTPTimeout)
			}

			// Verify ToDuration() conversion works
			if cfg.AIHTTPTimeout.ToDuration() != tt.expected.ToDuration() {
				t.Errorf("ToDuration() mismatch: expected %v, got %v",
					tt.expected.ToDuration(), cfg.AIHTTPTimeout.ToDuration())
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
