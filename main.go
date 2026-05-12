package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

const tokenEnvVar = "SCDL_TOKEN"

var tokenFlag string

var rootCmd = &cobra.Command{
	Use:   "scdl <url>",
	Short: "Download a song from a given SoundCloud URL",
	Long: `Download any song from SoundCloud and save it as a .mp3 with embedded cover art.

For Go+ high-quality downloads and private tracks, supply a SoundCloud OAuth
token via --token or the ` + tokenEnvVar + ` environment variable. Grab it from
your browser's DevTools: open the network tab on soundcloud.com, find any
request to api-v2.soundcloud.com, and copy the value after "OAuth " in the
Authorization request header.`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			if err := cmd.Usage(); err != nil {
				log.Fatal(err)
			}
			return
		}
		token := tokenFlag
		if token == "" {
			token = os.Getenv(tokenEnvVar)
		}
		NewClient(nil, token).Download(args[0])
	},
}

func main() {
	rootCmd.Flags().StringVar(&tokenFlag, "token", "", "SoundCloud OAuth token (overrides $"+tokenEnvVar+")")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
