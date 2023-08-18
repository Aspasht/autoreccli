package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// getUrlsCmd represents the getUrls command
var getUrlsCmd = &cobra.Command{
	Use:   "getUrls",
	Short: "Fetch all the URLs using waybackurls and gau",
	Long:  "Fetch all the URLs using waybackurls and gau",

	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if a file path argument is provided
		if filePath == "" {
			return fmt.Errorf("please provide a subdomain file path argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read from the list of subdomains
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}

		// Create a command to run cat
		catCmd := exec.Command("cat")

		// Connect the input of catCmd to the subdomain data
		catCmd.Stdin = bytes.NewReader(data)

		// Create a command to run waybackurls
		waybackCmd := exec.Command("waybackurls")

		// Create a command to run gau
		gauCmd := exec.Command("gau")

		// Create a pipe for connecting the output of catCmd to the inputs of waybackCmd and gauCmd
		pipeReader, pipeWriter := io.Pipe()
		catCmd.Stdout = pipeWriter
		waybackCmd.Stdin = pipeReader
		gauCmd.Stdin = pipeReader

		// Capture the output of waybackCmd and gauCmd
		var output bytes.Buffer
		waybackCmd.Stdout = &output
		gauCmd.Stdout = &output

		// Start the catCmd
		if err := catCmd.Start(); err != nil {
			return fmt.Errorf("error starting cat command: %w", err)
		}

		// Start the waybackCmd
		if err := waybackCmd.Start(); err != nil {
			return fmt.Errorf("error starting waybackurls command: %w", err)
		}

		// Start the gauCmd
		if err := gauCmd.Start(); err != nil {
			return fmt.Errorf("error starting gau command: %w", err)
		}

		// Wait for the catCmd to finish
		if err := catCmd.Wait(); err != nil {
			return fmt.Errorf("error waiting for cat command: %w", err)
		}

		// Close the pipeWriter to signal the end of input
		pipeWriter.Close()

		// Wait for the waybackCmd to finish
		if err := waybackCmd.Wait(); err != nil {
			return fmt.Errorf("error waiting for waybackurls command: %w", err)
		}

		// Wait for the gauCmd to finish
		if err := gauCmd.Wait(); err != nil {
			return fmt.Errorf("error waiting for gau command: %w", err)
		}

		// Remove duplicate lines from the output and sort the URLs
		uniqueURLs := make(map[string]struct{})
		var sortedURLs []string
		for _, line := range strings.Split(output.String(), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				uniqueURLs[line] = struct{}{}
			}
		}
		for url := range uniqueURLs {
			sortedURLs = append(sortedURLs, url)
		}
		sort.Strings(sortedURLs)

		// Print the sorted and unique URLs
		for _, url := range sortedURLs {
			fmt.Println(url)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getUrlsCmd)
	getUrlsCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the subdomain file")
	getUrlsCmd.MarkFlagRequired("file")
}