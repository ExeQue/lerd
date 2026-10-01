//go:build darwin

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/geodro/lerd/internal/config"
	"github.com/geodro/lerd/internal/podman"
)

// On macOS the FPM containers run in the Podman Machine VM, so the provided env
// lives in the VM's tmpfs. lerd writes it over `podman machine ssh` on stdin:
// the values go from the provider's stdout into VM memory and never touch the
// Mac's disk or an argv. lerd's containers run rootful in the VM, hence sudo.

func providedEnvSupported() error {
	if cfg, err := config.LoadGlobal(); err == nil && cfg.PHPRuntimeMode() == config.PHPRuntimeNative {
		return errors.New("env_provider is not supported on the native PHP runtime yet")
	}
	if selectedMachineName() == "" {
		return errors.New("env_provider needs the Podman Machine to be set up")
	}
	return nil
}

// ensureProvidedEnvDir creates the VM dir the FPM containers mount. The launchd
// units have no ExecStartPre, and podman refuses a missing bind source, so this
// runs on lerd start before any container comes up.
func ensureProvidedEnvDir() {
	if providedEnvSupported() != nil {
		return
	}
	if err := providedEnvSSH(providedEnvMkdirScript(), nil); err != nil {
		fmt.Fprintf(os.Stderr, "creating %s in the Podman Machine: %v\n", podman.ProvidedEnvVMDir, err)
	}
}

func storeProvidedEnv(siteName string, data []byte) error {
	return providedEnvSSH(providedEnvWriteScript(siteName), data)
}

func dropProvidedEnv(siteName string) {
	if !validProvidedEnvSite(siteName) || providedEnvSupported() != nil {
		return
	}
	_ = providedEnvSSH(providedEnvRemoveScript(siteName), nil)
}

func providedEnvSSH(script string, stdin []byte) error {
	cmd := podman.Cmd(providedEnvSSHArgs(selectedMachineName(), script)...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, bytes.TrimSpace(stderr.Bytes()))
	}
	return nil
}
