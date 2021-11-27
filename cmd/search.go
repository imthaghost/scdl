package cmd

import (
	"github.com/imthaghost/scdl/pkg/soundcloud"
	"github.com/spf13/cobra"
	"strings"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search <query string>",
	Short: "Command for searching for songs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		soundcloud.Search(strings.Join(args[:], ""))
	},
	Example: "scdl search darude - sandstorm",
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
