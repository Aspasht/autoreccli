package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"bytes"
	"io/ioutil"
	"regexp"
	"github.com/spf13/cobra"
)

var filename string
var subdomainfile string

// collectJSCmd represents the collectJS command
var collectJSCmd = &cobra.Command{
	Use:   "collectJS",
	Short: "A brief description of your command",
	Long:  `A longer description that spans multiple lines.`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if a file path argument is provided
		if filename == "" {
			return fmt.Errorf("please provide a subdomain file path argument")
		}

		// Validate the provided file path
		err := ValidateFile(filename)
		if err != nil {
			return err
		}

		// Validate the subdomain file path
		err = ValidateFile(subdomainfile)
		if err != nil {
			return err
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {

		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading file: %s", err)
		}
		// Define the regex pattern to match URLs
		urlRegex := regexp.MustCompile(`https?://[^\s?"]+\.js(?:\?[^\s"]*)?`)
		// Find all matches in the content
		matches := urlRegex.FindAllString(string(content), -1)

		// Print the matched URLs
		for _, match := range matches {
			fmt.Println(match)
		}


		// Using getJS tool to fetch js files
		data, err := os.ReadFile(subdomainfile)
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
		getjsCmd := exec.Command("getJS", "--complete")
		getjsCmd.Stdin = pipeReader
		getjsCmd.Stdout = os.Stdout

		// Start the getJS command asynchronously
		if err := getjsCmd.Start(); err != nil {
			return fmt.Errorf("error starting getJS command: %w", err)
		}

		// Start the cat command
		if err := catCmd.Start(); err != nil {
			return fmt.Errorf("error starting cat command: %w", err)
		}

		// Wait for the catCmd to finish
		if err := catCmd.Wait(); err != nil {
			return fmt.Errorf("error waiting for cat command: %w", err)
		}

		// Close the pipeWriter to signal the end of input
		pipeWriter.Close()

		// Wait for the getjsCmd to finish
		if err := getjsCmd.Wait(); err != nil {
			return fmt.Errorf("error waiting for getJS command: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(collectJSCmd)
	collectJSCmd.Flags().StringVarP(&filename, "file", "f", "", "Path to the file containing URLs.")
	collectJSCmd.Flags().StringVarP(&subdomainfile, "domains", "d", "", "Path to the subdomain file")
}
