package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	

	"github.com/spf13/cobra"
)

var filePath string
var outputFile string

// getSubdomainsCmd represents the getSubdomains command
var getSubdomainsCmd = &cobra.Command{
	Use:   "getSubdomains",
	Short: "Gathers subdomains related to a domain.",
	Long:  `This command uses different tools like subfinder, amass, assetfinder, etc. to gather subdomains.`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if a file path argument is provided
		if filePath == "" {
			return fmt.Errorf("please provide a subdomain file path argument")
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve the domain flag value
		domain, err := cmd.Flags().GetString("domain")
		if err != nil {
			return fmt.Errorf("failed to retrieve domain flag: %w", err)
		}

		// Validate domain name
		if domain == "" {
			return fmt.Errorf("please provide a valid domain")
		}

		// Use "go install -v github.com/tomnomnom/assetfinder@latest" command to install assetfinder.
		assetfinderCmd := exec.Command("assetfinder", "--subs-only", domain)
		assetfinderOutput, err := assetfinderCmd.Output()
		if err != nil {
			return fmt.Errorf("assetfinder error: %v", err)
		}
		assetfinderSubdomains := strings.TrimSpace(string(assetfinderOutput))
		fmt.Println(assetfinderSubdomains)

		// Use "go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest" to install subfinder.
		subfinderCmd := exec.Command("subfinder", "-d", domain)
		subfinderOutput, err := subfinderCmd.Output()
		if err != nil {
			return fmt.Errorf("subfinder error: %v", err)
		}
		fmt.Println("subfinder output:")
		subfinderSubdomains := strings.TrimSpace(string(subfinderOutput))
		fmt.Println(subfinderSubdomains)

		// Use "go install -v github.com/owasp-amass/amass/v3/...@master" to install amass.
		amassCmd := exec.Command("amass", "enum", "-passive", "-d", domain)
		amassOutput, err := amassCmd.Output()
		if err != nil {
			return fmt.Errorf("amass error: %v", err)
		}

		amassSubdomains := strings.TrimSpace(string(amassOutput))
		fmt.Println(amassSubdomains)

		// Run Gobuster command
		gobusterArgs := []string{"dns", "-d", domain, "-w", filePath}
		gobusterCmd := exec.Command("gobuster", gobusterArgs...)

		// Run Grep command to filter subdomains
		grepCmd := exec.Command("grep", "-oP", `(?<=Found: )\S+`)

		// Capture the output of grepCmd
		var grepOutput bytes.Buffer
		grepCmd.Stdout = &grepOutput

		// Pipe the output of gobusterCmd to grepCmd
		grepCmd.Stdin, _ = gobusterCmd.StdoutPipe()

		// Start gobusterCmd and grepCmd
		err = gobusterCmd.Start()
		if err != nil {
			return fmt.Errorf("error starting gobuster command: %w", err)
		}
		err = grepCmd.Start()
		if err != nil {
			return fmt.Errorf("error starting grep command: %w", err)
		}

		// Wait for gobusterCmd and grepCmd to finish
		err = gobusterCmd.Wait()
		if err != nil {
			return fmt.Errorf("error waiting for gobuster command: %w", err)
		}
		err = grepCmd.Wait()
		if err != nil {
			return fmt.Errorf("error waiting for grep command: %w", err)
		}

		// Process the output of grepCmd and save subdomains to a file
		gobusterOutput := strings.TrimSpace(grepOutput.String())

		// Combine all subdomains and remove duplicates
		uniqueSubdomains := make(map[string]bool)

		// Add subdomains from assetfinder
		assetfinderSubdomainsList := strings.Split(assetfinderSubdomains, "\n")
		for _, subdomain := range assetfinderSubdomainsList {
			if subdomain != "" {
				uniqueSubdomains[subdomain] = true
			}
		}

		// Add subdomains from subfinder
		subfinderSubdomainsList := strings.Split(subfinderSubdomains, "\n")
		for _, subdomain := range subfinderSubdomainsList {
			if subdomain != "" {
				uniqueSubdomains[subdomain] = true
			}
		}

		// Add subdomains from amass
		amassSubdomainsList := strings.Split(amassSubdomains, "\n")
		for _, subdomain := range amassSubdomainsList {
			if subdomain != "" {
				uniqueSubdomains[subdomain] = true
			}
		}

		// Add subdomains from gobuster
		gobusterSubdomainsList := strings.Split(gobusterOutput, "\n")
		for _, subdomain := range gobusterSubdomainsList {
			if subdomain != "" {
				uniqueSubdomains[subdomain] = true
			}
		}

		finalSubdomains := make([]string, 0, len(uniqueSubdomains))
		for subdomain := range uniqueSubdomains {
			finalSubdomains = append(finalSubdomains, subdomain)
		}

		// Print subdomains to stdout
		fmt.Println("Subdomains:")
		for _, subdomain := range finalSubdomains {
			fmt.Println(subdomain)
		}

		// Save subdomains to a file
		if outputFile != "" {
			subdomainsString := strings.Join(finalSubdomains, "\n")
			err = ioutil.WriteFile(outputFile, []byte(subdomainsString), 0644)
			if err != nil {
				return fmt.Errorf("error writing output file: %w", err)
			}
			fmt.Printf("Subdomains saved to %s\n", outputFile)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getSubdomainsCmd)
	getSubdomainsCmd.Flags().StringP("domain", "d", "", "Domain name")
	getSubdomainsCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the subdomain wordlist file")
	getSubdomainsCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Path to the output file")
	getSubdomainsCmd.MarkFlagRequired("domain")
	getSubdomainsCmd.MarkFlagRequired("file")
}
