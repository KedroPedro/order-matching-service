package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	jwtmanager "github.com/KedroPedro/order-matching-engine/internal/application/jwt_manager"
	"github.com/KedroPedro/order-matching-engine/internal/application/usecases"
	"github.com/KedroPedro/order-matching-engine/internal/controller/controllertypes"
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
	"github.com/rs/zerolog/log"
)

type UsersHandler struct {
	createUserUsecase *usecases.CreateUserUsecase
	loginUsecase      *usecases.LoginUsecase
	mux               *http.ServeMux
}

func NewUsersHandler(
	mux *http.ServeMux,
	createUserUsecase *usecases.CreateUserUsecase,
	loginUsecase *usecases.LoginUsecase,
) *UsersHandler {
	h := &UsersHandler{
		createUserUsecase: createUserUsecase,
		loginUsecase:      loginUsecase,
		mux:               mux,
	}

	h.mux.HandleFunc("POST /users/create", h.createUserHandler)
	h.mux.HandleFunc("GET /users/login", h.loginHandler)

	return h
}

func (this *UsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.mux.ServeHTTP(w, r)
}

func (this *UsersHandler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := this.createUser(ctx, w, r); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err).Send()
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (this *UsersHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := this.login(ctx, w, r); err != nil {
		log.Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (this *UsersHandler) createUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var user controllertypes.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return errs.NewHandlerError("body decode error", err)
	}

	if err := this.createUserUsecase.Execute(ctx, user.ToDomainEntity()); err != nil {
		return errs.NewHandlerError("create user error", err)
	}

	jwtToken, err := jwtmanager.Encode(user.Id)
	if err != nil {
		return errs.NewHandlerError("jwt encoding error", err)
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:     "jwt",
			Value:    jwtToken,
			Expires:  time.Now().Add(time.Hour * 24),
			Secure:   true,
			HttpOnly: true,
		},
	)
	return nil
}

func (this *UsersHandler) login(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var fields map[string]any

	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		return errs.NewHandlerError("body decode error", err)
	}

	login, ok := fields["login"]
	if !ok {
		return errs.NewHandlerError("login error", nil)
	}

	sLogin, ok := login.(string)
	if !ok {
		return errs.NewTypeError("converting login field to string error")
	}

	user, err := this.loginUsecase.Execute(ctx, sLogin)
	if err != nil {
		return errs.NewHandlerError("login error", err)
	}

	jwtToken, err := jwtmanager.Encode(user.Id)
	if err != nil {
		return errs.NewHandlerError("jwt encoding error", err)
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:     "jwt",
			Value:    jwtToken,
			Expires:  time.Now().Add(time.Hour * 24),
			Secure:   true,
			HttpOnly: true,
		},
	)

	data, err := json.Marshal(controllertypes.FromDomainEntity(user))
	if err != nil {
		return errs.NewHandlerError("user json marshaling error", err)
	}

	_, err = w.Write(data)
	if err != nil {
		return errs.NewHandlerError("data write error", err)
	}

	return nil
}
