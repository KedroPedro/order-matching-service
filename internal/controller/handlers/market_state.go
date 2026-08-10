package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/KedroPedro/order-matching-engine/internal/application/usecases"
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
	"github.com/coder/websocket"
	"github.com/rs/zerolog/log"
)

type MarketStateHandler struct {
	getStateUsecase *usecases.GetStateUsecase
	conns           map[string]*Connection
	mux             *http.ServeMux
	mu              sync.Mutex
}

func NewMarketStateHandler(
	ctx context.Context,
	getStateUsecase *usecases.GetStateUsecase,
	mux *http.ServeMux,
) *MarketStateHandler {
	h := &MarketStateHandler{
		getStateUsecase: getStateUsecase,
		conns:           make(map[string]*Connection),
		mux:             mux,
	}

	h.mux.HandleFunc("GET /ws/subscribe", h.subscribeHandler)

	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				fCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				asks, bids, err := h.getStateUsecase.Execute(fCtx)
				if err != nil {
					log.Err(err).Send()
					continue
				}
				msg, err := json.Marshal(map[string]any{
					"asks": asks,
					"bids": bids,
				})

				if err != nil {
					log.Err(err).Send()
					continue
				}

				go h.writeToAllConns(msg)

				cancel()
			case <-ctx.Done():
				for _, c := range h.conns {
					c.conn.Close(websocket.StatusNormalClosure, "server shutdown")
				}
				return
			}
		}

	}()

	return h
}

type Connection struct {
	conn      *websocket.Conn
	id        string
	msgCh     chan []byte
	closeSlow func()
}

func (this *MarketStateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.mux.ServeHTTP(w, r)
}

func (this *MarketStateHandler) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		log.Err(err).Send()
		return
	}

	id, err := getCookieValue(cookie)
	if err != nil {
		log.Err(err).Send()
		return
	}

	if err := this.subscribe(id, w, r); err != nil {
		log.Err(err).Send()
		return
	}
}

func (this *MarketStateHandler) subscribe(id string, w http.ResponseWriter, r *http.Request) error {
	var c *websocket.Conn
	var closed bool
	var mu sync.Mutex

	newConn := &Connection{
		conn:  c,
		id:    id,
		msgCh: make(chan []byte, 10),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			if !closed {
				closed = true
			}
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "too slow")
			}
		},
	}

	this.addConnection(id, newConn)
	defer this.removeConnection(id)

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return errs.NewHandlerError("create connection error", err)
	}

	mu.Lock()
	if closed {
		conn.CloseNow()
		mu.Unlock()
		return errs.NewHandlerError("connection closed", nil)
	}
	c = conn
	mu.Unlock()
	defer c.CloseNow()

	ctx := c.CloseRead(context.Background())

	for {
		select {
		case msg, ok := <-newConn.msgCh:
			if !ok {
				return errs.NewHandlerError("message chan closed", nil)
			}

			if err := writeMessage(ctx, 5, msg, c); err != nil {
				return errs.NewHandlerError("write message error", err)
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

func (this *MarketStateHandler) writeToAllConns(msg []byte) {
	this.mu.Lock()
	defer this.mu.Unlock()

	for _, conn := range this.conns {
		select {
		case conn.msgCh <- msg:
		default:
			go conn.closeSlow()
		}
	}
}

func (this *MarketStateHandler) addConnection(id string, conn *Connection) {
	this.mu.Lock()
	this.conns[id] = conn
	this.mu.Unlock()
}

func (this *MarketStateHandler) removeConnection(id string) {
	this.mu.Lock()
	delete(this.conns, id)
	this.mu.Unlock()
}

func writeMessage(ctx context.Context, timeout time.Duration, msg []byte, conn *websocket.Conn) error {
	ctx, c := context.WithTimeout(ctx, time.Second*timeout)
	defer c()

	return conn.Write(ctx, websocket.MessageBinary, msg)
}
