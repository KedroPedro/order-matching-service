package jwtmanager

import (
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
)

const (
	secretKey = "verysecretkey"
)

func Encode(id string) (string, error) {
	t := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id": id,
		},
	)

	return t.SignedString([]byte(secretKey))
}

func Decode(signed string) (string, error) {
	token, err := jwt.Parse(
		signed,
		func(t *jwt.Token) (any, error) {
			return []byte(secretKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errs.NewTypeError("converting token.Claims to jwt.MapClaims error")
	}

	id, ok := claims["id"].(string)
	if !ok {
		return "", errs.NewTypeError("converting claims[\"id\"] to string error")
	}

	return id, nil
}
