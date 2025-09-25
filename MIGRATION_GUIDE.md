# Security API: Strongly-Typed Interface Guide

This guide explains how to use the clean, strongly-typed security API that replaces the legacy `map[string]interface{}` based approach.

## Overview

The security package provides strongly-typed interfaces for:
- **Type Safety**: Compile-time type checking prevents runtime errors
- **Better IDE Support**: Auto-completion and better refactoring support
- **Validation**: Built-in validation for all data structures
- **Documentation**: Self-documenting code with clear field types
- **Clean API**: No legacy methods or compatibility layers

## Core Concepts

### Strongly-Typed Structures
The API uses well-defined structures instead of generic maps:
- `TokenRequest` - For token generation requests
- `TokenResponse` - For token generation responses
- `ValidationRequest` - For token validation requests
- `ValidationResponse` - For token validation responses
- `UserClaims` - For user authentication data
- `TokenMetadata` - For token context and metadata
- `CustomClaims` - For application-specific data

## API Usage

### Token Generation

#### Strongly-Typed Token Generation
```go
request := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:      userID,
        Username:    "john_doe",
        Email:       "john@example.com",
        Role:        "admin",
        Permissions: []string{"read", "write"},
    },
    Metadata: security.TokenMetadata{
        Purpose:     "login",
        Source:      "web",
        Environment: "production",
        Scopes:      []string{"read", "write"},
        CreatedAt:   time.Now(),
    },
    CustomClaims: security.CustomClaims{
        Theme:    "dark",
        Language: "en",
    },
}

response, err := securityService.GenerateToken(request)
if err != nil {
    return err
}

// Access strongly-typed response
accessToken := response.AccessToken
tokenType := response.TokenType
expiresAt := response.ExpiresAt
```

### Token Validation

#### Strongly-Typed Token Validation
```go
request := security.ValidationRequest{
    Token:            tokenString,
    RequiredScopes:   []string{"read"},
    RequiredAudience: "api",
}

response, err := securityService.ValidateToken(request)
if err != nil {
    return err
}

if !response.Valid {
    return errors.New("invalid token")
}

// Access strongly-typed data
username := response.UserClaims.Username
email := response.UserClaims.Email
permissions := response.UserClaims.Permissions
scopes := response.Scopes
expiresAt := response.ExpiresAt
```

## Common Usage Patterns

### Basic User Authentication Token
```go
request := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:   userID,
        Username: username,
        Email:    email,
    },
    Metadata: security.TokenMetadata{
        Purpose:     "login",
        Source:      "web",
        Environment: "production",
        CreatedAt:   time.Now(),
    },
}

response, err := securityService.GenerateToken(request)
```

### API Token with Scopes
```go
request := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID: userID,
    },
    Metadata: security.TokenMetadata{
        Purpose:     "api",
        Source:      "api",
        Environment: "production",
        Scopes:      []string{"api:read", "api:write"},
        CreatedAt:   time.Now(),
    },
}

response, err := securityService.GenerateToken(request)
```

### Web Session Token
```go
request := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:    userID,
        Username:  username,
        Email:     email,
        SessionID: sessionID,
        ClientID:  "web",
    },
    Metadata: security.TokenMetadata{
        Purpose:     "session",
        Source:      "web",
        Environment: "production",
        CreatedAt:   time.Now(),
    },
}

response, err := securityService.GenerateToken(request)
```

## Validation

### Built-in Validation
All request structures include built-in validation:

```go
request := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID: "", // ❌ Will fail validation - required field
        Email:  "invalid-email", // ❌ Will fail validation - invalid format
    },
}

// Validate before using
if err := request.Validate(); err != nil {
    // Handle validation errors
    if validationErrors, ok := err.(security.ValidationErrors); ok {
        for _, ve := range validationErrors {
            log.Printf("Field %s: %s", ve.Field, ve.Message)
        }
    }
    return err
}
```

## Complete Example

### Login Handler
```go
func (h *Handler) Login(c echo.Context) error {
    // ... authentication logic ...

    // Generate token using strongly-typed interface
    request := security.TokenRequest{
        UserClaims: security.UserClaims{
            UserID:   strconv.Itoa(int(user.ID)),
            Username: user.Username,
            Email:    user.Email,
            Role:     user.Role,
            ClientID: "web",
        },
        Metadata: security.TokenMetadata{
            Purpose:     "login",
            Source:      "web",
            Environment: "production",
            Scopes:      []string{"read", "write"},
            CreatedAt:   time.Now(),
        },
    }

    response, err := h.security.GenerateToken(request)
    if err != nil {
        return err
    }

    return c.JSON(200, map[string]interface{}{
        "access_token": response.AccessToken,
        "token_type":   response.TokenType,
        "expires_in":   response.ExpiresIn,
        "expires_at":   response.ExpiresAt,
        "user": map[string]interface{}{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
        },
    })
}
```

### Token Validation Handler
```go
func (h *Handler) ValidateToken(c echo.Context) error {
    token := extractTokenFromHeader(c)

    request := security.ValidationRequest{
        Token:          token,
        RequiredScopes: []string{"read"},
    }

    response, err := h.security.ValidateToken(request)
    if err != nil {
        return echo.NewHTTPError(401, "Invalid token")
    }

    if !response.Valid {
        return echo.NewHTTPError(401, "Token validation failed")
    }

    return c.JSON(200, map[string]interface{}{
        "valid":       response.Valid,
        "user_claims": response.UserClaims,
        "expires_at":  response.ExpiresAt,
        "scopes":      response.Scopes,
    })
}
```

