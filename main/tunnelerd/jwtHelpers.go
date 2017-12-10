package main

import (
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
	"time"
	"github.com/spf13/viper"
	"errors"
)

func generateToken(username string, expirationTime time.Duration) (string, error) {

	if viper.GetBool("SecretKey.isRandom") {
		return "", errors.New("SecretKey is random so token will be useless. Please specify SecretKey on config file")
	}

	if username == "" {
		return "", errors.New("need to specify username")
	}

	if expirationTime <= 0 {
		return "", errors.New("need to specify a positive expirationTime duration")
	}

	logger.Info("Generating token for user %s. Expiration time %d days.", username, expirationTime)

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		Issuer:    "tunnelerd",
		Subject:   username,
		ExpiresAt: now.Add(expirationTime * 24 * time.Hour).Unix(),
		IssuedAt:  now.Unix(),
	})

	tokenStr, err := token.SignedString(secret_key)

	if err != nil{
		return "", err
	}

	return tokenStr, nil
}

func validToken(callback func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, request *http.Request) {
		auth := strings.SplitN(
			request.Header.Get("Authorization"),
			" ",
			2,
		)

		if len(auth) != 2 || auth[0] != "Bearer" {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		tokenStr := auth[1]

		token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
			return secret_key, nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			logger.Info("Claims: %v", claims)
			callback(w, request)

		} else {
			logger.Error(err)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

	}
}
