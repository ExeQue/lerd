package siteops

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/geodro/lerd/internal/config"
)

// Unlinking a site drops its env_provider secrets from tmpfs, so a site lerd no
// longer serves leaves nothing behind for another site's PHP to find.
func TestTeardownSite_RemovesProvidedEnv(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("env_provider is Linux-only")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, "data"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	t.Setenv("XDG_RUNTIME_DIR", filepath.Join(home, "run"))
	file := config.ProvidedEnvFile("app")
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("SECRET=x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	RemoveProvidedEnv = func(name string) { _ = os.Remove(config.ProvidedEnvFile(name)) }
	t.Cleanup(func() { RemoveProvidedEnv = nil })
	TeardownSite(&config.Site{Name: "app", Path: t.TempDir(), Domains: []string{"app.test"}}, nil)
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("provided env should be removed on unlink: %v", err)
	}
}
