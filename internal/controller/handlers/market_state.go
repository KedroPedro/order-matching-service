package handlers

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	jwtmanager "github.com/KedroPedro/order-matching-engine/internal/application/jwt_manager"
	"github.com/coder/websocket"
)

type MarketStateHandler struct {
	conns map[string]*Connection
	mux   http.ServeMux
	mu    sync.Mutex
}

func NewMarketStateHandler() *MarketStateHandler {
	h := &MarketStateHandler{
		conns: make(map[string]*Connection),
	}

	h.mux.HandleFunc("POST /ws/subscribe", h.subscribeHandler)

	return h
}

type Connection struct {
	conn      *websocket.Conn
	id        string
	msgCh     chan []byte
	closeSlow func()
}

func (this *MarketStateHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	this.mux.ServeHTTP(w, r)
}

func (this *MarketStateHandler) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return //TODO: add logging
	}

	id, err := getCookieValue(cookie)
	if err != nil {
		return
	}

	if err := this.subscribe(id, w, r); err != nil {
		return
	}

	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}

	if err != nil {
		//add logging
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
		return err //TODO: add logging
	}

	mu.Lock()
	if closed {
		conn.CloseNow()
		mu.Unlock()
		return errors.New("connection closed")
	}
	c = conn
	mu.Unlock()
	defer c.CloseNow()

	ctx := c.CloseRead(context.Background())

	for {
		select {
		case msg, ok := <-newConn.msgCh:
			if !ok {
				return errors.New("msg chan closed")
			}

			if err := writeMessage(ctx, 5, msg, c); err != nil {
				return err
			}

		case <-ctx.Done():
			return ctx.Err()
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

func getCookieValue(cookie *http.Cookie) (string, error) {
	if cookie.Expires.After(time.Now()) {
		return " ", errors.New("cookies expired") //TODO: add logging
	}

	value, err := jwtmanager.Decode(cookie.Value)
	if err != nil {
		return "", err //TODO: add logging
	}

	return value, nil
}

func writeMessage(ctx context.Context, timeout time.Duration, msg []byte, conn *websocket.Conn) error {
	ctx, c := context.WithTimeout(ctx, time.Second*timeout)
	defer c()

	return conn.Write(ctx, websocket.MessageBinary, msg)
}
