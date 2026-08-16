package handlers

import (
	"net/http"
	"time"

	jwtmanager "github.com/KedroPedro/order-matching-engine/order_service/internal/application/jwt_manager"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
)

func getCookieValue(cookie *http.Cookie) (string, error) {
	if cookie.Expires.After(time.Now()) {
		return "", errs.NewHandlerError("cookies expired", nil)
	}

	value, err := jwtmanager.Decode(cookie.Value)
	if err != nil {
		return "", errs.NewHandlerError("cooket jwt decode error", nil)
	}

	return value, nil
}
