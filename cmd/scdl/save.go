// Package scdl /*
package scdl

import (
	"errors"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

var (
	path string
	// saveSongCmd represents the ssong command
	saveSongCmd = &cobra.Command{
		Use:   "ssong <song url>",
		Short: "Save song of specified url",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			resolvedPath, err := resolveFilepath(path)
			if err != nil {
				log.Fatal(err)
			}
			scdlSaveSong(url, resolvedPath)
		},
		// ToDo: autocompletion would be nice =)
		//BashCompletionFunction: func() {},
	}

	// savePlaylistCmd represents the splaylist command
	savePlaylistCmd = &cobra.Command{
		Use:   "splaylist <playlist url>",
		Short: "Save songs from playlist url",
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			resolvedPath, err := resolveFilepath(path)
			if err != nil {
				log.Fatal(err)
			}
			scdlSavePlaylist(url, resolvedPath)
		},
	}

	// saveLikesCmd represents the slikes command
	saveLikesCmd = &cobra.Command{
		Use:   "slikes",
		Short: "Save songs from user likes url",
		Run: func(cmd *cobra.Command, args []string) {
			url := args[0]
			resolvedPath, err := resolveFilepath(path)
			if err != nil {
				log.Fatal(err)
			}
			scdlSaveUserLikes(url, resolvedPath)
		},
	}
)

func init() {
	// This could be reworked to: scdl save [playlist, likes, song] <url> -f <path>
	saveSongCmd.Flags().StringVarP(&path, "filepath", "f", ".", "Save song to a custom filepath.")
	savePlaylistCmd.Flags().StringVarP(&path, "filepath", "f", ".", "Save songs from playlist to a custom filepath.")
	saveLikesCmd.Flags().StringVarP(&path, "filepath", "f", ".", "Save songs from user likes to a custom filepath.")
	rootCmd.AddCommand(saveLikesCmd)
	rootCmd.AddCommand(savePlaylistCmd)
	rootCmd.AddCommand(saveSongCmd)
}

func resolveFilepath(path string) (string, error) {
	switch true {
	case strings.HasPrefix(path, "./"):
		workDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = workDir + path[1:]

	case path == ".":
		workDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = workDir + string(filepath.Separator) + path[1:]

	case strings.HasPrefix(path, "~/"):
		usr, _ := user.Current()
		dir := usr.HomeDir
		path = filepath.Join(dir, path[2:])
	}
	if exists, err := dirPathExists(path); exists {
		return path, err
	}
	return "", errors.New("Cannot resolve path: \"" + path + "\" Does this path exist?")
}

func dirPathExists(path string) (bool, error) {
	file, err := os.Stat(path)
	if err == nil && file.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
