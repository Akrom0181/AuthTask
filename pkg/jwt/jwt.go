package jwt

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var config = struct {
	SignedKey []byte
}{
	SignedKey: []byte(os.Getenv("SECRET_KEY_JWT")),
}

// GenJWT generates an access token and a refresh token with claims.
func GenJWT(claims map[string]interface{}) (string, string, error) {
	accessTokenClaims := jwt.MapClaims{
		"iss": "user",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	refreshTokenClaims := jwt.MapClaims{
		"iss": "user",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(10 * 24 * time.Hour).Unix(),
	}

	for k, v := range claims {
		accessTokenClaims[k] = v
		refreshTokenClaims[k] = v
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	accessTokenString, err := accessToken.SignedString(config.SignedKey)
	if err != nil {
		return "", "", fmt.Errorf("error generating access token: %w", err)
	}

	refreshTokenString, err := refreshToken.SignedString(config.SignedKey)
	if err != nil {
		return "", "", fmt.Errorf("error generating refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, nil
}

func ExtractClaims(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return config.SignedKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// VerifyJWT verifies the validity of a JWT and extracts its claims.
func VerifyJWT(tokenStr string) (jwt.MapClaims, error) {
	claims, err := ExtractClaims(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	// Debug: Print claims
	fmt.Printf("Extracted claims: %+v\n", claims)

	// Example: Validate specific claim
	if claims["iss"] != "user" {
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}
