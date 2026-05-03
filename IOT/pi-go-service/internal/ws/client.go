package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"pi-go-service/internal/commands"

	"github.com/gorilla/websocket"
)

type SimulateHandler interface {
	HandleSimulateCommand(args json.RawMessage) (string, bool)
}

var simulateHandler SimulateHandler

func RegisterSimulateHandler(h SimulateHandler) {
	simulateHandler = h
}

var (
	Client    *WSClient
	ServerURL = "wss://mqtt.striel.xyz:443/ws"
	mu        sync.RWMutex
)

type WSClient struct {
	conn      *websocket.Conn
	done      chan struct{}
	hubID     string
	isClosing bool
	lastPing  time.Time
}

func NewWSClient(hubID string) *WSClient {
	return &WSClient{
		done:  make(chan struct{}),
		hubID: hubID,
	}
}

func (c *WSClient) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(ServerURL, nil)
	if err != nil {
		return fmt.Errorf("WS connection failed: %v", err)
	}
	c.conn = conn

	conn.SetPongHandler(func(appData string) error {
		c.lastPing = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		log.Printf("Pong received")
		return nil
	})

	conn.SetPingHandler(func(appData string) error {
		log.Printf("Ping received, sending pong")
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
	})

	reg := map[string]string{"type": "register", "device": c.hubID}
	if err := conn.WriteJSON(reg); err != nil {
		return fmt.Errorf("registration failed: %v", err)
	}
	log.Printf("WebSocket connected and registered as %s", c.hubID)
	c.lastPing = time.Now()
	return nil
}

func (c *WSClient) Start() {
	c.conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	go c.readLoop()
	go c.pingLoop()
	go c.healthCheck()
}

func (c *WSClient) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			if c.conn != nil && !c.isClosing {
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("Ping failed: %v", err)
					return
				}
				log.Printf("Ping sent")
			}
		}
	}
}

func (c *WSClient) healthCheck() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			if c.conn != nil && !c.isClosing {
				if time.Since(c.lastPing) > 120*time.Second {
					log.Printf("Health check: no pong received in 120s, closing connection")
					c.conn.Close()
					return
				}
				log.Printf("Health check: connection alive, last pong %v ago", time.Since(c.lastPing))
			}
		}
	}
}

func (c *WSClient) readLoop() {
	defer func() {
		if !c.isClosing {
			close(c.done)
			log.Printf("WS read loop ended, triggering reconnect")
			go Reconnect(c.hubID)
		}
	}()
	for {
		if c.conn == nil {
			return
		}
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if !c.isClosing {
				log.Printf("WS read error: %v", err)
			}
			return
		}
		c.handleMessage(msg)
	}
}

func (c *WSClient) handleMessage(data []byte) {
	var msg struct {
		Type   string          `json:"type"`
		Device string          `json:"device"`
		Cmd    string          `json:"cmd"`
		Args   json.RawMessage `json:"args"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	if msg.Device != c.hubID && msg.Device != "all" {
		return
	}

	if msg.Type == "command" {
		c.executeCommand(msg.Cmd, msg.Args)
	}
}

func (c *WSClient) executeCommand(cmd string, args json.RawMessage) {
	log.Printf("Executing command: %s", cmd)

	var success bool
	var output string

	switch cmd {
	case "reboot":
		output, success = commands.ExecuteReboot()
	case "wifi":
		var wifiArgs struct {
			SSID string `json:"ssid"`
			Pass string `json:"password"`
		}
		if err := json.Unmarshal(args, &wifiArgs); err == nil {
			output, success = commands.ExecuteWifi(wifiArgs.SSID, wifiArgs.Pass)
		} else {
			output, success = "invalid args", false
		}
	case "reset":
		output, success = commands.ExecuteReset()
	case "simulate_detection":
		if simulateHandler != nil {
			output, success = simulateHandler.HandleSimulateCommand(args)
		} else {
			output, success = "simulate handler not registered", false
		}
	default:
		output, success = fmt.Sprintf("unknown command: %s", cmd), false
	}

	result := map[string]interface{}{
		"type":    "result",
		"device":  c.hubID,
		"cmd":     cmd,
		"success": success,
		"output":  output,
	}
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	c.conn.WriteJSON(result)
}

func (c *WSClient) SendInfo(info map[string]interface{}) error {
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	info["type"] = "info"
	info["device_id"] = c.hubID
	return c.conn.WriteJSON(info)
}

func (c *WSClient) Close() {
	c.isClosing = true
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
}

func Reconnect(hubID string) {
	for {
		mu.Lock()
		if Client != nil {
			Client.isClosing = true
			if Client.conn != nil {
				Client.conn.Close()
			}
			Client = nil
		}
		mu.Unlock()

		log.Printf("Reconnecting in 2 seconds...")
		time.Sleep(2 * time.Second)

		client := NewWSClient(hubID)
		if err := client.Connect(); err != nil {
			log.Printf("WS reconnect failed: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}
		client.Start()

		mu.Lock()
		Client = client
		mu.Unlock()
		log.Printf("WS reconnected successfully as %s", hubID)
		return
	}
}