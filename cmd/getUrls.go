package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"

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

		// Create a pipe for connecting the commands
		pipeReader, pipeWriter := io.Pipe()

		// Create a command to run cat
		catCmd := exec.Command("cat")
		catCmd.Stdin = bytes.NewReader(data)
		catCmd.Stdout = pipeWriter

		// Create a command to run waybackurls
		waybackCmd := exec.Command("waybackurls")
		waybackCmd.Stdin = pipeReader

		// Create a command to run gau
		gauCmd := exec.Command("gau")
		gauCmd.Stdin = pipeReader

		// Capture the output of gauCmd
		var output bytes.Buffer
		gauCmd.Stdout = &output

		// Create a WaitGroup to wait for all commands to finish
		var wg sync.WaitGroup
		wg.Add(2) // 2 commands: waybackCmd and gauCmd

		// Start the catCmd
		if err := catCmd.Start(); err != nil {
			return fmt.Errorf("error starting cat command: %w", err)
		}

		// Start the waybackCmd
		go func() {
			defer wg.Done()
			if err := waybackCmd.Start(); err != nil {
				fmt.Printf("error starting waybackurls command: %v\n", err)
				return
			}
			if err := waybackCmd.Wait(); err != nil {
				fmt.Printf("error waiting for waybackurls command: %v\n", err)
				return
			}
		}()

		// Start the gauCmd
		go func() {
			defer wg.Done()
			if err := gauCmd.Start(); err != nil {
				fmt.Printf("error starting gau command: %v\n", err)
				return
			}
			if err := gauCmd.Wait(); err != nil {
				fmt.Printf("error waiting for gau command: %v\n", err)
				return
			}
		}()

		// Wait for the catCmd to finish
		if err := catCmd.Wait(); err != nil {
			return fmt.Errorf("error waiting for cat command: %w", err)
		}

		// Close the pipeWriter to signal the end of input
		pipeWriter.Close()

		// Wait for all commands (waybackCmd and gauCmd) to finish
		wg.Wait()

		// Process the URLs from the output buffer
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
