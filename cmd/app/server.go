package main

import (
	"net/http"
	"os"
	"time"

	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
)

const (
	serverAddr  = "SERVER_ADDR"
	certFile    = "CERT_FILE"
	certFileKey = "CERT_FILE_KEY"
)

func SetupServer(router http.Handler) error {
	addr := os.Getenv(serverAddr)
	if addr == "" {
		return errs.NewMissedEnvironmentVariableError(serverAddr)
	}

	srv := http.Server{
		Addr:         addr,
		Handler:      router,
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
	}

	cert := os.Getenv(certFile)
	if cert == "" {
		return errs.NewMissedEnvironmentVariableError(certFile)
	}

	key := os.Getenv(certFileKey)
	if key == "" {
		return errs.NewMissedEnvironmentVariableError(certFileKey)
	}

	return srv.ListenAndServeTLS(cert, key)
}
