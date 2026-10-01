package podman

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDumpBridge_LoadsProvidedEnv runs the real bridge as a prepend and checks
// it loads the site's provided-env file, unquotes values, keeps variables the
// process already has, and loads nothing for a LERD_SITE that walks out of the
// dir or a script outside the roots the file names.
func TestDumpBridge_LoadsProvidedEnv(t *testing.T) {
	php, err := exec.LookPath("php")
	if err != nil {
		t.Skip("php not installed")
	}
	bridge, err := DumpBridgePHP()
	if err != nil {
		t.Fatalf("DumpBridgePHP: %v", err)
	}
	dir := t.TempDir()
	bridgePath := filepath.Join(dir, "dump-bridge.php")
	envDir := filepath.Join(dir, "env")
	if err := os.MkdirAll(envDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bridgePath, []byte(bridge), 0o644); err != nil {
		t.Fatal(err)
	}
	siteDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	otherDir := t.TempDir()
	env := "#lerd-root=" + siteDir + "\n# comment\nPLAIN=one\nSINGLE='two words'\nexport DQ=\"a\\nb\"\nKEPT=from-file\n"
	if err := os.WriteFile(filepath.Join(envDir, "app.env"), []byte(env), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secret.env"), []byte("PLAIN=escaped\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	script := `<?php echo json_encode(array(getenv("PLAIN"), $_ENV["SINGLE"] ?? null, $_SERVER["DQ"] ?? null, getenv("KEPT")));`
	for _, d := range []string{siteDir, otherDir} {
		if err := os.WriteFile(filepath.Join(d, "probe.php"), []byte(script), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	run := func(site, probeDir string) string {
		probe := filepath.Join(probeDir, "probe.php")
		cmd := exec.Command(php, "-n", "-d", "auto_prepend_file="+bridgePath, "-d", "lerd.provided_env_dir="+envDir, probe)
		cmd.Env = append(os.Environ(), "LERD_SITE="+site, "KEPT=from-process")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("php: %v\n%s", err, out)
		}
		return strings.TrimSpace(string(out))
	}

	nothing := `[false,null,null,"from-process"]`
	if out := run("app", siteDir); !strings.HasPrefix(out, "[") {
		t.Skip("php cannot read host files (containerised/sandboxed wrapper); native php needed")
	} else if want := `["one","two words","a\nb","from-process"]`; out != want {
		t.Errorf("loaded env = %s, want %s", out, want)
	}
	if out := run("../secret", siteDir); out != nothing {
		t.Errorf("a traversing LERD_SITE must load nothing, got %s", out)
	}
	if out := run("app", otherDir); out != nothing {
		t.Errorf("a script outside the file's roots must load nothing, got %s", out)
	}
}

// TestFPMQuadlet_MountsProvidedEnvDir checks the FPM unit mounts the tmpfs
// provided-env dir read-only and creates it before start.
func TestFPMQuadlet_MountsProvidedEnvDir(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	mount, pre := providedEnvLines()
	if mount == "" {
		t.Skip("provided env is Linux-only")
	}
	if mount != "Volume=%t/lerd/env:/run/lerd/env:ro" {
		t.Errorf("mount = %q", mount)
	}
	if !strings.Contains(pre, "mkdir -p -m 0700 %t/lerd/env") {
		t.Errorf("ExecStartPre = %q", pre)
	}
	t.Setenv("XDG_RUNTIME_DIR", "")
	if m, p := providedEnvLines(); m != "" || p != "" {
		t.Errorf("without a runtime dir nothing is mounted, got %q %q", m, p)
	}
}
