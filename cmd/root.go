package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Find is --search flag
	Find bool

	// Root cmd
	rootCmd = &cobra.Command{
		Use:   "scdl <url>",
		Short: "Download a song from given SoundCloud URL",
		Long:  `Download any given song from SoundCloud. Scdl is a utility that allows you to download a song from SoundCloud and get a .mp3 file.`,
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// Print the usage if no args are passed in :)
			if len(args) < 1 {
				if err := cmd.Usage(); err != nil {
					log.Fatal(err)
				}

				return
			}

			// main function
			scdl(args)
		},
	}
)

// Execute the clone command
func Execute() {

	// Persistent Flags
	rootCmd.PersistentFlags().BoolVarP(&Find, "search", "s", false, "Option for searching for songs")

	// Execute the command :)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
