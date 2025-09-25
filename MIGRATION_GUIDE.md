# Migration Guide: From `interface{}` to Strong Types

This guide helps you migrate from the legacy `map[string]interface{}` based security API to the new strongly-typed API.

## Overview

The security package has been enhanced with strongly-typed interfaces to provide:
- **Type Safety**: Compile-time type checking prevents runtime errors
- **Better IDE Support**: Auto-completion and better refactoring support
- **Validation**: Built-in validation for all data structures
- **Documentation**: Self-documenting code with clear field types
- **Backward Compatibility**: Legacy methods still work during migration

## Migration Strategy

### Phase 1: Preparation
1. Update your imports to include the new types
2. Review your current usage of `GenerateToken` and `ValidateToken`
3. Identify the data you're currently passing in `map[string]interface{}`

### Phase 2: Gradual Migration
1. Start using the new `GenerateTokenV2` and `ValidateTokenV2` methods
2. Use migration helpers for quick conversion
3. Update one endpoint at a time

### Phase 3: Cleanup
1. Remove usage of legacy methods
2. Enable strict mode to prevent legacy method usage
3. Update tests to use new interfaces

## API Changes

### Token Generation

#### Before (Legacy)
```go
token, err := securityService.GenerateToken(userID, map[string]interface{}{
    "username": "john_doe",
    "email": "john@example.com",
    "role": "admin",
    "permissions": []string{"read", "write"},
    "custom": map[string]interface{}{
        "theme": "dark",
        "language": "en",
    },
})
```

#### After (Strongly-typed)
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

response, err := securityService.GenerateTokenV2(request)
```

#### Quick Migration (Using Builder)
```go
request := security.NewQuickMigrationBuilder(userID).
    WithUserInfo("john_doe", "john@example.com").
    WithRole("admin").
    WithPermissions("read", "write").
    WithMetadata("login", "web", "production").
    WithPreference("theme", "dark").
    WithPreference("language", "en").
    Build()

response, err := securityService.GenerateTokenV2(request)
```

### Token Validation

#### Before (Legacy)
```go
token, err := securityService.ValidateToken(tokenString)
if err != nil {
    return err
}

claims, err := securityService.ParseClaims(token)
if err != nil {
    return err
}

// Extract data manually from claims.Metadata
username := claims.GetMetadataString("username")
email := claims.GetMetadataString("email")
```

#### After (Strongly-typed)
```go
request := security.ValidationRequest{
    Token:            tokenString,
    RequiredScopes:   []string{"read"},
    RequiredAudience: "api",
}

response, err := securityService.ValidateTokenV2(request)
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
```

#### Quick Migration (Using Builder)
```go
request := security.NewValidationBuilder(tokenString).
    WithRequiredScopes("read", "write").
    WithRequiredAudience("api").
    Build()

response, err := securityService.ValidateTokenV2(request)
```

## Migration Helpers

### Automatic Conversion
The migration helpers can automatically convert your existing `map[string]interface{}` data:

```go
// Convert legacy claims to new format
legacyClaims := map[string]interface{}{
    "username": "john_doe",
    "email": "john@example.com",
    "role": "admin",
}

request, err := security.ConvertLegacyClaimsToTokenRequest(userID, legacyClaims)
if err != nil {
    return err
}

response, err := securityService.GenerateTokenV2(request)
```

### Common Migration Patterns

#### Basic User Token
```go
// Before
token, err := securityService.GenerateToken(userID, map[string]interface{}{
    "username": username,
    "email": email,
})

// After (using helper)
request := security.MigrateBasicUserToken(userID, username, email)
response, err := securityService.GenerateTokenV2(request)
```

#### API Token with Scopes
```go
// Before
token, err := securityService.GenerateToken(userID, map[string]interface{}{
    "scopes": []string{"api:read", "api:write"},
})

// After (using helper)
request := security.MigrateAPIToken(userID, []string{"api:read", "api:write"})
response, err := securityService.GenerateTokenV2(request)
```

#### Web Session Token
```go
// Before
token, err := securityService.GenerateToken(userID, map[string]interface{}{
    "username": username,
    "email": email,
    "session_id": sessionID,
})

