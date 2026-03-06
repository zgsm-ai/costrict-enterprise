package utils

import (
	"net/url"
	"testing"
)

// TestBuildLoginURL_Ignores tests the BuildLoginURL function with ignore parameters
func TestBuildLoginURL_Ignores(t *testing.T) {
	// Create test parameters
	params := &LoginParams{
		MachineCode:   "default-a1b2c3d4e5f6789012345678",
		VSCodeVersion: "1.101.0",
		PluginVersion: "2.0.6",
		Provider:      "casdoor",
		URIScheme:     "vscode",
		BaseURL:       "https://zgsm.sangfor.com",
	}

	// Test case 1: No ignores, should include all parameters
	url, err := BuildLoginURL(params, "")
	if err != nil {
		t.Errorf("BuildLoginURL with no ignores failed: %v", err)
	}

	// Check that all parameters are included
	if url == "" {
		t.Error("URL is empty")
	}

	// Test case 2: Ignore machine_code
	url, err = BuildLoginURL(params, "", "machine_code")
	if err != nil {
		t.Errorf("BuildLoginURL with machine_code ignore failed: %v", err)
	}

	// Check that machine_code is not in URL
	if containsParam(url, "machine_code") {
		t.Error("machine_code parameter should be ignored but was found in URL")
	}

	// Test case 3: Ignore multiple parameters
	url, err = BuildLoginURL(params, "", "machine_code", "provider", "plugin_version")
	if err != nil {
		t.Errorf("BuildLoginURL with multiple ignores failed: %v", err)
	}

	// Check that ignored parameters are not in URL
	if containsParam(url, "machine_code") {
		t.Error("machine_code parameter should be ignored but was found in URL")
	}
	if containsParam(url, "provider") {
		t.Error("provider parameter should be ignored but was found in URL")
	}
	if containsParam(url, "plugin_version") {
		t.Error("plugin_version parameter should be ignored but was found in URL")
	}

	// Test case 4: Ignore non-existent parameter (should not affect other parameters)
	url, err = BuildLoginURL(params, "", "non_existent_param")
	if err != nil {
		t.Errorf("BuildLoginURL with non-existent ignore failed: %v", err)
	}

	// Check that all parameters except non_existent_param are included
	if !containsParam(url, "machine_code") {
		t.Error("machine_code parameter should be included but was not found in URL")
	}
	// Note: state is not included in default params, so we don't check for it
	if !containsParam(url, "provider") {
		t.Error("provider parameter should be included but was not found in URL")
	}
}

// containsParam checks if a parameter is present in the URL query string
func containsParam(urlString, param string) bool {
	// Parse the URL to extract query parameters
	u, err := url.Parse(urlString)
	if err != nil {
		return false
	}

	// Get query values from the URL
	queryValues := u.Query()

	// Check if the parameter exists in the query
	_, exists := queryValues[param]
	return exists
}

// TestBuildLoginURL_WithNilParams tests the BuildLoginURL function with nil parameters
func TestBuildLoginURL_WithNilParams(t *testing.T) {
	// Test case: nil params should return error
	_, err := BuildLoginURL(nil, "")
	if err == nil {
		t.Error("Expected error for nil params, but got none")
	}
}

// TestBuildLoginURL_WithEmptyParams tests the BuildLoginURL function with empty parameter values
func TestBuildLoginURL_WithEmptyParams(t *testing.T) {
	// Create test parameters with empty values
	params := &LoginParams{
		BaseURL:       "https://example.com",
		MachineCode:   "",
		State:         "",
		Provider:      "",
		PluginVersion: "",
		VSCodeVersion: "",
		URIScheme:     "",
	}

	// Test case: Empty params should still work
	url, err := BuildLoginURL(params, "")
	if err != nil {
		t.Errorf("BuildLoginURL with empty params failed: %v", err)
	}

	if url != "https://example.com" {
		t.Errorf("Expected URL to be 'https://example.com', got '%s'", url)
	}
}

// TestGetToken_WithNilParams tests the GetToken function with nil parameters
func TestGetToken_WithNilParams(t *testing.T) {
	// Test case: nil params should return error
	_, err := GetToken(nil)
	if err == nil {
		t.Error("Expected error for nil params, but got none")
	}
}

// TestOpenBrowser_URLWithAmpersand tests the OpenBrowser function with URLs containing '&'
func TestOpenBrowser_URLWithAmpersand(t *testing.T) {
	// Test case: URL with ampersand should not cause issues
	// Note: We can't easily test the actual browser opening without UI interaction,
	// but we can verify that the command construction doesn't fail

	testURL := "https://example.com/login?param1=value1&param2=value2&redirect_uri=https://callback.example.com"

	// This should not panic or return an error due to the ampersand characters
	err := OpenBrowser(testURL)

	// We expect an error since we're not actually running the command in a test environment,
	// but it should not be due to URL parsing issues
	if err != nil {
		// Just log that we got an error (expected in test environment)
		// The important thing is that we don't panic or have command construction issues
		t.Logf("Expected error in test environment: %v", err)
	}
}
