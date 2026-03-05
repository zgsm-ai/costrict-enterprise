#!/bin/bash

# GitHub OAuth Test Tool Startup Script

echo "🚀 Starting GitHub OAuth Test Tool"
echo ""

# Check environment variables
if [ -z "$GITHUB_CLIENT_ID" ]; then
    echo "❌ Please set GITHUB_CLIENT_ID environment variable"
    echo "   export GITHUB_CLIENT_ID=\"Your Client ID\""
    exit 1
fi

if [ -z "$GITHUB_CLIENT_SECRET" ]; then
    echo "❌ Please set GITHUB_CLIENT_SECRET environment variable"
    echo "   export GITHUB_CLIENT_SECRET=\"Your Client Secret\""
    exit 1
fi

echo "✅ Environment variable check passed"
echo "   GITHUB_CLIENT_ID: $GITHUB_CLIENT_ID"
echo "   GITHUB_CLIENT_SECRET: ${GITHUB_CLIENT_SECRET:0:4}****${GITHUB_CLIENT_SECRET: -4}"
echo ""

# Switch to project directory
cd "$(dirname "$0")"

# Initialize Go module dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Start service
echo ""
echo "🚀 Starting server..."
go run main.go