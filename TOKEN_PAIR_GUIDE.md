# Token Pair Generation Guide

This guide shows how to generate and use token pairs (access token + refresh token) with the strongly-typed security API.

## Overview

Token pairs provide enhanced security by separating short-lived access tokens from long-lived refresh tokens:

- **Access Token**: Short-lived (1 hour), used for API requests
- **Refresh Token**: Long-lived (7-30 days), used to generate new access tokens

## 🎯 **How to Generate Token Pairs**

### 1. Basic Token Pair Generation

```go
// Enable refresh token in your TokenRequest
tokenRequest := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:   userID,
        Username: username,
        Email:    email,
        ClientID: "web",
    },
    Metadata: security.TokenMetadata{
        Purpose:     "login",
        Source:      "web",
        Environment: "production",
        Scopes:      []string{"read", "write"},
        CreatedAt:   time.Now(),
    },
    RefreshToken: true, // 🔑 This enables refresh token generation
}

response, err := securityService.GenerateToken(tokenRequest)
if err != nil {
    return err
}

// Response will include both tokens
fmt.Printf("Access Token: %s\n", response.AccessToken)
fmt.Printf("Refresh Token: %s\n", response.RefreshToken) // Will be populated
fmt.Printf("Expires In: %d seconds\n", response.ExpiresIn)
```

### 2. Different Client Types

#### Web Application
```go
tokenRequest := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:   userID,
        Username: username,
        Email:    email,
        ClientID: "web",
    },
    Metadata: security.TokenMetadata{
        Purpose:     "login",
        Source:      "web",
        Environment: "production",
        Scopes:      []string{"read", "write"},
        CreatedAt:   time.Now(),
    },
    RefreshToken: true, // Web apps typically use refresh tokens
}
```

#### Mobile Application
```go
tokenRequest := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:   userID,
        Username: username,
        Email:    email,
        ClientID: "mobile",
        DeviceID: deviceID,
    },
    Metadata: security.TokenMetadata{
        Purpose:     "login",
        Source:      "mobile",
        Environment: "production",
        Scopes:      []string{"read", "write", "offline"},
        CreatedAt:   time.Now(),
    },
    RefreshToken: true, // Mobile apps definitely need refresh tokens
}
```

#### API-only Access (No Refresh Token)
```go
tokenRequest := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:   userID,
        ClientID: "api",
    },
    Metadata: security.TokenMetadata{
        Purpose:     "api",
        Source:      "api",
        Environment: "production",
        Scopes:      []string{"api:read", "api:write"},
        CreatedAt:   time.Now(),
    },
    RefreshToken: false, // API tokens typically don't need refresh
}
```

## 🔄 **How to Use Refresh Tokens**

### 1. Refresh Access Token

```go
// When access token expires, use refresh token to get a new one
newTokenResponse, err := securityService.RefreshAccessToken(refreshToken)
if err != nil {
    // Refresh token is invalid or expired - user needs to login again
    return redirectToLogin()
}

// Use the new access token
accessToken := newTokenResponse.AccessToken
expiresAt := newTokenResponse.ExpiresAt
```

### 2. HTTP Endpoint Example

```go
func (h *Handler) RefreshToken(c echo.Context) error {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(400, "Invalid request body")
    }
    
    if req.RefreshToken == "" {
        return echo.NewHTTPError(400, "Refresh token is required")
    }
    
    // Use security service to refresh the access token
    tokenResponse, err := h.security.RefreshAccessToken(req.RefreshToken)
    if err != nil {
        return echo.NewHTTPError(401, "Invalid or expired refresh token")
    }
    
    return c.JSON(200, map[string]interface{}{
        "access_token": tokenResponse.AccessToken,
        "token_type":   tokenResponse.TokenType,
        "expires_in":   tokenResponse.ExpiresIn,
        "expires_at":   tokenResponse.ExpiresAt,
        "message":      "Token refreshed successfully",
    })
}
```

## 📱 **Client-Side Usage Patterns**

### JavaScript/Frontend Example

```javascript
class TokenManager {
    constructor() {
        this.accessToken = localStorage.getItem('access_token');
        this.refreshToken = localStorage.getItem('refresh_token');
    }
    
    async login(username, password) {
        const response = await fetch('/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        // Store both tokens
        this.accessToken = data.access_token;
        this.refreshToken = data.refresh_token;
        
        localStorage.setItem('access_token', this.accessToken);
        localStorage.setItem('refresh_token', this.refreshToken);
        
        return data;
    }
    
    async refreshAccessToken() {
        if (!this.refreshToken) {
            throw new Error('No refresh token available');
        }
        
        const response = await fetch('/refresh', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: this.refreshToken })
        });
        
        if (!response.ok) {
            // Refresh token expired - redirect to login
            this.logout();
            window.location.href = '/login';
            return;
        }
        
        const data = await response.json();
        
        // Update access token
        this.accessToken = data.access_token;
        localStorage.setItem('access_token', this.accessToken);
        
        return data;
    }
    
    async apiCall(url, options = {}) {
        // Add access token to request
        const headers = {
            'Authorization': `Bearer ${this.accessToken}`,
            ...options.headers
        };
        
        let response = await fetch(url, { ...options, headers });
        
        // If access token expired, try to refresh
        if (response.status === 401) {
            await this.refreshAccessToken();
            
            // Retry with new access token
            headers['Authorization'] = `Bearer ${this.accessToken}`;
            response = await fetch(url, { ...options, headers });
        }
        
        return response;
    }
    
    logout() {
        this.accessToken = null;
        this.refreshToken = null;
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
    }
}
```

