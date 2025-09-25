package jwt

import "github.com/golang-jwt/jwt/v5"

type claimsParser struct {
	config Config
}

func ParserServiceProvider(config Config) ClaimsParser {
	return NewClaimsParser(config)
}

func NewClaimsParser(config Config) ClaimsParser {
	return &claimsParser{
		config: config,
	}
}

func (p *claimsParser) ParseClaims(token *jwt.Token) (*Claims, error) {
	if token == nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return p.ExtractClaims(claims)
}

func (p *claimsParser) ExtractClaims(claims jwt.MapClaims) (*Claims, error) {
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidClaims
	}

	parsedClaims := &Claims{
		UserID:    userID,
		TokenType: p.getStringClaim(claims, "token_type"),
		ClientID:  p.getStringClaim(claims, "client_id"),
		DeviceID:  p.getStringClaim(claims, "device_id"),
		SessionID: p.getStringClaim(claims, "session_id"),
		Issuer:    p.getStringClaim(claims, "iss"),
		Audience:  p.getStringClaim(claims, "aud"),
		Subject:   p.getStringClaim(claims, "sub"),
		ExpiresAt: p.getInt64Claim(claims, "exp"),
		IssuedAt:  p.getInt64Claim(claims, "iat"),
		NotBefore: p.getInt64Claim(claims, "nbf"),
		Scopes:    p.getStringArrayClaim(claims, "scopes"),
		Metadata:  p.getMapClaim(claims, "metadata"),
		Custom:    p.getCustomClaims(claims),
	}

	return parsedClaims, nil
}

func (p *claimsParser) getStringClaim(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (p *claimsParser) getInt64Claim(claims jwt.MapClaims, key string) int64 {
	if value, ok := claims[key]; ok {
		switch v := value.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		}
	}
	return 0
}

func (p *claimsParser) getStringArrayClaim(claims jwt.MapClaims, key string) []string {
	if value, ok := claims[key]; ok {
		if arr, ok := value.([]interface{}); ok {
			var result []string
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return nil
}

func (p *claimsParser) getMapClaim(claims jwt.MapClaims, key string) map[string]interface{} {
	if value, ok := claims[key]; ok {
		if m, ok := value.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

func (p *claimsParser) getCustomClaims(claims jwt.MapClaims) map[string]interface{} {
	protectedKeys := map[string]bool{
		"user_id":    true,
		"token_type": true,
		"client_id":  true,
		"device_id":  true,
		"session_id": true,
		"scopes":     true,
		"metadata":   true,
		"iss":        true,
		"aud":        true,
		"sub":        true,
		"exp":        true,
		"iat":        true,
		"nbf":        true,
	}

	customClaims := make(map[string]interface{})
	for key, value := range claims {
		if !protectedKeys[key] {
			customClaims[key] = value
		}
	}
	return customClaims
}
