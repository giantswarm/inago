package fleet

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/coreos/fleet/ssh"
)

// SSHTunnel contains the information needed to create a new ssh client
type SSHTunnel struct {
	Tunnel                string
	Username              string
	Timeout               time.Duration
	StrictHostKeyChecking bool
	KnownHostsFile        string
}

// NewDialFunc returns an http.Dial function which uses the given ssh tunnel as
// to connect to the endpoint
func (t *SSHTunnel) NewDialFunc(config Config) func(string, string) (net.Conn, error) {
	sshClient, err := ssh.NewSSHClient(t.Username, t.Tunnel, t.getChecker(), true, t.Timeout)
	if err != nil {
		fmt.Printf("failed to initialize ssh client: %v\n", maskAny(err))
		os.Exit(1)
	}
	tunnelFunc := sshClient.Dial

	if config.Endpoint.Scheme == "unix" || config.Endpoint.Scheme == "file" {
		tgt := config.Endpoint.Path
		tunnelFunc = func(string, string) (net.Conn, error) {
			cmd := fmt.Sprintf(`fleetctl fd-forward %s`, tgt)
			return ssh.DialCommand(sshClient, cmd)
		}
	}

	return tunnelFunc
}

// getChecker creates and returns a HostKeyChecker, or nil if any error is encountered
func (t *SSHTunnel) getChecker() *ssh.HostKeyChecker {
	if !t.StrictHostKeyChecking {
		return nil
	}

	keyFile := ssh.NewHostKeyFile(t.KnownHostsFile)
	return ssh.NewHostKeyChecker(keyFile)
}
