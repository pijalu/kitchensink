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

	"github.com/pijalu/kitchensink/tool/proxy"
	"github.com/spf13/cobra"
)

var pxy proxy.Proxy

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy [bind.address]:port target:port",
	Short: "Start a proxy server to connect to a remote address",
	Long:  `This command will start a proxy server that will forward all packet to a given address/port. This can be used to create a reroute to a remote ip:port`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		pxy.SourceAddr = &args[0]
		pxy.TargetAddr = &args[1]

		pxy.Run()
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	pxy = proxy.Proxy{
		Protocol:    proxyCmd.Flags().StringP("protocol", "p", "tcp", "Protocol: tcp or udp."),
		DialTimeOut: proxyCmd.Flags().DurationP("timeout", "t", 30*time.Second, "Timeout for connect."),
		QuietFlag:   &quietFlag,
	}
}