## Benefits of Strongly-Typed API

### Type Safety
```go
// Compile-time error if field doesn't exist
username := response.UserClaims.Username // ✅ Safe
email := response.UserClaims.Email       // ✅ Safe
invalid := response.UserClaims.Invalid   // ❌ Compile error
```

### Automatic Validation
```go
request := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID: "", // ❌ Will fail validation
        Email:  "invalid-email", // ❌ Will fail validation
    },
}

err := request.Validate() // Returns detailed validation errors
```

### Enhanced Developer Experience
- **Auto-completion** for all fields
- **Go to definition** works correctly
- **Refactoring** is safer and more reliable
- **Documentation** is available inline
- **Self-documenting** code with clear field types

## Common Patterns and Solutions

### 1. Custom Application Data
**Use Case**: Store application-specific data in tokens
```go
request := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:   userID,
        Username: username,
        Email:    email,
    },
    CustomClaims: security.CustomClaims{
        Theme:    "dark",
        Language: "en",
        Data: map[string]interface{}{
            "subscription_tier": "premium",
            "feature_flags":     []string{"new_ui", "beta_features"},
        },
    },
}
```

### 2. Handling Validation Errors
**Use Case**: Detailed error handling for validation failures
```go
if err := request.Validate(); err != nil {
    if validationErrors, ok := err.(security.ValidationErrors); ok {
        for _, ve := range validationErrors {
            log.Printf("Validation error - Field: %s, Message: %s", ve.Field, ve.Message)
        }
        return echo.NewHTTPError(400, "Invalid request data")
    }
    return echo.NewHTTPError(500, "Internal validation error")
}
```

### 3. Conditional Token Features
**Use Case**: Different token configurations based on context
```go
request := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID:   userID,
        Username: username,
        Email:    email,
    },
    Metadata: security.TokenMetadata{
        Purpose:     "login",
        Source:      source, // "web", "mobile", "api"
        Environment: environment,
        CreatedAt:   time.Now(),
    },
}

// Add scopes based on user role
if user.Role == "admin" {
    request.Metadata.Scopes = []string{"read", "write", "admin"}
} else {
    request.Metadata.Scopes = []string{"read"}
}

// Enable refresh token for mobile clients
if source == "mobile" {
    request.RefreshToken = true
}
```

## Testing

### Unit Tests
```go
func TestTokenGeneration(t *testing.T) {
    request := security.TokenRequest{
        UserClaims: security.UserClaims{
            UserID:   "user123",
            Username: "john",
            Email:    "john@example.com",
        },
        Metadata: security.TokenMetadata{
            Purpose:     "test",
            Source:      "unit_test",
            Environment: "test",
            CreatedAt:   time.Now(),
        },
    }

    response, err := securityService.GenerateToken(request)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.AccessToken)
    assert.Equal(t, "Bearer", response.TokenType)
}
```

### Integration Tests
```go
func TestEndToEndTokenFlow(t *testing.T) {
    // Generate token
    request := security.TokenRequest{
        UserClaims: security.UserClaims{
            UserID:   "user123",
            Username: "john",
            Email:    "john@example.com",
        },
        Metadata: security.TokenMetadata{
            Purpose:     "test",
            Source:      "integration_test",
            Environment: "test",
            Scopes:      []string{"read", "write"},
            CreatedAt:   time.Now(),
        },
    }

    response, err := securityService.GenerateToken(request)
    require.NoError(t, err)

    // Validate token
    validationRequest := security.ValidationRequest{
        Token:          response.AccessToken,
        RequiredScopes: []string{"read"},
    }

    validationResponse, err := securityService.ValidateToken(validationRequest)
    require.NoError(t, err)
    assert.True(t, validationResponse.Valid)
    assert.Equal(t, "john", validationResponse.UserClaims.Username)
}
```

## Error Handling

### Common Error Messages
- `"validation error for field 'user_id': is required"` - UserID field is missing or empty
- `"validation error for field 'email': must be a valid email address"` - Invalid email format
- `"token has invalid claims: token is expired"` - Token has expired
- `"insufficient scopes: required [read write], got [read]"` - Token doesn't have required scopes

### Debugging Tips
1. **Check validation errors** for specific field issues
2. **Review token expiration** settings in your JWT configuration
3. **Verify scopes** match between token generation and validation
4. **Use structured logging** to track token lifecycle
5. **Test with known good data** to isolate issues

## Best Practices

### Security Considerations
- Always validate tokens on protected endpoints
- Use appropriate scopes for different operations
- Set reasonable token expiration times
- Include environment information in metadata
- Log security events for monitoring

### Performance Tips
- Cache validation results when appropriate
- Use connection pooling for database operations
- Consider token refresh strategies for long-lived sessions
- Monitor token generation and validation metrics

### Code Organization
- Create helper functions for common token patterns
- Use constants for scope definitions
- Implement middleware for automatic token validation
- Structure your claims data consistently across the application

This strongly-typed API provides significant benefits in terms of type safety, validation, and maintainability while offering a clean, modern interface for security operations.
