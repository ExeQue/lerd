package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/geodro/lerd/internal/config"
	gitpkg "github.com/geodro/lerd/internal/git"
)

// refreshProvidedEnv runs the project's env_provider on the host and keeps its
// dotenv output in the site's tmpfs file, which the PHP prepend loads for web
// requests, workers and CLI alike. A project without a provider has any stale
// file removed so secrets do not outlive the setting. approve records consent
// up front, for `lerd env --yes` where there is no terminal to prompt on.
func refreshProvidedEnv(site config.Site, approve bool) error {
	file := config.ProvidedEnvFile(site.Name)
	proj, _ := config.LoadProjectConfig(site.Path)
	if proj == nil || proj.EnvProvider == "" {
		if file != "" {
			_ = os.Remove(file)
		}
		return nil
	}
	if file == "" {
		return errors.New("env_provider needs a tmpfs runtime dir (XDG_RUNTIME_DIR) and is Linux-only for now")
	}
	if approve {
		if err := config.ApproveSiteCommand(site.Name, proj.EnvProvider); err != nil {
			return err
		}
	}
	if err := approveHostCommand(site.Name, proj.EnvProvider, "env_provider"); err != nil {
		return err
	}

	var stdout bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", proj.EnvProvider)
	cmd.Dir = site.Path
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("env_provider failed: %w", err)
	}
	return writeProvidedEnv(file, append(providedEnvHeader(site), stdout.Bytes()...))
}

// providedEnvHeader names the directories the file may be loaded from: the site
// and its worktrees. The prepend refuses a script outside them, so a stale or
// wrong LERD_SITE (a shell cd'd into another project) loads nothing.
func providedEnvHeader(site config.Site) []byte {
	roots := []string{site.Path}
	if wts, _ := gitpkg.DetectWorktrees(site.Path, site.PrimaryDomain()); len(wts) > 0 {
		for _, wt := range wts {
			roots = append(roots, wt.Path)
		}
	}
	var b bytes.Buffer
	for _, r := range roots {
		if real, err := filepath.EvalSymlinks(r); err == nil {
			r = real
		}
		if strings.ContainsAny(r, "\r\n") {
			continue
		}
		fmt.Fprintf(&b, "#lerd-root=%s\n", r)
	}
	return b.Bytes()
}

// writeProvidedEnv replaces the file atomically with owner-only permissions, so
// PHP never reads a half-written file and no other user can read it.
func writeProvidedEnv(file string, data []byte) error {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".provided-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), file)
}
