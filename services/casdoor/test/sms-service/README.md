# SMS Verification Code Service

This is an HTTP service that provides SMS verification code functionality for Casdoor.

## Features

- Provides POST interface for sending SMS verification codes
- Supports JSON and form data formats
- Includes health check interface
- Detailed logging
- CORS support

## API Documentation

### 1. Send SMS Verification Code
- **URL**: `POST /oidc_auth/send/sms`
- **Port**: `8083`
- **Request Format**: JSON or form data

#### JSON Request Example:
```json
{
    "phone": "13800138000",
    "code": "123456"
}
```

#### Form Data Request Example:
```bash
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -d "phone=13800138000&code=123456"
```

#### Response Example:
```json
{
    "success": true,
    "message": "Verification code has been successfully sent to phone number 13800138000",
    "data": {
        "phone": "13800138000",
        "code": "123456",
        "timestamp": 1672531200
    }
}
```

### 2. Health Check
- **URL**: `GET /health`
- **Response Example**:
```json
{
    "status": "ok",
    "time": "2024-01-01 12:00:00",
    "service": "SMS Verification Code Service"
}
```

## Start Service

```bash
cd test/sms-service
go run main.go
```

After starting the service, it will display in the console:
```
SMS verification code service started, listening on port: :8083
SMS sending interface: POST http://localhost:8083/oidc_auth/send/sms
Health check interface: GET http://localhost:8083/health
```

## Configuration in Casdoor

When creating an SMS provider in the Casdoor management interface, configure as follows:

- **Type**: Custom HTTP SMS
- **Endpoint**: `http://localhost:8083/oidc_auth/send/sms`
- **Method**: POST
- **Parameters**: code (optional phone parameter configuration)

## Testing

### Testing with curl:
```bash
# JSON format
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","code":"123456"}'

# Form format
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -d "phone=13800138000&code=123456"

# Health check
curl http://localhost:8083/health
```

## Notes

1. This is a mock service. For actual SMS sending, you need to integrate with real SMS service provider APIs (such as Alibaba Cloud, Tencent Cloud, etc.)
2. The current version only logs and does not actually send SMS
3. The service supports CORS and can be called directly by the frontend
4. It is recommended to add more security verification and error handling in the production environment