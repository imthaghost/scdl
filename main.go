package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const tokenEnvVar = "SCDL_TOKEN"

var (
	flagToken  string
	flagOutput string
	flagProxy  string
)

var rootCmd = &cobra.Command{
	Use:   "scdl <url>",
	Short: "Download a song or playlist from a given SoundCloud URL",
	Long: `Download a song or playlist from SoundCloud and save it as .mp3
files with embedded cover art.

For Go+ high-quality downloads and private tracks, supply a SoundCloud OAuth
token via --token or the ` + tokenEnvVar + ` environment variable. Grab it from
your browser's DevTools: open the network tab on soundcloud.com, find any
request to api-v2.soundcloud.com, and copy the value after "OAuth " in the
Authorization request header.`,
	Args:          cobra.ArbitraryArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		token := flagToken
		if token == "" {
			token = os.Getenv(tokenEnvVar)
		}
		sc, err := NewClient(Options{
			Token:     token,
			OutputDir: flagOutput,
			ProxyURL:  flagProxy,
		})
		if err != nil {
			return err
		}
		return sc.Download(args[0])
	},
}

func main() {
	rootCmd.Flags().StringVar(&flagToken, "token", "", "SoundCloud OAuth token (overrides $"+tokenEnvVar+")")
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "output directory (default: current directory)")
	rootCmd.Flags().StringVar(&flagProxy, "proxy", "", "proxy URL, e.g. http://user:pass@host:port (defaults to $HTTPS_PROXY/$HTTP_PROXY)")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "scdl:", err)
		os.Exit(1)
	}
}
