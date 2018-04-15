package tunnel

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"

	"os/user"

	"github.com/pijalu/kitchensink/quietlog"
	"golang.org/x/crypto/ssh"
)

// Config represents configuration for SSH tunnel
type Config struct {
	QuietFlag *bool

	Protocol   *string
	SourceAddr *string
	SSHAddr    *string
	TargetAddr *string

	RemoteCmd *string

	Username *string
	KeyFile  *string
	Password *string

	DialTimeOut *time.Duration
	Log         *quietlog.QuietLogger
}

// Quiet returns true if the tool should keep beeing quiet
func (c *Config) Quiet() bool {
	return (c.QuietFlag != nil) && *c.QuietFlag
}

// Return a logger
func (c *Config) log() *quietlog.QuietLogger {
	if c.Log == nil {
		c.Log = quietlog.DefaultLogger(c)
	}
	return c.Log
}

type tunnelServer struct {
	c  *Config
	m  sync.Mutex
	wg sync.WaitGroup

	client *ssh.Client

	ctx    context.Context
	cancel context.CancelFunc
}

// loadKey load a private key and return signer
func (t *tunnelServer) loadKey(privateKey string) (ssh.Signer, error) {
	key, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

// clientConfig builds a client config
func (t *tunnelServer) clientConfig() *ssh.ClientConfig {
	config := ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Get current user
	user, err := user.Current()
	if err != nil {
		t.c.log().Fatalf("Failed to determine current user: %v", err)
		os.Exit(1)
	}

	// Username
	if *t.c.Username != "" {
		config.User = *t.c.Username
	} else {
		config.User = user.Username
	}

	// Password
	if *t.c.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(*t.c.Password))
	}

	// Key
	if *t.c.KeyFile != "" {
		signer, err := t.loadKey(*t.c.KeyFile)
		if err != nil {
			t.c.log().Fatalf("Failed to load key %s: %v", *t.c.KeyFile, err)
			os.Exit(1)
		} else {
			t.c.log().Fatalf("Loaded key %s")
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	} else { // load usual home key
		for _, keyName := range []string{"id_rsa", "id_dsa"} {
			keyFile := fmt.Sprintf("%s%c.ssh%c%s",
				user.HomeDir,
				os.PathSeparator,
				os.PathSeparator,
				keyName)
			signer, err := t.loadKey(keyFile)
			if err != nil {
				t.c.log().Printf("Could not load key %s: %v", keyFile, err)
			} else {
				config.Auth = append(config.Auth, ssh.PublicKeys(signer))
			}
		}
	}

	if len(config.Auth) < 1 {
		t.c.log().Fatalf("No authentiation method could be found !")
		os.Exit(1)
	}

	return &config
}

func (t *tunnelServer) connect() {
	t.wg.Add(1)

	t.m.Lock()
	defer t.m.Unlock()

	if t.client != nil {
		return
	}

	client, err := ssh.Dial("tcp", *t.c.SSHAddr, t.clientConfig())
	if err != nil {
		t.c.log().Fatalf("Failed to connect to %s: %v", *t.c.SSHAddr, err)
		os.Exit(1)
	}
	t.client = client

	// Start new root context
	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel

	// Start session
	session, err := client.NewSession()
	if err != nil {
		t.c.log().Fatalf("Failed to start session on %s: %v", *t.c.SSHAddr, err)
		os.Exit(1)
	}

	// Run session
	go func() {
		if t.c.Quiet() {
			session.Stdout = ioutil.Discard
			session.Stderr = ioutil.Discard
		} else {
			session.Stdout = os.Stdout
			session.Stderr = os.Stderr
		}

		if err := session.Run(*t.c.RemoteCmd); err != nil {
			select {
			case <-ctx.Done():
				/* ignore error as we are closing connection */
			default:
				t.c.log().Fatalf("Error running %s on  %s: %v",
					*t.c.RemoteCmd,
					*t.c.SSHAddr,
					err)
			}
		}
		t.c.log().Printf("Closing session on %s", *t.c.SSHAddr)
		t.cancel()
	}()

	// Shutdown connection if no clients
	go func() {
		t.wg.Wait()
		t.c.log().Printf("No more client, Sending close request for  %s", *t.c.SSHAddr)
		// No more client running - close
		t.cancel()
	}()

	// Close when context is done
	go func() {
		<-t.ctx.Done()
		t.m.Lock()
		defer t.m.Unlock()

		t.c.log().Printf("No more client, Closing session to %s", *t.c.SSHAddr)
		// Closing session and client
		session.Close()
		client.Close()
		// Reset
		t.client = nil
	}()
}

func (t *tunnelServer) handle(inputConn net.Conn) {
	// Connect as needed
	t.connect()

	outputConn, err := t.client.Dial(*t.c.Protocol, *t.c.TargetAddr)
	if err != nil {
		t.c.log().Fatalf("Failed to dial %s/%s", *t.c.TargetAddr, *t.c.Protocol)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(t.ctx)

	go func() {
		<-ctx.Done()
		t.wg.Done()

		inputConn.Close()
		outputConn.Close()

		t.c.log().Printf("Closing tunnel to %s/%s for %s",
			*t.c.TargetAddr,
			*t.c.Protocol,
			inputConn.RemoteAddr())
	}()

	copyFunc := func(r io.Reader, w io.Writer) {
		defer cancel()
		_, err := io.Copy(w, r)
		if err != nil {
			select {
			case <-ctx.Done():
				/* no issues - we are closing */
			default:
				t.c.log().Fatalf("Error during copy: %v", err)
			}
		}
	}

	go copyFunc(inputConn, outputConn)
	go copyFunc(outputConn, inputConn)
}

// Run tunnel
func (c *Config) Run() {
	t := tunnelServer{
		c: c,
	}

	listener, err := net.Listen(*t.c.Protocol, *t.c.SourceAddr)
	if err != nil {
		t.c.log().Fatalf("Error listening on %s/%s: %v",
			*t.c.SourceAddr,
			*t.c.Protocol,
			err)
		os.Exit(1)
	}
	defer listener.Close()
	t.c.log().Printf("Listening on %s", *t.c.SourceAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			t.c.log().Fatalf("Error during accept: %v",
				err)
			os.Exit(1)
		}
		t.c.log().Printf("Got connection from %s", conn.RemoteAddr())
		go t.handle(conn)
	}
}
