/*
Copyright © 2025 Juan Orbegoso

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	version = "v0.0.10"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "checksum-utils",
	Version: version,
	Short:   "Multiplatform checksum utils.",
	Long: `A multiplatform checksum utils for NAS admins.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println()

		printResultsCheckingChecksumFiles(resultsCheckingChecksumFiles)
		printErrorsCheckingChecksumFiles()

		printResultsCreatingChecksumFiles(resultsCreatingChecksumFiles)
		printErrorsCreatingChecksumFiles()

		os.Exit(1)
	}()
}

func printHeader() {
	fmt.Println("Checksum-Utils", version)
	fmt.Println("https://github.com/JuanOrbegoso/checksum-utils")
}

func formatDuration(d time.Duration) string {
	rounded := d.Round(time.Millisecond)

	if rounded >= time.Hour {
		totalSeconds := int64(rounded.Round(time.Second).Seconds())
		hours := totalSeconds / 3600
		minutes := (totalSeconds % 3600) / 60
		seconds := totalSeconds % 60
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}

	if rounded >= time.Minute {
		totalSeconds := int64(rounded.Round(time.Second).Seconds())
		minutes := totalSeconds / 60
		seconds := totalSeconds % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}

	if rounded >= time.Second {
		seconds := int64(rounded.Round(time.Second).Seconds())
		return fmt.Sprintf("%ds", seconds)
	}

	return fmt.Sprintf("%dms", rounded.Milliseconds())
}
