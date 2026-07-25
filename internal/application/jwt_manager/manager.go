package jwtmanager

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

const (
	secretKey = "verysecretkey"
)

func Encode(id string) (string, error) {
	t := jwt.NewWithClaims(
		jwt.SigningMethodES256,
		jwt.MapClaims{
			"id": id,
		},
	)

	return t.SignedString(secretKey)
}

func Decode(signed string) (string, error) {
	token, err := jwt.Parse(
		signed,
		func(t *jwt.Token) (any, error) {
			return secretKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}),
	)

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("error") //TODO: add normal error
	}

	id, ok := claims["id"].(string)
	if !ok {
		return "", errors.New("error") //TODO: add normal error
	}

	return id, nil
}
