package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (

	// Root cmd
	rootCmd = &cobra.Command{
		Use:                "scdl",
		Args:               cobra.NoArgs,
		DisableFlagParsing: true,
		Short:              "Download a song from given SoundCloud URL",
		Long:               `Download any given song from SoundCloud. Scdl is a utility that allows you to download a song from SoundCloud and get a .mp3 file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Usage(); err != nil {
				log.Fatal(err)
			}
		},
		Example: "scdl ssong https://soundcloud.com/darude/sandstorm-radio-edit -f ~/myMusic/\n" +
			"scdl search darude - sandstorm",
	}
)

// Execute the clone command
func Execute() {
	// Execute the command :)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
