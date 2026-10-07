package ui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/geodro/lerd/internal/cli"
	"github.com/geodro/lerd/internal/config"
)

// exit is a seam so the handler can be tested without ending the test binary.
var exit = os.Exit

// cleanUpOnShutdown kills the public tunnels this process started when it is
// asked to stop, and saves the buffered debug events for the next start. Every ordinary stop arrives as a signal: `lerd stop`,
// `launchctl bootout`, a logout, a systemd restart. Without this a tunnel is
// left serving the site publicly, because it runs in a process group of its own
// and macOS has no equivalent of the systemd control-group kill.
func cleanUpOnShutdown() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ch
		cli.StopAllTunnels()
		if srv := dumpsServer.Load(); srv != nil {
			if err := srv.Save(config.DumpsBufferFile()); err != nil {
				fmt.Printf("[WARN] saving debug events: %v\n", err)
			}
		}
		exit(0)
	}()
}
