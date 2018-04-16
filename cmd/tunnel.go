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

	"github.com/pijalu/kitchensink/tool/tunnel"
	"github.com/spf13/cobra"
)

var tunnelConfig tunnel.Config

// tunnelCmd represents the tunnel command
var tunnelCmd = &cobra.Command{
	Use:   "tunnel [bind.address]:port sshServer:sshPort remoteServer:remotePort",
	Short: "tunnel create a on-demand ssh tunnel to a given host/port  ",
	Long:  `tunnel command start a local server that will redirect all connection to a remote node via a ssh connection`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		tunnelConfig.SourceAddr = &args[0]
		tunnelConfig.SSHAddr = &args[1]
		tunnelConfig.TargetAddr = &args[2]

		tunnelConfig.Run()
	},
}

func init() {
	rootCmd.AddCommand(tunnelCmd)

	tunnelConfig = tunnel.Config{
		Protocol:    tunnelCmd.Flags().StringP("protocol", "p", "tcp", "Protocol: tcp or udp. Default tcp"),
		DialTimeOut: tunnelCmd.Flags().DurationP("timeout", "t", 30*time.Second, "Timeout for connect. Default 30sec"),
		QuietFlag:   &quietFlag,
		RemoteCmd:   tunnelCmd.Flags().StringP("cmd", "c", "vmstat 5", "Remote command to run on ssh host. Default to 'vmstat 5'"),
		Username:    tunnelCmd.Flags().StringP("user", "u", "", "Username to use for remote connection. Default to current username"),
		Password:    tunnelCmd.Flags().StringP("password", "w", "", "Password to use for authentication. Default: none"),
		KeyFile:     tunnelCmd.Flags().StringP("keyfile", "k", "", "Private key file to use."),
		Force:       tunnelCmd.Flags().BoolP("force", "f", false, "Keep trying to connect to ssh host even if down. Default: false"),
	}
}
