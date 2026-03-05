#!/bin/bash

echo "Starting SMS verification code service..."
echo "Service address: http://localhost:8083"
echo "Interface address: http://localhost:8083/oidc_auth/send/sms"
echo "Health check: http://localhost:8083/health"
echo "Press Ctrl+C to stop service"
echo "========================"

go run main.go