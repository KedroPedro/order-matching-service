package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/application/usecases"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/controller/controllertypes"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
	"github.com/rs/zerolog/log"
)

type OrdersHandler struct {
	mux           *http.ServeMux
	addUsecase    *usecases.AddOrderUsecase
	cancelUsecase *usecases.CancelOrderUsecase
}

func NewOrdersHandler(
	addUsecase *usecases.AddOrderUsecase,
	cancelUsecase *usecases.CancelOrderUsecase,
	mux *http.ServeMux,
) *OrdersHandler {
	h := &OrdersHandler{
		addUsecase:    addUsecase,
		cancelUsecase: cancelUsecase,
		mux:           mux,
	}
	h.mux.HandleFunc("POST /order", h.addOrderHandler)
	h.mux.HandleFunc("DELETE /order", h.cancelOrderHandler)
	return h
}

func (this *OrdersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.mux.ServeHTTP(w, r)
}

func (this *OrdersHandler) addOrderHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Err(err).Send()
		return
	}

	id, err := getCookieValue(cookie)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Err(err).Send()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := this.addOrder(ctx, id, r); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err).Send()
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (this *OrdersHandler) cancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Err(err).Send()
		return
	}

	id, err := getCookieValue(cookie)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Err(err).Send()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := this.cancelOrder(ctx, id, r); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err).Send()
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (this *OrdersHandler) addOrder(ctx context.Context, userId string, r *http.Request) error {
	var newOrder controllertypes.Order

	if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
		return errs.NewHandlerError("body decoding error", err)
	}

	convertedOrder, err := newOrder.ToDomainEntity(userId)
	if err != nil {
		return errs.NewHandlerError("order converting error", err)
	}

	if err := this.addUsecase.Execute(ctx, convertedOrder); err != nil {
		return errs.NewHandlerError("add order error", err)
	}

	return nil
}

func (this *OrdersHandler) cancelOrder(ctx context.Context, userId string, r *http.Request) error {
	var values map[string]any

	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		return errs.NewHandlerError("body decoding error", err)
	}

	orderId, ok := values["order_id"]
	if !ok {
		return errs.NewHandlerError("order id parameter is missed", nil)
	}

	sOrderId, ok := orderId.(string)
	if !ok {
		return errs.NewHandlerError("cancel order error", nil)
	}

	if err := this.cancelUsecase.Execute(ctx, sOrderId, userId); err != nil {
		return errs.NewHandlerError("cancel order error", err)
	}

	return nil
}
