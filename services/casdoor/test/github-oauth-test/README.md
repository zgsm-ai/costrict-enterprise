# GitHub OAuth Callback Handler Service

This is a Go microservice specifically designed to handle GitHub OAuth 2.0 authorization callbacks.

## Features

- 🔗 Focused on GitHub OAuth callback handling
- 📱 Supports both GET and POST requests
- 🔐 Secure authorization code processing
- 📊 JSON formatted response data
- 🔍 Detailed logging output for debugging
- 🏥 Health check endpoint

## API Endpoints

### 1. Health Check
```
GET /health
```

Response:
```json
{
  "status": "ok",
  "timestamp": 1640995200,
  "service": "github-oauth-callback"
}
```

### 2. GitHub OAuth Callback Handler
```
GET|POST /auth/github/callback
```

Parameters:
- `code` (required) - GitHub authorization code
- `state` (optional) - State parameter

Success Response:
```json
{
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
}
```

Error Response:
```json
{
  "success": false,
  "message": "Processing failed",
  "error": "Error details"
}
```

## Prerequisites

### 1. Set Environment Variables

```bash
export GITHUB_CLIENT_ID="your Client ID"
export GITHUB_CLIENT_SECRET="your Client Secret"
```

Or on Windows:

```cmd
set GITHUB_CLIENT_ID=your Client ID
set GITHUB_CLIENT_SECRET=your Client Secret
```

### 2. Create GitHub OAuth App (Optional)

If you need to create a new GitHub OAuth application:

1. Visit [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in application information
4. Record the `Client ID` and `Client Secret`

## Running the Service

### Method 1: Direct Run

```bash
cd test/github-oauth-test
go mod tidy
go run main.go
```

### Method 2: Using Start Script

```bash
cd test/github-oauth-test
./start.sh
```

### Method 3: Build and Run

```bash
cd test/github-oauth-test
go build -o github-oauth-test
./github-oauth-test
```

## Testing Methods

### 1. Health Check

```bash
curl http://localhost:8080/health
```

### 2. Test Callback Endpoint

#### Using GET request
```bash
curl "http://localhost:8080/auth/github/callback?code=YOUR_GITHUB_AUTH_CODE"
```

#### Using POST request
```bash
curl -X POST "http://localhost:8080/auth/github/callback" \
     -d "code=YOUR_GITHUB_AUTH_CODE"
```

#### Using curl to send form data
```bash
curl -X POST "http://localhost:8080/auth/github/callback" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "code=YOUR_GITHUB_AUTH_CODE&state=optional_state"
```

## Debug Information

The service outputs detailed log information to the console, including:

- 📨 Received callback request information (request method, authorization code, etc.)
- ✅ Authorization code and access token acquisition process
- 👤 User information retrieval process
- ❌ Any error messages

## Retrieved User Information

- User ID
- Username (login)
- Display name (name)
- Email address (including private emails)
- Avatar URL
- Company information
- Location information
- Biography
- Blog link

## Security Notes

- Supports state parameter validation
- Authorization codes and access tokens are masked in logs
- Supports retrieving user's private email addresses

## Integration Example

### Calling from Other Applications

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
)

func handleGitHubCallback(code string) error {
    // Construct request data
    data := url.Values{}
    data.Set("code", code)

    // Send request to callback service
    resp, err := http.Post(
        "http://localhost:8080/auth/github/callback",
        "application/x-www-form-urlencoded",
        bytes.NewBufferString(data.Encode()),
    )
    if err != nil {
        return err
    }

    // Process response
    defer resp.Body.Close()
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return err
    }

    if success, ok := result["success"].(bool); ok && success {
        fmt.Println("GitHub OAuth successful")
        return nil
    }

    return fmt.Errorf("OAuth failed")
}
```

## Environment Configuration

Create an `.env` file (copy from `env.example`):

```bash
cp env.example .env
```

Edit the `.env` file with your GitHub OAuth credentials:

```
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_client_secret_here
```

## Error Handling

The service handles various error scenarios:

- Missing authorization code
- Invalid authorization code
- GitHub API failures
- Network connectivity issues
- Invalid client credentials

## Logging

All requests and responses are logged with appropriate log levels:

- `INFO`: Normal operation logs
- `ERROR`: Error conditions
- `DEBUG`: Detailed debugging information

## Development

### Project Structure

```
github-oauth-test/
├── main.go              # Main application
├── README.md           # This documentation
├── env.example         # Environment variable template
├── start.sh           # Startup script
└── go.mod             # Go module file
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## Troubleshooting

### Common Issues

1. **Invalid client credentials**: Verify your `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET`
2. **Authorization code expired**: GitHub authorization codes expire quickly, use them immediately
3. **Network issues**: Check your internet connection and GitHub API status

### Getting Help

If you encounter issues:

1. Check the console logs for detailed error messages
2. Verify your GitHub OAuth app configuration
3. Ensure environment variables are set correctly
4. Test with a fresh authorization code

## License

This project is part of the Casdoor authentication system and follows the same licensing terms.