### Mobile App Example (React Native/Flutter concept)

```javascript
class MobileTokenManager {
    async login(username, password) {
        const response = await fetch('/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        // Store tokens securely (use Keychain/Keystore in real apps)
        await SecureStore.setItemAsync('access_token', data.access_token);
        await SecureStore.setItemAsync('refresh_token', data.refresh_token);
        
        return data;
    }
    
    async getValidAccessToken() {
        let accessToken = await SecureStore.getItemAsync('access_token');
        
        // Check if token is expired (decode JWT and check exp claim)
        if (this.isTokenExpired(accessToken)) {
            // Try to refresh
            const refreshToken = await SecureStore.getItemAsync('refresh_token');
            
            if (refreshToken) {
                const newTokens = await this.refreshAccessToken(refreshToken);
                accessToken = newTokens.access_token;
            } else {
                // No refresh token - need to login
                throw new Error('Authentication required');
            }
        }
        
        return accessToken;
    }
}
```

## 🔒 **Security Best Practices**

### 1. Token Storage
- **Access Tokens**: Can be stored in memory or short-term storage
- **Refresh Tokens**: Store securely (HttpOnly cookies, secure storage)
- **Never store tokens in plain text** in databases

### 2. Token Rotation
```go
// Implement refresh token rotation for enhanced security
func (h *Handler) RefreshTokenWithRotation(c echo.Context) error {
    // ... validate refresh token ...
    
    // Generate new token pair (both access and refresh)
    newTokenRequest := security.TokenRequest{
        UserClaims: extractedClaims,
        Metadata: security.TokenMetadata{
            Purpose:     "refresh",
            Source:      "web",
            Environment: "production",
            CreatedAt:   time.Now(),
        },
        RefreshToken: true, // Generate new refresh token too
    }
    
    tokenResponse, err := h.security.GenerateToken(newTokenRequest)
    if err != nil {
        return echo.NewHTTPError(500, "Failed to generate new tokens")
    }
    
    // Invalidate old refresh token (implement token blacklist)
    h.tokenBlacklist.Add(oldRefreshToken)
    
    return c.JSON(200, map[string]interface{}{
        "access_token":  tokenResponse.AccessToken,
        "refresh_token": tokenResponse.RefreshToken, // New refresh token
        "token_type":    tokenResponse.TokenType,
        "expires_in":    tokenResponse.ExpiresIn,
        "expires_at":    tokenResponse.ExpiresAt,
    })
}
```

### 3. Scope Management
```go
// Different scopes for different token types
tokenRequest := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:   userID,
        Username: username,
        Email:    email,
    },
    Metadata: security.TokenMetadata{
        Purpose: "login",
        Source:  "web",
        // Access token gets full scopes
        Scopes: []string{"read", "write", "admin"},
    },
    RefreshToken: true,
}

// When refreshing, you might want to limit scopes
refreshTokenRequest := security.TokenRequest{
    UserClaims: extractedClaims,
    Metadata: security.TokenMetadata{
        Purpose: "refresh",
        Source:  "web",
        // Refresh might have limited scopes
        Scopes: []string{"read", "write"}, // No admin scope
    },
}
```

## 🧪 **Testing Token Pairs**

### 1. Test Login with Token Pair
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"username": "john_doe123", "password": "SecurePass123!"}' \
  http://localhost:8888/login
```

**Expected Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "expires_at": "2025-09-25T16:45:12+07:00",
  "message": "Login successful",
  "user": {
    "id": 1,
    "username": "john_doe123",
    "email": "john.doe@example.com"
  }
}
```

### 2. Test Refresh Token
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"refresh_token": "eyJhbGciOiJIUzI1NiIs..."}' \
  http://localhost:8888/refresh
```

**Expected Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer", 
  "expires_in": 3600,
  "expires_at": "2025-09-25T17:45:12+07:00",
  "message": "Token refreshed successfully"
}
```

## 🎯 **Summary**

Token pairs provide a secure and flexible authentication system:

1. **Enable refresh tokens** by setting `RefreshToken: true` in `TokenRequest`
2. **Store tokens securely** on the client side
3. **Use access tokens** for API requests
4. **Use refresh tokens** to get new access tokens when they expire
5. **Implement token rotation** for enhanced security
6. **Handle token expiration** gracefully in your applications

The strongly-typed API makes it easy to work with token pairs while maintaining type safety and validation throughout the process.
