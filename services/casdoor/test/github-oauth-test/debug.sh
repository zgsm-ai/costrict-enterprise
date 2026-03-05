#!/bin/bash

echo "🔧 GitHub OAuth Diagnostic Script"
echo "============================="

# Check environment variables
echo "📋 Current Environment Variables:"
echo "GITHUB_CLIENT_ID: ${GITHUB_CLIENT_ID}"
if [ -n "${GITHUB_CLIENT_SECRET}" ]; then
    SECRET_LEN=${#GITHUB_CLIENT_SECRET}
    echo "GITHUB_CLIENT_SECRET: ${GITHUB_CLIENT_SECRET:0:4}...${GITHUB_CLIENT_SECRET: -4} (Length: $SECRET_LEN)"
else
    echo "GITHUB_CLIENT_SECRET: Not set"
fi
echo ""

# Check network connectivity
echo "🌐 Testing Network Connectivity:"
if curl -s --max-time 10 https://api.github.com/user > /dev/null; then
    echo "✅ GitHub API connection successful"
else
    echo "❌ GitHub API connection failed"
fi

if curl -s --max-time 10 https://github.com > /dev/null; then
    echo "✅ GitHub main site connection successful"
else
    echo "❌ GitHub main site connection failed"
fi
echo ""

# Run Go debug program
echo "🐛 Running Detailed Diagnostics:"
go run debug_test.go -c 'RunDebugTests()'

echo ""
echo "🔍 Additional Checks:"

# Check port usage
echo "1. Checking port usage:"
if command -v lsof > /dev/null; then
    echo "   Port 8000: $(lsof -ti:8000 | wc -l) processes"
    echo "   Port 8080: $(lsof -ti:8080 | wc -l) processes"
else
    echo "   lsof command not available, skipping port check"
fi

# Check system time
echo "2. System time: $(date)"

# Check DNS resolution
echo "3. DNS resolution test:"
if command -v dig > /dev/null; then
    DIG_RESULT=$(dig +short api.github.com)
    if [ -n "$DIG_RESULT" ]; then
        echo "   ✅ api.github.com resolution successful: $DIG_RESULT"
    else
        echo "   ❌ api.github.com resolution failed"
    fi
else
    echo "   dig command not available, skipping DNS check"
fi

echo ""
echo "📝 Solution Suggestions:"
echo "1. If all checks pass, the issue might be an expired authorization code"
echo "2. Try clearing browser cache and re-authorize"
echo "3. Ensure GitHub OAuth application status is active"
echo "4. If network issues exist, check firewall and proxy settings"