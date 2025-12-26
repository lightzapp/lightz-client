package lightz

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lightzapp/lightz-client/internal/logger"
	"github.com/mitchellh/mapstructure"
)

const reconnectInterval = 15 * time.Second
const pingInterval = 30 * time.Second
const pongWait = 5 * time.Second
const readBufferSize = 1024 * 1024 // 1MB
const updatesChannelBuffer = 50

type SwapUpdate struct {
	SwapStatusResponse `mapstructure:",squash"`
	Id                 string `json:"id"`
}

type Websocket struct {
	Updates     chan SwapUpdate
	updatesLock sync.Mutex

	apiUrl            string
	subscriptions     chan bool
	conn              *websocket.Conn
	connLock          sync.RWMutex
	closed            bool
	dialer            *websocket.Dialer
	swapIds           []string
	swapIdsLock       sync.Mutex
	reconnectInterval time.Duration
}

type wsResponse struct {
	Event   string `json:"event"`
	Error   string `json:"error"`
	Channel string `json:"channel"`
	Args    []any  `json:"args"`
}

func (lightz *Api) NewWebsocket() *Websocket {
	httpTransport, ok := lightz.Client.Transport.(*http.Transport)

	dialer := *websocket.DefaultDialer
	if ok {
		dialer.Proxy = httpTransport.Proxy
	}
	dialer.ReadBufferSize = readBufferSize

	return &Websocket{
		apiUrl:            lightz.URL,
		subscriptions:     make(chan bool),
		dialer:            &dialer,
		Updates:           make(chan SwapUpdate, updatesChannelBuffer),
		reconnectInterval: reconnectInterval,
	}
}

func (lightz *Websocket) Connect() error {
	if lightz.closed {
		return errors.New("websocket is closed")
	}
	wsUrl, err := url.Parse(lightz.apiUrl)
	if err != nil {
		return err
	}
	wsUrl.Path += "/v2/ws"

	switch wsUrl.Scheme {
	case "https":
		wsUrl.Scheme = "wss"
	case "http":
		wsUrl.Scheme = "ws"
	}

	conn, _, err := lightz.dialer.Dial(wsUrl.String(), nil)
	if err != nil {
		return fmt.Errorf("could not connect to lightzd ws at %s: %w", wsUrl, err)
	}

	lightz.setConn(conn)

	logger.Infof("Connected to Lightz ws at %s", wsUrl)

	setDeadline := func() error {
		return conn.SetReadDeadline(time.Now().Add(pingInterval + pongWait))
	}
	_ = setDeadline()
	conn.SetPongHandler(func(string) error {
		logger.Silly("Received pong")
		return setDeadline()
	})
	pingTicker := time.NewTicker(pingInterval)

	go func() {
		defer pingTicker.Stop()
		for range pingTicker.C {
			// Will not wait longer with writing than for the response
			err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(pongWait))
			if err != nil {
				if lightz.closed {
					return
				}
				logger.Errorf("could not send ping: %s", err)
				return
			}
		}
	}()

	go func() {
		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				// if `close` was intentionally called, `Connected` will return false
				// since the connection has already been set to nil.
				if lightz.Connected() {
					lightz.setConn(nil)
					logger.Error("could not receive message: " + err.Error())
				} else {
					return
				}
				break
			}

			logger.Silly("Received websocket message: " + string(message))

			switch msgType {
			case websocket.TextMessage:
				if err := lightz.handleTextMessage(message); err != nil {
					logger.Errorf("could not handle websocket message: %s", err)
				}
			default:
				logger.Warnf("unknown message type: %v", msgType)
			}
		}
		pingTicker.Stop()
		for {
			logger.Errorf("lost connection to lightzd ws, reconnecting in %s", lightz.reconnectInterval)
			time.Sleep(lightz.reconnectInterval)
			err := lightz.Connect()
			if err == nil {
				return
			}
		}
	}()

	lightz.swapIdsLock.Lock()
	swapIds := make([]string, len(lightz.swapIds))
	copy(swapIds, lightz.swapIds)
	lightz.swapIdsLock.Unlock()

	if len(swapIds) > 0 {
		if err := lightz.subscribe(swapIds); err != nil {
			return fmt.Errorf("failed to subscribe to existing swaps: %w", err)
		}
	}

	return nil
}

