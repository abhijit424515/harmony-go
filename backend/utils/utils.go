package utils

import (
	"encoding/base64"
	"fmt"
	"maps"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateAccessToken(payload map[string]any, expirationTime time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	maps.Copy(claims, payload)

	t := time.Now()
	claims["exp"] = t.Add(expirationTime).Unix()
	claims["iat"] = t.Unix()

	decodedKey, err := base64.StdEncoding.DecodeString(os.Getenv("JWT_SK"))
	if err != nil {
		return "", fmt.Errorf("[error] decoding signing key: %v", err)
	}

	tokenString, err := token.SignedString(decodedKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyAndDecodeToken(tokenString string) (map[string]any, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		decodedKey, err := base64.StdEncoding.DecodeString(os.Getenv("JWT_SK"))
		if err != nil {
			return "", fmt.Errorf("[error] decoding signing key: %v", err)
		}
		return decodedKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				return nil, fmt.Errorf("token has expired")
			}
		} else {
			return nil, fmt.Errorf("missing expiration claim")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
