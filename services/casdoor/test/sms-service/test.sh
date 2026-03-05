#!/bin/bash

echo "Testing SMS verification code service..."
echo "========================"

# Check if service is running
echo "1. Checking service health status..."
curl -s http://localhost:8083/health | jq '.' 2>/dev/null || curl -s http://localhost:8083/health

echo -e "\n2. Testing JSON format SMS sending..."
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","code":"123456"}' | jq '.' 2>/dev/null || curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","code":"123456"}'

echo -e "\n3. Testing form format SMS sending..."
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -d "phone=13900139000&code=654321" | jq '.' 2>/dev/null || curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -d "phone=13900139000&code=654321"

echo -e "\n4. Testing error case (missing phone number)..."
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"code":"123456"}' | jq '.' 2>/dev/null || curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"code":"123456"}'

echo -e "\nTesting completed!"