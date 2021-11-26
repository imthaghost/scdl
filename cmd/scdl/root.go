package scdl

import (
	"github.com/spf13/cobra"
	"log"
)

var (
	// Find is --search flag
	Find bool

	// Root cmd
	rootCmd = &cobra.Command{
		Use:   "scdl <url>",
		Short: "Download a song from given SoundCloud URL",
		Long:  `Download any given song from SoundCloud. Scdl is a utility that allows you to download a song from SoundCloud and get a .mp3 file.`,
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Usage(); err != nil {
				log.Fatal(err)
			}

			// main function
			scdl(args)
		},
		Example: "scdl https://soundcloud.com/darude/sandstorm-radio-edit\n" +
			"scdl https://soundcloud.com/darude/sandstorm-radio-edit --path ~/myMusic/",
	}
)

// Execute the clone command
func Execute() {

	// Persistent Flags
	rootCmd.PersistentFlags().BoolVarP(&Find, "search", "s", false, "Option for searching for songs")
	rootCmd.Flags().StringP("path", "p", "", "Option for saving song to custom path")
	// Execute the command :)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
