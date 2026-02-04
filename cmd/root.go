package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "conduit",
	Short: "Expose a local port to a public domain",
	Long:  `A CLI tool to expose local services behind NAT over a multiplexed TCP tunnel`,
	Args:  cobra.ExactArgs(2),
}
