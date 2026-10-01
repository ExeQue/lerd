package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/config"
)

// Console commands (lerd artisan …) must carry LERD_SITE, or the prepend
// cannot pick the site's provided env and artisan runs without its secrets.
func TestConsoleCmdArgs_CarriesLerdSite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, "data"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	site := config.Site{Name: "app", Path: t.TempDir()}
	if err := config.SaveSites(&config.SiteRegistry{Sites: []config.Site{site}}); err != nil {
		t.Fatalf("SaveSites: %v", err)
	}
	args := consoleCmdArgs(site.Path, "lerd-php85-fpm", "artisan", false, []string{"about"})
	for i, a := range args {
		if a == "LERD_SITE=app" && i > 0 && args[i-1] == "--env" {
			return
		}
	}
	t.Errorf("console exec is missing --env LERD_SITE=app: %v", args)
}

func TestValidProvidedEnvSite(t *testing.T) {
	for name, want := range map[string]bool{
		"app": true, "my-app.v2_x": true, "": false, "..": false, "a/b": false,
		"a b": false, "a;rm -rf /": false, "a'b": false, "a$(x)": false,
	} {
		if got := validProvidedEnvSite(name); got != want {
			t.Errorf("validProvidedEnvSite(%q) = %v, want %v", name, got, want)
		}
	}
}

// The macOS VM scripts write through sudo into the VM's tmpfs, owner-only,
// atomically, with the values on stdin and never in the command line.
func TestProvidedEnvVMScripts(t *testing.T) {
	args := providedEnvSSHArgs("lerd", providedEnvWriteScript("app"))
	if strings.Join(args[:3], " ") != "machine ssh lerd" || len(args) != 4 {
		t.Fatalf("ssh args = %v", args)
	}
	write := args[3]
	for _, want := range []string{"sudo sh -c '", "umask 077", "mkdir -p /run/lerd/env", "cat > /run/lerd/env/.provided-app", "mv /run/lerd/env/.provided-app /run/lerd/env/app.env"} {
		if !strings.Contains(write, want) {
			t.Errorf("write script %q is missing %q", write, want)
		}
	}
	if got := providedEnvRemoveScript("app"); got != "sudo rm -f /run/lerd/env/app.env" {
		t.Errorf("remove script = %q", got)
	}
	if got := providedEnvMkdirScript(); got != "sudo sh -c 'umask 077; mkdir -p /run/lerd/env'" {
		t.Errorf("mkdir script = %q", got)
	}
}
