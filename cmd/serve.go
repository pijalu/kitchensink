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
	"net/http"

	"github.com/pijalu/kitchensink/quietlog"
	"github.com/spf13/cobra"
)

type serveConfig struct {
	bindAddr  *string
	servePath *string
	QuietFlag *bool
}

func (s *serveConfig) Quiet() bool {
	return s.QuietFlag != nil && *s.QuietFlag
}

var serveCfg serveConfig

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a static file http server",
	Long:  `This command start a basic http server`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log := quietlog.DefaultLogger(&serveCfg)
		_ = log
		log.Printf("Starting server %s for %s",
			*serveCfg.bindAddr,
			*serveCfg.servePath)

		fs := http.FileServer(http.Dir(*serveCfg.servePath))
		http.Handle("/", fs)
		if err := http.ListenAndServe(*serveCfg.bindAddr, nil); err != nil {
			log.Fatalf("Error during serve: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCfg = serveConfig{
		bindAddr:  serveCmd.Flags().StringP("listen", "l", "0.0.0.0:8080", "Bind address. Default: 0.0.0.0:8080"),
		servePath: serveCmd.Flags().StringP("path", "p", ".", "Serve path. Default: working dir"),
	}
}
