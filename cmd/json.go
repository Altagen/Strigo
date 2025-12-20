package cmd

import (
	"encoding/json"
	"fmt"
)

// Global variables for JSON mode
var (
	jsonOutput bool // Flag for JSON output
	jsonLogs   bool // Flag for JSON logs
)

// GetJsonOutput returns the value of the JSON flag
func GetJsonOutput() bool {
	return jsonOutput
}

// CommandOutput structure for JSON output
type CommandOutput struct {
	Types         []string `json:"types,omitempty"`
	Distributions []string `json:"distributions,omitempty"`
	Versions      []string `json:"versions,omitempty"`
	Error         string   `json:"error,omitempty"`
}

// OutputJSON handles JSON output for all commands
func OutputJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	return nil
}
