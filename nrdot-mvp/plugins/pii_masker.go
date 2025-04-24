// Build script to compile this to WASM:
// GOOS=wasip1 GOARCH=wasm go build -o plugins/pii_masker.wasm plugins/pii_masker.go

package main

import (
	"encoding/json"
	"regexp"
)

// Configuration for PII masking
type Config struct {
	MaskPassword      bool   `json:"mask_password"`
	MaskSSN           bool   `json:"mask_ssn"`
	MaskCreditCard    bool   `json:"mask_credit_card"`
	PasswordFields    string `json:"password_fields"`
	SSNRegex          string `json:"ssn_regex"`
	CreditCardRegex   string `json:"credit_card_regex"`
}

// Default configuration
var defaultConfig = Config{
	MaskPassword:    true,
	MaskSSN:         true,
	MaskCreditCard:  true,
	PasswordFields:  "password|passwd|pwd",
	SSNRegex:        "\\b\\d{3}-\\d{2}-\\d{4}\\b",
	CreditCardRegex: "\\b(?:\\d[ -]*?){13,16}\\b",
}

var config Config

// Compiled regexes
var (
	passwordFieldsRegex *regexp.Regexp
	ssnRegex           *regexp.Regexp
	creditCardRegex    *regexp.Regexp
)

// Initialize the WASM module
func init() {
	config = defaultConfig
	compileRegexes()
}

// compileRegexes compiles the regular expressions from the config
func compileRegexes() {
	passwordFieldsRegex = regexp.MustCompile(config.PasswordFields)
	ssnRegex = regexp.MustCompile(config.SSNRegex)
	creditCardRegex = regexp.MustCompile(config.CreditCardRegex)
}

// External functions imported from the host
//export log_utf8
func log_utf8(ptr, len uint32)

//export read_attr
func read_attr(name_ptr, name_len uint32, value_ptr, value_len uint32) uint32

//export write_attr
func write_attr(name_ptr, name_len uint32, value_ptr, value_len uint32) uint32

//export drop_record
func drop_record() uint32

// Main entry point for the WASM module
//export process_record
func process_record(config_ptr, config_len uint32) uint32 {
	// Parse configuration if provided
	if config_len > 0 {
		// In a real implementation, we would read the config from the host
		// For this example, we'll use the default config
		parseConfig([]byte(`{
			"mask_password": true,
			"mask_ssn": true,
			"mask_credit_card": true,
			"password_fields": "password|passwd|pwd",
			"ssn_regex": "\\\\b\\\\d{3}-\\\\d{2}-\\\\d{4}\\\\b",
			"credit_card_regex": "\\\\b(?:\\\\d[ -]*?){13,16}\\\\b"
		}`))
	}

	// Process attributes
	processAttributes()

	return 0 // Success
}

// parseConfig parses the JSON configuration
func parseConfig(configData []byte) {
	// Parse JSON
	var newConfig Config
	if err := json.Unmarshal(configData, &newConfig); err != nil {
		// Log error
		return
	}

	// Update config
	config = newConfig
	compileRegexes()
}

// processAttributes processes all attributes for PII
func processAttributes() {
	// In a real implementation, we would iterate through all attributes
	// For this example, we'll just demonstrate the concept

	// Example: Check for password fields
	if config.MaskPassword {
		// Mask password fields (implementation omitted)
	}

	// Example: Check for SSNs
	if config.MaskSSN {
		// Mask SSNs (implementation omitted)
	}

	// Example: Check for credit card numbers
	if config.MaskCreditCard {
		// Mask credit card numbers (implementation omitted)
	}
}

// maskValue replaces a value with asterisks, keeping the first and last characters
func maskValue(value string) string {
	if len(value) <= 2 {
		return "**"
	}

	masked := value[0:1]
	for i := 1; i < len(value)-1; i++ {
		masked += "*"
	}
	masked += value[len(value)-1:]

	return masked
}

// Required main function for Go WASM
func main() {}
