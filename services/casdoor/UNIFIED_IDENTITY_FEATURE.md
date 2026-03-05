# Casdoor Unified Identity Management System - New Feature Summary

## 🚀 Feature Overview

This update introduces a complete **Unified Identity Management System** to Casdoor, enabling user account merging, multi-authentication method binding, and a unified authentication mechanism. This is a brand new feature that allows users to log into the same account using different authentication methods (GitHub OAuth, phone number, email, custom OAuth, etc.).

## 📋 New Feature Checklist

### 🆕 New API Endpoints

#### 1. User Account Merge API
**POST `/api/identity/merge`**

- **Function**: Merges two user accounts into one, preserving one account and deleting the other
- **Authentication**: Requires two valid JWT Tokens
- **Request Body**:
```json
{
    "reserved_user_token": "eyJhbGciOiJSUzI1NiIs...",
    "deleted_user_token": "eyJhbGciOiJSUzI1NiIs..."
}
```
- **Response**:
```json
{
    "status": "ok",
    "universal_id": "90ea5f8b-38f8-452b-b4cf-1cd721a2ce27",
    "deleted_user_id": "550e8400-e29b-41d4-a716-446655440001",
    "merged_auth_methods": [
        {
            "auth_type": "phone",
            "auth_value": "+86138000000"
        },
        {
            "auth_type": "github",
            "auth_value": "123456789"
        }
    ]
}
```

#### 2. Identity Information Query API
**GET `/api/identity/info`**

- **Function**: Query all authentication methods bound to the current user
- **Authentication**: Bearer Token
- **Response**:
```json
{
    "universal_id": "90ea5f8b-38f8-452b-b4cf-1cd721a2ce27",
    "bound_auth_methods": [
        {
            "auth_type": "github",
            "auth_value": "123456789"
        },
        {
            "auth_type": "phone",
            "auth_value": "+86138000000"
        },
        {
            "auth_type": "email",
            "auth_value": "user@example.com"
        }
    ]
}
```

#### 3. Identity Binding Management API
**POST `/api/identity/bind`**

- **Function**: Bind a new authentication method to the current user
- **Authentication**: Bearer Token
- **Request Body**:
```json
{
    "auth_type": "email",
    "auth_value": "newuser@example.com"
}
```

**POST `/api/identity/unbind`**

- **Function**: Unbind a specified authentication method from the current user
- **Authentication**: Bearer Token
- **Request Body**:
```json
{
    "auth_type": "phone"
}
```

### 🗄️ Database Changes

#### 1. User Table Extension
```sql
-- Add universal_id field
ALTER TABLE user ADD COLUMN universal_id VARCHAR(100) INDEX;
```

#### 2. New User Identity Binding Table
```sql
CREATE TABLE user_identity_binding (
    id VARCHAR(100) PRIMARY KEY,
    universal_id VARCHAR(100) NOT NULL,
    auth_type VARCHAR(50) NOT NULL,
    auth_value VARCHAR(255) NOT NULL,
    created_time VARCHAR(100) NOT NULL,
    INDEX idx_universal_id (universal_id),
    INDEX idx_auth (auth_type, auth_value),
    UNIQUE KEY unique_auth (auth_type, auth_value)
);
```

**Field Description**:
- `universal_id`: Unified identity ID, linked to the UniversalId field in the User table
- `auth_type`: Authentication type (github, phone, email, password, custom, etc.)
- `auth_value`: Authentication value (GitHub ID, phone number, email address, etc.)

### 🔧 Core Feature Implementation

#### 1. JWT Token Enhancement
Added fields to JWT Token:
```json
{
    "universal_id": "90ea5f8b-38f8-452b-b4cf-1cd721a2ce27",
    "phone_number": "+86138000000",
    "github_account": "123456789",
    // ... other existing fields
}
```

#### 2. Unified Identity Login Mechanism
- **New Function**: `GetUserByFieldWithUnifiedIdentity()`
- **Function**: Prioritize finding users through identity binding table, fallback to traditional method if not found
- **Impact Scope**: All OAuth login flows (GitHub, Google, WeChat, custom, etc.)

#### 3. Enhanced User Creation Process
- **New Function**: `createIdentityBindings()`
- **Function**: Automatically establish corresponding identity binding records when users are created
- **Supported Authentication Types**:
  - `password`: Username and password
  - `phone`: Phone number
  - `email`: Email address
  - `github`: GitHub OAuth
  - `google`: Google OAuth
  - `wechat`: WeChat login
  - `custom`: Custom OAuth provider
  - And more...

#### 4. Complete User Merge Process
- **Authentication**: Validate JWT Tokens of both users
- **Data Transfer**: Transfer identity bindings of the deleted user to the preserved user
- **Data Cleanup**: Delete all related data of the deleted user:
  - User records
  - Token records
  - Session records
  - Verification records
  - Resource records
  - Payment records
  - Transaction records
  - Subscription records
- **Transaction Safety**: Use database transactions to ensure operation atomicity

### 🎯 Business Scenario Support

#### 1. Account Merge Scenario
```
User A: GitHub login (universal_id_A)
User B: Phone login (universal_id_B)
↓ User discovers duplicate accounts, requests merge
Call /api/identity/merge API
↓ Merge result
Preserve User A, delete User B
User A can now login with GitHub or phone number
```

#### 2. Multi-method Login Scenario
```
User registration: GitHub OAuth
Bind phone: Call /api/identity/bind
Bind email: Call /api/identity/bind
↓ User can now login to the same account via:
- GitHub OAuth
- Phone verification code
- Email verification code
```
