package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/pijalu/kitchensink/quietlog"
)

// Proxy represent a proxy
type Proxy struct {
	QuietFlag   *bool
	SourceAddr  *string
	TargetAddr  *string
	Protocol    *string
	DialTimeOut *time.Duration
	Log         *quietlog.QuietLogger
}

// Quiet returns true if the tool should keep beeing quiet
func (proxy *Proxy) Quiet() bool {
	return (proxy.QuietFlag != nil) && *proxy.QuietFlag
}

func (proxy *Proxy) log() *quietlog.QuietLogger {
	if proxy.Log == nil {
		proxy.Log = quietlog.DefaultLogger(proxy)
	}
	return proxy.Log
}

// Run proxy
func (proxy *Proxy) Run() {
	listener, err := net.Listen(*proxy.Protocol, *proxy.SourceAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer listener.Close()
	proxy.log().Printf("Listening on %s", *proxy.SourceAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		proxy.log().Printf("Go connection from %s", conn.RemoteAddr())
		go proxy.handle(conn)
	}
}

type proxyRequest struct {
	proxy  *Proxy
	ctx    context.Context
	cancel context.CancelFunc
}

func (r *proxyRequest) pipe(input io.Reader, output io.Writer) error {
	defer func() {
		r.cancel()
	}()

	for {
		select {
		case <-r.ctx.Done():
			break
		default:
			_, err := io.Copy(output, input)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

func (r *proxyRequest) copyConn(input io.Reader, output io.Writer) {
	if err := r.pipe(input, output); err != nil {
		select {
		case <-r.ctx.Done():
			// Don't print error - context is done so we closed stream so pending read/write may fail !
		default:
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
}

func (proxy *Proxy) handle(inputConn net.Conn) {
	proxy.log().Printf("Opening proxy to %s/%s for %s", *proxy.TargetAddr, *proxy.Protocol, inputConn.RemoteAddr())
	outputConn, err := net.DialTimeout(*proxy.Protocol, *proxy.TargetAddr, *proxy.DialTimeOut)
	if err != nil {
		proxy.log().Fatalf("Failed to dial %s/%s", *proxy.TargetAddr, *proxy.Protocol)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := proxyRequest{
		proxy:  proxy,
		ctx:    ctx,
		cancel: cancel,
	}

	// Close stream
	go func() {
		<-ctx.Done()
		inputConn.Close()
		outputConn.Close()

		proxy.log().Printf("Closing proxy to %s/%s for %s", *proxy.TargetAddr, *proxy.Protocol, inputConn.RemoteAddr())
	}()

	// Read proxy
	go r.copyConn(inputConn, outputConn)
	// Write proxy
	go r.copyConn(outputConn, inputConn)
}
