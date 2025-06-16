package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"gopkg.in/yaml.v3"
)

const TFCURL = "https://app.terraform.io/app/"

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		return info.Main.Version
	}
	return ""
}

func main() {
	printFlag := flag.Bool("print", false, "Print the URL instead of opening it")
	printFlagShort := flag.Bool("p", false, "Print the URL instead of opening it (shorthand)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	versionFlagShort := flag.Bool("v", false, "Print version and exit (shorthand)")
	registryFlag := flag.Bool("registry", false, "Open the TFC private module registry")
	registryFlagShort := flag.Bool("r", false, "Open the TFC private module registry (shorthand)")
	flag.Parse()

	if *versionFlag || *versionFlagShort {
		fmt.Println(getVersion())
		return
	}

	printOnly := *printFlag || *printFlagShort
	showRegistry := *registryFlag || *registryFlagShort

	url, err := getUrl(showRegistry)

	if err != nil {
		log.Fatal(err)
	}

	openOrPrintURL(url, printOnly)
}

func findConfig() (*Config, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}

	for currentDir != "/" {
		// Try to find .tfcopen file
		configFilePath := filepath.Join(currentDir, ".tfcopen")
		if fileInfo, err := os.Stat(configFilePath); err == nil {
			if fileInfo.Size() == 0 {
				return nil, fmt.Errorf("config file found at %s but it is empty. please add configuration keys", configFilePath)
			}
			cfg, err := ReadConfig(configFilePath)
			if err != nil {
				return nil, fmt.Errorf("error reading config: %v", err)
			}
			if !hasKnownKeys(cfg) {
				// return the config anyway as it might have the org in which is used to print the registry url,
				// if that was what the user asked for
				return cfg, fmt.Errorf("config file found at %s but contains none of the expected keys (workspace, search, project). please check for typos", configFilePath)
			}
			return cfg, nil
		}

		// Check for git directory and use its name as search term if found
		if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
			fmt.Println("found git root, guessing the terraform cloud search string from its name")
			return &Config{Search: filepath.Base(currentDir)}, nil
		}

		currentDir = filepath.Dir(currentDir)
	}

	return nil, fmt.Errorf("reached / without finding a .tfcopen file. cannot continue")
}

// handleCommand and handleRegistryCommand have similar URL opening logic
func openOrPrintURL(url string, printOnly bool) {
	if printOnly {
		fmt.Println(url)
	} else {
		OpenURL(url)
	}
}

func getUrl(registry bool) (string, error) {
	cfg, err := findConfig()
	if err != nil && !registry {
		return "", err
	}

	org, err := resolveOrg(cfg)
	if err != nil {
		return "", err
	}

	var uri string
	if registry {
		uri = "/registry/private/modules"
	} else {
		uri = buildWorkspacesURI(cfg)
	}
	openURL := fmt.Sprintf("%s%s%s", TFCURL, org, uri)

	return openURL, nil
}

func hasKnownKeys(cfg *Config) bool {
	return cfg.Workspace != "" || cfg.Search != "" || cfg.Project != ""
}

func resolveOrg(cfg *Config) (string, error) {
	if cfg != nil && cfg.Org != "" {
		return cfg.Org, nil
	}

	if org := os.Getenv("TFCOPEN_DEFAULT_ORG"); org != "" {
		return org, nil
	}

	return "", fmt.Errorf("error: no org was found in any config file and the TFCOPEN_DEFAULT_ORG environment variable is not set. we cannot generate a link without knowing this")
}

func buildWorkspacesURI(cfg *Config) string {

	switch {
	case cfg.Workspace != "":
		return fmt.Sprintf("/workspaces/%s", cfg.Workspace)
	case cfg.Search != "":
		return fmt.Sprintf("/workspaces?search=%s", cfg.Search)
	case cfg.Project != "":
		return fmt.Sprintf("/projects/%s/workspaces", cfg.Project)
	default:
		return ""
	}
}

func OpenURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "windows": // Windows
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default: // Linux and other Unix-like systems
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

// Config holds the configuration values parsed from the .tfcopen file.
type Config struct {
	Workspace string `yaml:"workspace"`
	Search    string `yaml:"search"`
	Project   string `yaml:"project"`
	Org       string `yaml:"org"`
}

// ReadConfig reads the configuration from the specified file.
func ReadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &config, nil
}
