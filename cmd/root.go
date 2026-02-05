package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   `conduit`,
	Short: `Expose a local port to a public domain`,
	Long:  `A CLI tool to expose local services behind NAT over a multiplexed TCP tunnel`,
	Args:  cobra.ExactArgs(2),
}

var (
	localPort int
	domain    string
)

// TODO: Implement expose functionality
// 1. Connect to the conduit server
// 2. Create yamux session (multiplexed TCP tunnel)
// 3. Forward local port to the conduit server/ Tell service to listen on the public domain
// 4. Relay traffic/put a reverse proxy between local port and conduit server
var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Expose a local port through a public domain",
	Long:  "Expose a local port (e.g. 8080) through a public domain behind NAT => conduit expose -p <local-port> -d <public-domain>",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Exposing localhost:%d at %s\n", localPort, domain)
	},
}

func init() {
	exposeCmd.Flags().StringVarP(&domain, "domain", "d", "", "Public domain to expose service on (required)")
	exposeCmd.Flags().IntVarP(&localPort, "port", "p", 8080, "Local port to expose")
	exposeCmd.MarkFlagRequired("port")
	exposeCmd.MarkFlagRequired("domain")

	rootCmd.AddCommand(exposeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
