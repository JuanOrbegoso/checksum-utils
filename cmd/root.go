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
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	version = "v0.0.11"
)

const progressBarWidth = 10

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

type progressSpinner struct {
	done    chan struct{}
	wg      sync.WaitGroup
	enabled bool
}

func startProgress(prefix string) *progressSpinner {
	if !isStdoutTTY() {
		return &progressSpinner{enabled: false}
	}

	spinner := &progressSpinner{done: make(chan struct{})}
	spinner.enabled = true
	spinner.wg.Add(1)

	go func() {
		defer spinner.wg.Done()
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()

		position := 0
		for {
			select {
			case <-spinner.done:
				return
			default:
			}

			fmt.Printf("\r%s%s", prefix, buildProgressFrame(position))
			position = (position + 1) % (progressBarWidth + 1)

			select {
			case <-spinner.done:
				return
			case <-ticker.C:
			}
		}
	}()

	return spinner
}

func (s *progressSpinner) Stop() {
	if !s.enabled {
		return
	}
	close(s.done)
	s.wg.Wait()
}

func (s *progressSpinner) Enabled() bool {
	return s.enabled
}

func buildProgressFrame(position int) string {
	if position >= progressBarWidth {
		return progressDoneBar()
	}

	var builder strings.Builder
	builder.Grow(progressBarWidth + 2)
	builder.WriteByte('[')
	builder.WriteString(strings.Repeat("=", position))
	builder.WriteByte('>')
	builder.WriteString(strings.Repeat(" ", progressBarWidth-position-1))
	builder.WriteByte(']')
	return builder.String()
}

func progressDoneBar() string {
	return "[" + strings.Repeat("=", progressBarWidth) + "]"
}

func clearProgressLine(prefix string) {
	barLen := progressBarWidth + 2
	fmt.Printf("\r%s%s\r%s", prefix, strings.Repeat(" ", barLen), prefix)
}

func isStdoutTTY() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func isStdinTTY() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func readPathsFromStdin() ([]string, error) {
	if isStdinTTY() {
		return nil, nil
	}

	var paths []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, ".sha512") {
			continue
		}
		paths = append(paths, line)
	}
	if err := scanner.Err(); err != nil {
		return paths, err
	}

	return paths, nil
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

func expandArgs(args []string) ([]string, []error, bool) {
	var expanded []string
	var errs []error
	hadGlob := false

	for _, arg := range args {
		if hasGlobMeta(arg) {
			hadGlob = true
			matches, err := filepath.Glob(arg)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if len(matches) == 0 {
				errs = append(errs, fmt.Errorf("no matches for %q", arg))
				continue
			}
			expanded = append(expanded, matches...)
			continue
		}
		expanded = append(expanded, arg)
	}

	return expanded, errs, hadGlob
}

func gatherPaths(args []string) ([]string, []error, bool) {
	paths, errs, hadGlob := expandArgs(args)
	var readErr error

	if len(args) == 0 || !isStdinTTY() {
		stdinPaths, err := readPathsFromStdin()
		if err != nil {
			readErr = err
		}
		if len(stdinPaths) > 0 {
			paths = append(paths, stdinPaths...)
			hadGlob = true
		}
	}

	if readErr != nil {
		errs = append(errs, readErr)
	}

	return paths, errs, hadGlob
}

func hasGlobMeta(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

func processPaths(paths []string, errorsList *[]error, handler func(string) error) {
	for _, path := range paths {
		argFileInfo, err := os.Stat(path)
		if err != nil {
			*errorsList = append(*errorsList, err)
			continue
		}

		if argFileInfo.IsDir() {
			directoryAbsolutePath, err := filepath.Abs(path)
			if err != nil {
				*errorsList = append(*errorsList, err)
				continue
			}

			if err := filepath.Walk(directoryAbsolutePath, func(filePath string, fileInfo os.FileInfo, err error) error {
				if err != nil {
					*errorsList = append(*errorsList, err)
					fmt.Println("Error: ", err)
					return err
				}

				if fileInfo.IsDir() {
					return nil
				}

				return handler(filePath)
			}); err != nil {
				*errorsList = append(*errorsList, err)
				fmt.Println("Error: ", err)
			}
			continue
		}

		fileAbsolutePath, err := filepath.Abs(path)
		if err != nil {
			*errorsList = append(*errorsList, err)
			continue
		}

		ext := filepath.Ext(fileAbsolutePath)
		if ext == ".sha512" {
			isChecksumFileError := fmt.Errorf("%s is a checksum file.", fileAbsolutePath)
			*errorsList = append(*errorsList, isChecksumFileError)
			continue
		}

		if err := handler(path); err != nil {
			*errorsList = append(*errorsList, err)
		}
	}
}
