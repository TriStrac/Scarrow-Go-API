package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	Client       *WSClient
	ServerURL    = "wss://mqtt.striel.xyz:443/ws"
	ConnectedPis = make(map[string]bool)
	mu           sync.RWMutex
)

type DetectionLogHandler interface {
	HandleDetectionLog(logID, deviceID, nodeID, logType, pestType string, freqHz float64, duration int, payload string) error
}

var detectionLogHandler DetectionLogHandler

func RegisterDetectionLogHandler(h DetectionLogHandler) {
	detectionLogHandler = h
}

type WSClient struct {
	Conn       *websocket.Conn
	done       chan struct{}
	isClosing  bool
	lastPing   time.Time
}

func NewWSClient() *WSClient {
	return &WSClient{
		done: make(chan struct{}),
	}
}

func (c *WSClient) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(ServerURL, nil)
	if err != nil {
		return fmt.Errorf("WS connection failed: %v", err)
	}
	c.Conn = conn

	conn.SetPongHandler(func(appData string) error {
		c.lastPing = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		log.Printf("Pong received")
		return nil
	})

	conn.SetPingHandler(func(appData string) error {
		log.Printf("Ping received, sending pong")
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
	})

	reg := map[string]string{"type": "register", "device": "api-server"}
	if err := conn.WriteJSON(reg); err != nil {
		return fmt.Errorf("registration failed: %v", err)
	}
	log.Println("API WebSocket connected as api-server")
	c.lastPing = time.Now()
	return nil
}

func (c *WSClient) Start() {
	c.Conn.SetReadDeadline(time.Now().Add(90 * time.Second))
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
			if c.Conn != nil && !c.isClosing {
				c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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
			if c.Conn != nil && !c.isClosing {
				if time.Since(c.lastPing) > 120*time.Second {
					log.Printf("Health check: no pong received in 120s, closing connection")
					c.Conn.Close()
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
			go Reconnect()
		}
	}()
	for {
		if c.Conn == nil {
			return
		}
		_, msg, err := c.Conn.ReadMessage()
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
	var common struct {
		Type    string `json:"type"`
		Device  string `json:"device"`
		Cmd     string `json:"cmd"`
		Success bool   `json:"success"`
		Output  string `json:"output"`
	}
	if err := json.Unmarshal(data, &common); err != nil {
		return
	}

	switch common.Type {
	case "result":
		if common.Device != "" {
			log.Printf("Command result from %s: cmd=%s success=%v output=%s",
				common.Device, common.Cmd, common.Success, common.Output)
		}
	case "detection_log":
		c.handleDetectionLog(data)
	default:
		if common.Type == "command" || common.Type == "register" {
			return
		}
	}
}

type detectionLogMsg struct {
	Type            string  `json:"type"`
	Device         string  `json:"device"`
	NodeID         string  `json:"node_id"`
	LogType        string  `json:"log_type"`
	PestType       string  `json:"pest_type"`
	FrequencyHz    float64 `json:"frequency_hz"`
	DurationSecs   int     `json:"duration_seconds"`
	Payload        string  `json:"payload"`
	Timestamp      string  `json:"timestamp"`
}

func (c *WSClient) handleDetectionLog(data []byte) {
	var msg detectionLogMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("DetectionLog: failed to unmarshal: %v", err)
		return
	}

	if msg.Device == "" {
		log.Printf("DetectionLog: missing device_id, skipping")
		return
	}
	if msg.PestType == "" {
		log.Printf("DetectionLog: missing pest_type, skipping")
		return
	}

	logID := fmt.Sprintf("dl_%s", uuid.New().String())
	if detectionLogHandler != nil {
		if err := detectionLogHandler.HandleDetectionLog(
			logID,
			msg.Device,
			msg.NodeID,
			msg.LogType,
			msg.PestType,
			msg.FrequencyHz,
			msg.DurationSecs,
			msg.Payload,
		); err != nil {
			log.Printf("DetectionLog: failed to store: %v", err)
		}
	}

	c.sendDetectionAck(msg.Device, logID)
}

func (c *WSClient) sendDetectionAck(deviceID, logID string) {
	mu.RLock()
	if Client == nil || Client.Conn == nil || Client.isClosing {
		mu.RUnlock()
		return
	}
	mu.RUnlock()

	ack := map[string]interface{}{
		"type":      "ack",
		"device":    deviceID,
		"log_id":    logID,
		"status":    "stored",
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := Client.Conn.WriteJSON(ack); err != nil {
		log.Printf("DetectionLog: failed to send ack: %v", err)
	}
}

func (c *WSClient) SendCommand(hubID, cmd string, args map[string]interface{}) error {
	mu.RLock()
	if Client == nil || Client.Conn == nil || Client.isClosing {
		mu.RUnlock()
		return fmt.Errorf("WebSocket not connected")
	}
	if c != Client {
		mu.RUnlock()
		return fmt.Errorf("Stale connection")
	}
	mu.RUnlock()

	msg := map[string]interface{}{
		"type":   "command",
		"device": hubID,
		"cmd":    cmd,
		"args":   args,
	}
	c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.Conn.WriteJSON(msg)
}

func (c *WSClient) Close() {
	c.isClosing = true
	close(c.done)
	if c.Conn != nil {
		c.Conn.Close()
	}
}

func Init() error {
	Client = NewWSClient()
	if err := Client.Connect(); err != nil {
		return err
	}
	Client.Start()
	return nil
}

func Reconnect() {
	for {
		mu.Lock()
		if Client != nil {
			Client.isClosing = true
			if Client.Conn != nil {
				Client.Conn.Close()
			}
			Client = nil
		}
		mu.Unlock()

		log.Printf("Reconnecting in 2 seconds...")
		time.Sleep(2 * time.Second)

		client := NewWSClient()
		if err := client.Connect(); err != nil {
			log.Printf("API WS reconnect failed: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}
		client.Start()

		mu.Lock()
		Client = client
		mu.Unlock()
		log.Println("API WS reconnected successfully")
		return
	}
}

func IsConnected() bool {
	mu.RLock()
	defer mu.RUnlock()
	return Client != nil && Client.Conn != nil && !Client.isClosing
}