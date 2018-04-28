package waitconn

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pijalu/kitchensink/quietlog"
)

// Command stores all waitconn configs
type Command struct {
	// Protocol for the tool
	Protocol *string
	// Address to check
	addr *string
	// Number of tries to run, 0 for unlimited number of run
	Tries *int
	// Delay to wait between tries
	WaitDelay *time.Duration
	// Timeout for connection
	ConnectionTimeout *time.Duration
	// Be quiet
	QuietFlag *bool
}

// Quiet returns true if this flags have are quiet
func (w *Command) Quiet() bool {
	return *w.QuietFlag
}

// Addr set the address and return the command
func (w *Command) Addr(addr *string) *Command {
	w.addr = addr
	return w
}

// Run waitconn command
func (w *Command) Run() {
	logger := quietlog.DefaultLogger(w)
	logger.Printf("Checking for connection for %s/%s", *w.addr, *w.Protocol)

	for t := 0; ; t++ {
		conn, err := net.DialTimeout(*w.Protocol, *w.addr, *w.ConnectionTimeout)
		if err == nil {
			logger.Printf("Connection successful !")
			conn.Close()
			break
		}
		var prefix string
		if *w.Tries != 0 && !w.Quiet() {
			prefix = fmt.Sprintf("Try %d of %d ", t+1, *w.Tries)
		}

		logger.Printf("%s No reply...", prefix)

		if *w.Tries == 0 || t+1 < *w.Tries {
			logger.Printf("Waiting for %s", w.WaitDelay)
			time.Sleep(*w.WaitDelay)
		} else {
			logger.Printf("No reply after %d requests - stopping !", t+1)
			os.Exit(-1)
		}
	}
}
