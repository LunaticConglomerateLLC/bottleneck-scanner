package main

import (
	"os"
	"testing"
)

// writeTeamFile writes content to a temporary YAML file and returns its path.
// The file is automatically removed when the test ends via t.TempDir().
func writeTeamFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "team-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestLoadTeamConfig(t *testing.T) {
	t.Run("valid full config", func(t *testing.T) {
		path := writeTeamFile(t, `
name: my-team
repos:
  - org/repo1
  - org/repo2
members:
  - alice
  - bob
service_accounts:
  - renovate-bot
`)
		cfg, err := loadTeamConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Name != "my-team" {
			t.Errorf("expected name 'my-team', got %q", cfg.Name)
		}
		if len(cfg.Repos) != 2 {
			t.Errorf("expected 2 repos, got %d", len(cfg.Repos))
		}
		if len(cfg.Members) != 2 {
			t.Errorf("expected 2 members, got %d", len(cfg.Members))
		}
		if len(cfg.ServiceAccounts) != 1 || cfg.ServiceAccounts[0] != "renovate-bot" {
			t.Errorf("unexpected service_accounts: %v", cfg.ServiceAccounts)
		}
	})

	t.Run("minimal valid config", func(t *testing.T) {
		path := writeTeamFile(t, "name: t\nrepos:\n  - org/repo\n")
		cfg, err := loadTeamConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Repos[0] != "org/repo" {
			t.Errorf("expected org/repo, got %q", cfg.Repos[0])
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := loadTeamConfig("/nonexistent/path/team.yml")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		path := writeTeamFile(t, "repos: [unclosed\nname: {bad")
		_, err := loadTeamConfig(path)
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})

	t.Run("no repos returns error", func(t *testing.T) {
		path := writeTeamFile(t, "name: my-team\nmembers:\n  - alice\n")
		_, err := loadTeamConfig(path)
		if err == nil {
			t.Error("expected error when no repos listed")
		}
	})
}
