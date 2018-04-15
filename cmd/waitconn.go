// Copyright © 2018 Pierre Poissinger <pierre.poissinger@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"time"

	"github.com/pijalu/kitchensink/tool/waitconn"
	"github.com/spf13/cobra"
)

// Local command object
var waitConn waitconn.Command

// waitconnCmd represents the waitconn command
var waitconnCmd = &cobra.Command{
	Use:   "waitconn address:port",
	Short: "Wait for a socket to be open",
	Long:  `waitcon will try to open a connection and retry every given delay (1sec by default)`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		waitConn.Addr(&args[0]).Run()
	},
}

func init() {
	rootCmd.AddCommand(waitconnCmd)
	waitConn = waitconn.Command{
		Protocol:          waitconnCmd.Flags().StringP("protocol", "p", "tcp", "Protocol: tcp or udp. Default tcp"),
		Tries:             waitconnCmd.Flags().IntP("tries", "n", 0, "Number of tries, 0 for no limits. Default 0"),
		WaitDelay:         waitconnCmd.Flags().DurationP("wait", "w", 1000*time.Millisecond, "Time to wait between tries. Default 1sec"),
		ConnectionTimeout: waitconnCmd.Flags().DurationP("timeout", "t", 30000*time.Millisecond, "Timeout for connection. Default 30sec"),
		QuietFlag:         &quietFlag,
	}
}