func (lightz *Websocket) handleTextMessage(data []byte) error {
	lightz.updatesLock.Lock()
	defer lightz.updatesLock.Unlock()
	var response wsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("invalid json: %s", err)
	}
	if response.Error != "" {
		return fmt.Errorf("lightzd error: %s", response.Error)
	}

	switch response.Event {
	case "update":
		switch response.Channel {
		case "swap.update":
			for _, arg := range response.Args {
				var update SwapUpdate
				if err := mapstructure.Decode(arg, &update); err != nil {
					return fmt.Errorf("invalid lightzd response: %v", err)
				}
				lightz.Updates <- update
			}
		default:
			logger.Warnf("unknown update channel: %s", response.Channel)
		}
	case "subscribe":
		lightz.subscriptions <- true
	default:
		logger.Warnf("unknown ws event: %s", response.Event)
	}
	return nil
}

func (lightz *Websocket) writeJson(data any) error {
	lightz.connLock.Lock()
	defer lightz.connLock.Unlock()
	if lightz.conn == nil {
		return errors.New("websocket is not connected")
	}
	return lightz.conn.WriteJSON(data)
}

func (lightz *Websocket) subscribe(swapIds []string) error {
	if lightz.closed {
		return errors.New("websocket is closed")
	}
	logger.Infof("Subscribing to Swaps: %v", swapIds)
	if len(swapIds) == 0 {
		return nil
	}
	if err := lightz.writeJson(map[string]any{
		"op":      "subscribe",
		"channel": "swap.update",
		"args":    swapIds,
	}); err != nil {
		return err
	}
	select {
	case <-lightz.subscriptions:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("no answer from lightzd")
	}
}

func (lightz *Websocket) Subscribe(swapIds []string) error {
	if len(swapIds) == 0 {
		return nil
	}
	if !lightz.Connected() {
		if err := lightz.Connect(); err != nil {
			return fmt.Errorf("could not connect lightzd ws: %w", err)
		}
	}
	if err := lightz.subscribe(swapIds); err != nil {
		// the connection might be dead, so forcefully reconnect
		if err := lightz.Reconnect(); err != nil {
			return fmt.Errorf("could not reconnect lightzd ws: %w", err)
		}
		if err := lightz.subscribe(swapIds); err != nil {
			return err
		}
	}
	lightz.swapIdsLock.Lock()
	lightz.swapIds = append(lightz.swapIds, swapIds...)
	lightz.swapIdsLock.Unlock()
	return nil
}

func (lightz *Websocket) Unsubscribe(swapId string) {
	lightz.swapIdsLock.Lock()
	defer lightz.swapIdsLock.Unlock()
	lightz.swapIds = slices.DeleteFunc(lightz.swapIds, func(id string) bool {
		return id == swapId
	})
	logger.Debugf("Unsubscribed from swap %s", swapId)
	if len(lightz.swapIds) == 0 {
		logger.Debugf("No more pending swaps, disconnecting websocket")
		if err := lightz.close(); err != nil {
			logger.Warnf("could not close lightzd ws: %v", err)
		}
	}
}

func (lightz *Websocket) close() error {
	lightz.connLock.Lock()
	defer lightz.connLock.Unlock()

	if conn := lightz.conn; conn != nil {
		lightz.conn = nil
		return conn.Close()
	}
	return nil
}

func (lightz *Websocket) Close() error {
	// setting this flag will cause the `Updates` channel to be closed
	// in the receiving goroutine. this isn't done here to avoid a situation
	// where we close the channel while the receiving routine is processing an incoming message
	// and then tries to send on the closed channel.
	lightz.updatesLock.Lock()
	defer lightz.updatesLock.Unlock()
	if lightz.closed {
		return nil
	}
	close(lightz.Updates)
	lightz.closed = true
	return lightz.close()
}

func (lightz *Websocket) Connected() bool {
	lightz.connLock.RLock()
	defer lightz.connLock.RUnlock()
	return lightz.conn != nil
}

func (lightz *Websocket) Reconnect() error {
	logger.Infof("Force reconnecting to Lightz ws")
	if err := lightz.close(); err != nil {
		logger.Warnf("could not close lightzd ws: %v", err)
	}
	return lightz.Connect()
}

func (lightz *Websocket) setConn(conn *websocket.Conn) {
	lightz.connLock.Lock()
	lightz.conn = conn
	lightz.connLock.Unlock()
}