// After (using helper)
request := security.MigrateWebSessionToken(userID, username, email, sessionID)
response, err := securityService.GenerateTokenV2(request)
```

## Compatibility Configuration

### Default Mode (Recommended)
```go
// Both old and new methods work, with deprecation warnings
config := security.DefaultCompatibilityConfig()
// Mode: CompatibilityModeEnabled
// LogDeprecationWarnings: true
// FailOnLegacyMethods: false
```

### Strict Mode (For New Projects)
```go
config := security.CompatibilityConfig{
    Mode:                   security.StrictMode,
    LogDeprecationWarnings: true,
    FailOnLegacyMethods:    true,
}
```

### Legacy Mode (For Gradual Migration)
```go
config := security.CompatibilityConfig{
    Mode:                   security.LegacyMode,
    LogDeprecationWarnings: false,
    FailOnLegacyMethods:    false,
}
```

## Step-by-Step Migration Example

### 1. Current Code
```go
func (h *Handler) Login(c echo.Context) error {
    // ... authentication logic ...
    
    token, err := h.security.GenerateToken(user.ID, map[string]interface{}{
        "username": user.Username,
        "email": user.Email,
        "role": user.Role,
    })
    
    return c.JSON(200, map[string]interface{}{
        "token": token,
    })
}
```

### 2. Add New Method (Parallel Implementation)
```go
func (h *Handler) Login(c echo.Context) error {
    // ... authentication logic ...
    
    // New strongly-typed method
    request := security.NewQuickMigrationBuilder(user.ID).
        WithUserInfo(user.Username, user.Email).
        WithRole(user.Role).
        WithMetadata("login", "web", "production").
        Build()
    
    response, err := h.security.GenerateTokenV2(request)
    if err != nil {
        return err
    }
    
    return c.JSON(200, map[string]interface{}{
        "access_token": response.AccessToken,
        "token_type":   response.TokenType,
        "expires_in":   response.ExpiresIn,
        "expires_at":   response.ExpiresAt,
    })
}
```

### 3. Update Validation
```go
func (h *Handler) ValidateToken(c echo.Context) error {
    token := extractTokenFromHeader(c)
    
    request := security.NewValidationBuilder(token).
        WithRequiredScopes("api:read").
        Build()
    
    response, err := h.security.ValidateTokenV2(request)
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
    })
}
```

## Benefits After Migration

### Type Safety
```go
// Compile-time error if field doesn't exist
username := response.UserClaims.Username // ✅ Safe
email := response.UserClaims.Email       // ✅ Safe
invalid := response.UserClaims.Invalid   // ❌ Compile error
```

### Validation
```go
request := security.TokenRequest{
    UserClaims: security.UserClaims{
        UserID: "", // ❌ Will fail validation
        Email:  "invalid-email", // ❌ Will fail validation
    },
}

err := request.Validate() // Returns detailed validation errors
```

### Better IDE Support
- Auto-completion for all fields
- Go to definition works correctly
- Refactoring is safer and more reliable
- Documentation is available inline

## Common Pitfalls and Solutions

### 1. Nested Map Conversion
**Problem**: Complex nested maps don't convert automatically
```go
// This won't convert properly
claims := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "preferences": map[string]interface{}{
                "theme": "dark",
            },
        },
    },
}
```

**Solution**: Use the builder pattern or manual construction
```go
request := security.NewQuickMigrationBuilder(userID).
    WithPreference("theme", "dark").
    Build()
```

### 2. Custom Data Types
**Problem**: Custom types in `interface{}` fields
```go
// Custom types won't convert automatically
claims := map[string]interface{}{
    "custom_data": MyCustomStruct{},
}
```

**Solution**: Use the `Data` field in `CustomClaims`
```go
request.CustomClaims.Data = map[string]interface{}{
    "custom_data": myCustomStruct,
}
```

### 3. Validation Errors
**Problem**: Validation fails with unclear errors

**Solution**: Check validation errors for details
```go
if err := request.Validate(); err != nil {
    if validationErrors, ok := err.(security.ValidationErrors); ok {
        for _, ve := range validationErrors {
            log.Printf("Field %s: %s", ve.Field, ve.Message)
        }
    }
    return err
}
```

## Testing Your Migration

### Unit Tests
```go
func TestTokenGeneration(t *testing.T) {
    request := security.NewQuickMigrationBuilder("user123").
        WithUserInfo("john", "john@example.com").
        Build()
    
    response, err := securityService.GenerateTokenV2(request)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.AccessToken)
    assert.Equal(t, "Bearer", response.TokenType)
}
```

### Integration Tests
```go
func TestEndToEndTokenFlow(t *testing.T) {
    // Generate token
    request := security.NewQuickMigrationBuilder("user123").
        WithUserInfo("john", "john@example.com").
        WithScopes("read", "write").
        Build()
    
    response, err := securityService.GenerateTokenV2(request)
    require.NoError(t, err)
    
    // Validate token
    validationRequest := security.NewValidationBuilder(response.AccessToken).
        WithRequiredScopes("read").
        Build()
    
    validationResponse, err := securityService.ValidateTokenV2(validationRequest)
    require.NoError(t, err)
    assert.True(t, validationResponse.Valid)
    assert.Equal(t, "john", validationResponse.UserClaims.Username)
}
```

## Timeline Recommendations

### Week 1-2: Preparation
- Review current usage
- Add new types to your project
- Set up compatibility mode

### Week 3-4: Parallel Implementation
- Implement new methods alongside old ones
- Use migration helpers
- Test thoroughly

### Week 5-6: Gradual Rollout
- Replace old methods one endpoint at a time
- Monitor for issues
- Update tests

### Week 7-8: Cleanup
- Remove legacy method usage
- Enable strict mode
- Update documentation

## Support and Troubleshooting

### Enable Debug Logging
```go
config := security.CompatibilityConfig{
    Mode:                   security.CompatibilityModeEnabled,
    LogDeprecationWarnings: true,
    FailOnLegacyMethods:    false,
}
```

### Common Error Messages
- `"validation error for field 'user_id': is required"` - UserID is missing
- `"validation error for field 'email': must be a valid email address"` - Invalid email format
- `"legacy method GenerateToken is deprecated, use GenerateTokenV2 instead"` - Use new method

### Getting Help
1. Check validation errors for specific field issues
2. Use migration helpers for common patterns
3. Review examples in this guide
4. Enable debug logging to see what's happening

This migration provides significant benefits in terms of type safety, validation, and maintainability while maintaining backward compatibility during the transition period.
