package fleet

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/coreos/fleet/ssh"
	"github.com/giantswarm/inago/logging"
)

// SSHTunnelConfig contains the information needed to create a new ssh tunnel.
type SSHTunnelConfig struct {
	Endpoint              url.URL
	KnownHostsFile        string
	Logger                logging.Logger
	StrictHostKeyChecking bool
	Tunnel                string
	Timeout               time.Duration
	Username              string
}

func DefaultSSHTunnelConfig() SSHTunnelConfig {
	URL, err := url.Parse("file:///var/run/fleet.sock")
	if err != nil {
		panic(err)
	}

	newConfig := SSHTunnelConfig{
		Endpoint:              *URL,
		KnownHostsFile:        "~/.fleetctl/known_hosts",
		Logger:                logging.NewLogger(logging.DefaultConfig()),
		StrictHostKeyChecking: true,
		Tunnel:                "",
		Timeout:               10 * time.Second,
		Username:              "core",
	}

	return newConfig
}

// SSHTunnel provides a synchronized wrapper around http.RoundTripper for safe
// use of tunneled ssh connections.
type SSHTunnel interface {
	http.RoundTripper

	// IsActive determines whether the tunnel should be used. This is decided by
	// the given tunnel configuration. In case this is empty, the SSH tunnel is
	// not active.
	IsActive() bool
}

// NewSSHTunnel creates a new SSH tunnel that is configured with the given
// settings.
func NewSSHTunnel(config SSHTunnelConfig) (SSHTunnel, error) {
	newSSHTunnel := &sshTunnel{
		SSHTunnelConfig: config,
		Mutex:           sync.Mutex{},
	}

	newDialFunc, err := newSSHTunnel.NewDialFunc()
	if err != nil {
		return nil, maskAny(err)
	}
	newSSHTunnel.HTTPTransport = &http.Transport{
		Dial: newDialFunc,
	}

	return newSSHTunnel, nil
}

type sshTunnel struct {
	SSHTunnelConfig

	HTTPTransport http.RoundTripper
	Mutex         sync.Mutex
}

func (t *sshTunnel) IsActive() bool {
	return t.Tunnel != ""
}

// NewDialFunc returns an http.Dial function which uses the given ssh tunnel to
// to connect to the configured endpoint.
func (t *sshTunnel) NewDialFunc() (func(string, string) (net.Conn, error), error) {
	forwardAgent := true
	sshClient, err := ssh.NewSSHClient(t.Username, t.Tunnel, t.NewHostKeyChecker(), forwardAgent, t.Timeout)
	if err != nil {
		return nil, maskAny(err)
	}

	if t.Endpoint.Scheme == "unix" || t.Endpoint.Scheme == "file" {
		return func(string, string) (net.Conn, error) {
			cmd := fmt.Sprintf(`fleetctl fd-forward %s`, t.Endpoint.Path)
			return ssh.DialCommand(sshClient, cmd)
		}, nil
	}

	return sshClient.Dial, nil
}

// NewHostKeyChecker creates a new HostKeyChecker, or nil if any error is
// encountered.
func (t *sshTunnel) NewHostKeyChecker() *ssh.HostKeyChecker {
	if !t.StrictHostKeyChecking {
		return nil
	}

	keyFile := ssh.NewHostKeyFile(t.KnownHostsFile)
	return ssh.NewHostKeyChecker(keyFile)
}

// RoundTrip provides implementation of http.RoundTripper. This method is
// simply a synchronized wrapper to the underlying http.Transport.
func (t *sshTunnel) RoundTrip(req *http.Request) (*http.Response, error) {
	t.Mutex.Lock()

	fmt.Printf("req.URL.Path: %#v?%s\n", req.URL.Path, req.URL.Query().Encode())
	resp, err := t.HTTPTransport.RoundTrip(req)

	/**
	Learning:

	The fleet ssh tunnel does not support (for unknown reasons) more than one
	command-session in parallel. Just adding a mutex here does not 100% help,
	as the connection is only returned to the idle pool when the body gets closed.
	If multiple go routines try to call RoundTrip() the first one performs the request,
	then return the response. Then the second goroutine would start calling its
	request and since the initial connection is not yet closed would try to open a second
	connection which would fail with "forward request denied".

	Thus we only unlock the mutex when the body of the response has been closed.
	This ensures the connection has been returned to the pool.
	**/
	resp.Body = &synchronousReadCloser{
		body:  resp.Body,
		mutex: &t.Mutex,
	}

	return resp, err
}

type synchronousReadCloser struct {
	body  io.ReadCloser
	mutex *sync.Mutex
}

func (s *synchronousReadCloser) Read(p []byte) (n int, err error) {
	return s.body.Read(p)
}

func (s *synchronousReadCloser) Close() error {
	defer s.mutex.Unlock()
	return s.body.Close()
}
