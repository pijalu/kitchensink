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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/kabukky/httpscerts"
	"github.com/pijalu/kitchensink/quietlog"
	"github.com/spf13/cobra"
)

type serveConfig struct {
	bindAddr  *string
	servePath *string
	useSSL    *bool
	QuietFlag *bool
}

func (s *serveConfig) Quiet() bool {
	return s.QuietFlag != nil && *s.QuietFlag
}

var serveCfg serveConfig

func secretDirectory() string {
	// Get current user
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s%c.kitchensink%c",
		user.HomeDir,
		os.PathSeparator,
		os.PathSeparator)
}

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

		// Setup TLS followin cloudfare advice (https://blog.cloudflare.com/exposing-go-on-the-internet/)
		tlsConfig := &tls.Config{
			// Causes servers to use Go's default ciphersuite preferences,
			// which are tuned to avoid attacks. Does nothing on clients.
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519, // Go 1.8 only
			},
		}

		// Setup server
		srv := &http.Server{
			Addr:         *serveCfg.bindAddr,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			TLSConfig:    tlsConfig,
		}

		fs := http.FileServer(http.Dir(*serveCfg.servePath))
		http.Handle("/", fs)

		var err error
		if *serveCfg.useSSL {
			secretDir := secretDirectory()
			cert := fmt.Sprintf("%s%s", secretDir, "serve-cert.pem")
			key := fmt.Sprintf("%s%s", secretDir, "serve-key.pem")
			if err := httpscerts.Check(cert, key); err != nil {
				if err := os.MkdirAll(secretDir, 0755); err != nil {
					log.Printf("Failed to create secret directory: %v", err)
					os.Exit(1)
				}
				log.Printf("Generating new certificate in %s", secretDir)
				if err := httpscerts.Generate(cert, key, *serveCfg.bindAddr); err != nil {
					log.Printf("Failed to generate certificate: %v", err)
					os.Exit(1)
				}
			}
			err = srv.ListenAndServeTLS(cert, key)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil {
			log.Fatalf("Error during serve: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCfg = serveConfig{
		bindAddr:  serveCmd.Flags().StringP("listen", "l", "0.0.0.0:8080", "Bind address. Default: 0.0.0.0:8080"),
		servePath: serveCmd.Flags().StringP("path", "p", ".", "Serve path. Default: working dir"),
		useSSL: serveCmd.Flags().BoolP("ssl", "s", false,
			fmt.Sprintf("Serve via ssl protocol. Default: false. This command will use %sserve-cert.pem and %sserve-key.pem. If not present, these files will be created",
				secretDirectory(),
				secretDirectory())),
	}
}
