package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDebugbar_ResolveFillsDefaultsAndDropsUnknown(t *testing.T) {
	got := Debugbar{Style: "compact", Edge: "sideways", Theme: "dark"}.Resolve()
	want := Debugbar{Style: "compact", Edge: "bottom", Corner: "bottom-right", Theme: "dark"}
	if got != want {
		t.Fatalf("Resolve = %+v, want %+v", got, want)
	}
}

func TestDebugbar_ValidateNamesUnknownValue(t *testing.T) {
	if err := (Debugbar{Style: "dock", Corner: "top-left"}).Validate(); err != nil {
		t.Fatal(err)
	}
	err := Debugbar{Corner: "middle"}.Validate()
	if err == nil || !strings.Contains(err.Error(), `"middle"`) {
		t.Fatalf("Validate = %v, want an error naming middle", err)
	}
}

// The setting lands in the registry and never in the project's .lerd.yaml,
// since showing the bar is one developer's choice.
func TestSaveDebugbar_KeepsItOutOfTheProject(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("XDG_DATA_HOME", tmp)
	dir := t.TempDir()
	project := []byte("php_version: \"8.4\"\n")
	if err := os.WriteFile(filepath.Join(dir, ".lerd.yaml"), project, 0644); err != nil {
		t.Fatal(err)
	}
	if err := AddSite(Site{Name: "shop", Domains: []string{"shop.test"}, Path: dir}); err != nil {
		t.Fatal(err)
	}
	s, _ := FindSite("shop")
	if err := SaveDebugbar(*s, true); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, ".lerd.yaml")); string(got) != string(project) {
		t.Fatalf(".lerd.yaml changed to %q", got)
	}
	invalidateSitesCache()
	s, _ = FindSite("shop")
	if !DebugbarFor(*s) {
		t.Fatal("DebugbarFor = false after saving true")
	}
}
