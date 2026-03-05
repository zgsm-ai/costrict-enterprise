package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	githubOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
		RedirectURL:  "http://localhost:8080/auth/github/callback",
	}
)

// GitHubUser represents user information obtained from GitHub API
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	Bio       string `json:"bio"`
	Blog      string `json:"blog"`
}

// GitHubEmail represents email information obtained from GitHub API
type GitHubEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

// CallbackResponse callback response structure
type CallbackResponse struct {
	Success     bool        `json:"success"`
	Message     string      `json:"message"`
	User        *GitHubUser `json:"user,omitempty"`
	AccessToken string      `json:"access_token,omitempty"`
	Error       string      `json:"error,omitempty"`
}

func main() {
	// Check if debug mode is enabled
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		RunDebugTests()
		return
	}

	// Check environment variables
	if githubOauthConfig.ClientID == "" || githubOauthConfig.ClientSecret == "" {
		log.Fatal("Please set GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET environment variables")
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/auth/github/callback", handleGitHubCallback)
	http.HandleFunc("/callback", handleGitHubCallback)
	http.HandleFunc("/health", handleHealth)

	fmt.Println("🚀 GitHub OAuth callback handling service started at http://localhost:8080")
	fmt.Println("📝 Callback endpoints: POST/GET http://localhost:8080/auth/github/callback and /callback")
	fmt.Println("🧠 Smart redirect URL detection: Automatically tries multiple possible callback URLs")
	fmt.Println("⚙️  Environment variables:")
	fmt.Printf("   GITHUB_CLIENT_ID: %s\n", githubOauthConfig.ClientID)
	fmt.Printf("   GITHUB_CLIENT_SECRET: %s\n", maskSecret(githubOauthConfig.ClientSecret))
	fmt.Println("")
	fmt.Println("📋 API endpoints:")
	fmt.Println("   GET  /                           - Service status page")
	fmt.Println("   GET  /health                     - Health check")
	fmt.Println("   POST /callback                   - GitHub OAuth callback handler (Casdoor style)")
	fmt.Println("   GET  /callback                   - GitHub OAuth callback handler (Casdoor style)")
	fmt.Println("   POST /auth/github/callback       - GitHub OAuth callback handler (test service style)")
	fmt.Println("   GET  /auth/github/callback       - GitHub OAuth callback handler (test service style)")
	fmt.Println("")
	fmt.Println("💡 Tip: Make sure your GitHub OAuth app includes the following callback URLs:")
	fmt.Println("   - http://localhost:8000/callback")
	fmt.Println("   - http://localhost:8080/auth/github/callback")
	fmt.Println("   - http://127.0.0.1:8000/callback")
	fmt.Println("   - http://127.0.0.1:8080/auth/github/callback")
	fmt.Println("")
	fmt.Println("🐛 Debug mode: go run main.go --debug")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return strings.Repeat("*", len(secret))
	}
	return secret[:4] + strings.Repeat("*", len(secret)-8) + secret[len(secret)-4:]
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>GitHub OAuth Callback Service</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 50px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            max-width: 800px;
            margin: 0 auto;
        }
        .endpoint {
            background: #f6f8fa;
            padding: 15px;
            border-radius: 5px;
            margin: 10px 0;
            border-left: 4px solid #0366d6;
        }
        .method {
            background: #28a745;
            color: white;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 12px;
            margin-right: 10px;
        }
        .method.get { background: #17a2b8; }
        .method.post { background: #28a745; }
        h1 { color: #24292e; }
        code {
            background: #f6f8fa;
            padding: 2px 4px;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
        }
        pre {
            background: #f6f8fa;
            padding: 15px;
            border-radius: 6px;
            overflow-x: auto;
            border: 1px solid #e1e4e8;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔗 GitHub OAuth Callback Handling Service</h1>
        <p>This is a service specifically designed to handle GitHub OAuth 2.0 authorization callbacks.</p>

        <h2>📋 API Endpoints</h2>

        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/health</code>
            <p>Health check endpoint that returns service status.</p>
        </div>

        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="method post">POST</span>
            <code>/auth/github/callback</code>
            <p>GitHub OAuth callback handling endpoint.</p>
            <strong>Parameters:</strong>
            <ul>
                <li><code>code</code> - GitHub authorization code</li>
                <li><code>state</code> - State parameter (optional)</li>
            </ul>
        </div>

        <h2>📝 Response Format</h2>
        <p>Success response example:</p>
        <pre>{
  "success": true,
  "message": "GitHub OAuth processing successful",
  "user": {
    "id": 12345,
    "login": "username",
    "name": "User Name",
    "email": "user@example.com",
    "avatar_url": "https://avatars.githubusercontent.com/...",
    "company": "Company Name",
    "location": "Location",
    "bio": "User bio",
    "blog": "https://blog.example.com"
  },
  "access_token": "gho_xxxxxxxxxxxx"
}</pre>

        <p>Error response example:</p>
        <pre>{
  "success": false,
  "message": "Processing failed",
  "error": "Error details"
}</pre>

        <h2>🔧 Testing Method</h2>
        <p>You can test the callback endpoint using curl:</p>
        <pre>curl -X POST "http://localhost:8080/auth/github/callback" \
     -d "code=YOUR_GITHUB_AUTH_CODE"</pre>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "github-oauth-callback",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	// 支持 GET 和 POST 请求
	var code, state string

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			sendErrorResponse(w, "Failed to parse form data", err.Error())
			return
		}
		code = r.FormValue("code")
		state = r.FormValue("state")
	} else {
		code = r.URL.Query().Get("code")
		state = r.URL.Query().Get("state")
	}

	log.Printf("📨 Received GitHub callback request: method=%s, code=%s, state=%s", r.Method, maskCode(code), state)

	// Check authorization code
	if code == "" {
		log.Printf("❌ No authorization code received")
		sendErrorResponse(w, "No authorization code received", "Missing code parameter")
		return
	}

	log.Printf("✅ Authorization code received: %s", maskCode(code))

	// Smart detection of correct redirect URL
	// Try multiple possible redirect URLs until finding a valid one
	possibleRedirectURLs := []string{
		"http://localhost:8000/callback",             // Casdoor default
		"http://localhost:8080/auth/github/callback", // Test service default
		"http://127.0.0.1:8000/callback",             // Casdoor localhost variant
		"http://127.0.0.1:8080/auth/github/callback", // Test service localhost variant
	}

	// If there's an environment variable specified, use it first
	if customURL := os.Getenv("GITHUB_REDIRECT_URL"); customURL != "" {
		possibleRedirectURLs = append([]string{customURL}, possibleRedirectURLs...)
	}

	var token *oauth2.Token
	var err error
	var successRedirectURL string

	for _, redirectURL := range possibleRedirectURLs {
		log.Printf("🔍 Trying redirect URL: %s", redirectURL)

		config := *githubOauthConfig
		config.RedirectURL = redirectURL

		token, err = config.Exchange(context.Background(), code)
		if err == nil {
			successRedirectURL = redirectURL
			log.Printf("✅ Successfully used redirect URL: %s", redirectURL)
			break
		} else {
			log.Printf("❌ Redirect URL failed %s: %v", redirectURL, err)
		}
	}

	if err != nil {
		log.Printf("❌ All redirect URLs failed, last error: %v", err)
		sendErrorResponse(w, "Failed to get access token", fmt.Sprintf("All possible redirect URLs failed. Last error: %v", err))
		return
	}

	log.Printf("✅ Access token obtained successfully: %s (using redirect URL: %s)", maskToken(token.AccessToken), successRedirectURL)

	// Use access token to get user information
	userInfo, err := getUserInfo(token.AccessToken)
	if err != nil {
		log.Printf("❌ Failed to get user information: %v", err)
		sendErrorResponse(w, "Failed to get user information", err.Error())
		return
	}

	log.Printf("✅ User information retrieved successfully: %s (%s)", userInfo.Login, userInfo.Email)

	// Return success response
	response := CallbackResponse{
		Success:     true,
		Message:     "GitHub OAuth processing successful",
		User:        userInfo,
		AccessToken: token.AccessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, message, error string) {
	response := CallbackResponse{
		Success: false,
		Message: message,
		Error:   error,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func maskCode(code string) string {
	if len(code) <= 10 {
		return strings.Repeat("*", len(code))
	}
	return code[:5] + strings.Repeat("*", len(code)-10) + code[len(code)-5:]
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return strings.Repeat("*", len(token))
	}
	return token[:5] + strings.Repeat("*", len(token)-10) + token[len(token)-5:]
}

func getUserInfo(accessToken string) (*GitHubUser, error) {
	// Create HTTP client
	client := &http.Client{Timeout: 10 * time.Second}

	// Get user basic information
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status code %d): %s", resp.StatusCode, string(body))
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// If user's public email is empty, try to get private email
	if user.Email == "" {
		email, err := getUserEmail(client, accessToken)
		if err != nil {
			log.Printf("⚠️ Failed to get user email: %v", err)
		} else {
			user.Email = email
		}
	}

	return &user, nil
}

func getUserEmail(client *http.Client, accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error (status code %d): %s", resp.StatusCode, string(body))
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Prioritize primary email, then return verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email found")
}

// DebugConfig debug configuration information
func DebugConfig() {
	fmt.Println("🔧 === GitHub OAuth Debug Information ===")
	fmt.Printf("GITHUB_CLIENT_ID: %s\n", os.Getenv("GITHUB_CLIENT_ID"))
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if len(clientSecret) > 8 {
		fmt.Printf("GITHUB_CLIENT_SECRET: %s...%s (length: %d)\n",
			clientSecret[:4], clientSecret[len(clientSecret)-4:], len(clientSecret))
	} else {
		fmt.Printf("GITHUB_CLIENT_SECRET: %s (length: %d)\n", clientSecret, len(clientSecret))
	}
	fmt.Println()
}

// TestGitHubAPI test GitHub API connection
func TestGitHubAPI() error {
	fmt.Println("🌐 Testing GitHub API connection...")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return fmt.Errorf("unable to connect to GitHub API: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("✅ GitHub API response status: %d\n", resp.StatusCode)
	if resp.StatusCode == 401 {
		fmt.Println("✅ GitHub API connection normal (unauthorized response as expected)")
	}
	return nil
}

// TestTokenExchange test token exchange (using invalid code)
func TestTokenExchange() {
	fmt.Println("🔑 Testing OAuth configuration...")

	config := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
		RedirectURL:  "http://localhost:8000/callback",
	}

	// Use an obviously invalid code to test configuration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	_, err := config.Exchange(ctx, "invalid_test_code_12345")
	duration := time.Since(start)

	fmt.Printf("⏱️  Token exchange time: %v\n", duration)

	if err != nil {
		// Analyze error type
		if duration > 5*time.Second {
			fmt.Printf("⚠️  Response too slow (>5s), possible network issues\n")
		}

		errStr := err.Error()
		if strings.Contains(errStr, "invalid_grant") || strings.Contains(errStr, "bad_verification_code") {
			fmt.Println("✅ OAuth configuration correct (received expected invalid authorization code error)")
		} else if strings.Contains(errStr, "invalid_client") {
			fmt.Println("❌ Client ID or Secret error")
		} else if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
			fmt.Println("❌ Network timeout, check network connection")
		} else {
			fmt.Printf("❓ Unknown error: %v\n", err)
		}
	}
}

// ValidateEnvironment validate environment variables
func ValidateEnvironment() bool {
	fmt.Println("🔍 Validating environment variables...")

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	if clientID == "" {
		fmt.Println("❌ GITHUB_CLIENT_ID not set")
		return false
	}

	if clientSecret == "" {
		fmt.Println("❌ GITHUB_CLIENT_SECRET not set")
		return false
	}

	// Validate Client ID format (GitHub Client IDs usually start with Iv)
	if len(clientID) < 16 || !strings.Contains(clientID, "Iv") {
		fmt.Printf("⚠️  Client ID format may be incorrect: %s\n", clientID)
	} else {
		fmt.Println("✅ Client ID format correct")
	}

	// Validate Client Secret length
	if len(clientSecret) != 40 {
		fmt.Printf("⚠️  Client Secret length abnormal: %d (expected 40)\n", len(clientSecret))
	} else {
		fmt.Println("✅ Client Secret length correct")
	}

	return true
}

// RunDebugTests run all debug tests
func RunDebugTests() {
	fmt.Println("🐛 === GitHub OAuth Problem Diagnosis ===\n")

	DebugConfig()

	if !ValidateEnvironment() {
		fmt.Println("❌ Environment variable validation failed, please check configuration")
		return
	}

	fmt.Println()
	if err := TestGitHubAPI(); err != nil {
		fmt.Printf("❌ GitHub API test failed: %v\n", err)
	}

	fmt.Println()
	TestTokenExchange()

	fmt.Println("\n💡 Suggestions:")
	fmt.Println("1. If OAuth configuration is correct but still fails, please get a new authorization code")
	fmt.Println("2. Ensure authorization code is used immediately after obtaining (within 10 minutes)")
	fmt.Println("3. Check network connection and firewall settings")
	fmt.Println("4. Confirm GitHub OAuth application status is normal")
}
