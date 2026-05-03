# MQTT Command Control Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add ability to send commands (reboot, wifi change, unpair) to Pi hubs via MQTT from the API.

**Architecture:** API receives command via HTTP endpoint, publishes to MQTT topic. Pi subscribes to `hub/{hub_id}/commands` and executes locally.

**Tech Stack:** Go (paho.mqtt.v5), Mosquitto broker, Gin framework

---

### Task 1: Create Mosquitto Docker Setup

**Files:**
- Create: `D:\Codes\mosquitto-mqtt\docker-compose.yml`
- Create: `D:\Codes\mosquitto-mqtt\config\mosquitto.conf`

- [ ] **Step 1: Create folder structure**

```bash
mkdir -p D:\Codes\mosquitto-mqtt\config
mkdir -p D:\Codes\mosquitto-mqtt\data
mkdir -p D:\Codes\mosquitto-mqtt\log
```

- [ ] **Step 2: Create docker-compose.yml**

```yaml
version: "3.8"

services:
  mosquitto:
    image: eclipse-mosquitto:2.0.22
    container_name: mosquitto-mqtt
    restart: unless-stopped
    ports:
      - "1883:1883"
    volumes:
      - ./config:/mosquitto/config:ro
      - ./data:/mosquitto/data
      - ./log:/mosquitto/log
    command: ["mosquitto", "-c", "/mosquitto/config/mosquitto.conf"]
    networks:
      - mosquitto_network

networks:
  mosquitto_network:
    driver: bridge
```

- [ ] **Step 3: Create mosquitto.conf**

```
listener 1883
allow_anonymous true
persistence true
persistence_location /mosquitto/data/
log_dest file /mosquitto/log/mosquitto.log
log_type error
log_type warning
log_type notice
log_type information
```

- [ ] **Step 4: Start Mosquitto**

```bash
cd D:\Codes\mosquitto-mqtt && docker-compose up -d
```

- [ ] **Step 5: Test connection**

```bash
docker exec mosquitto-mqtt mosquitto_sub -t "test/topic" &
docker exec mosquitto-mqtt mosquitto_pub -t "test/topic" -m "hello"
```

---

### Task 2: Add MQTT Client to API

**Files:**
- Create: `D:\Codes\Scarrow-Go-API\internal\mqtt\client.go`
- Modify: `D:\Codes\Scarrow-Go-API\go.mod` - add paho.mqtt.v5 dependency
- Modify: `D:\Codes\Scarrow-Go-API\internal\api\routes\device.go` - add command endpoint

- [ ] **Step 1: Add paho.mqtt to go.mod**

```bash
cd D:\Codes\Scarrow-Go-API && go get github.com/eclipse/paho.mqtt.golang/v2
```

- [ ] **Step 2: Create mqtt/client.go**

```go
package mqtt

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang/v2"
)

type Client struct {
	client mqtt.Client
	broker string
}

type CommandPayload struct {
	Cmd           string `json:"cmd"`
	WifiSSID      string `json:"wifi_ssid,omitempty"`
	WifiPassword  string `json:"wifi_password,omitempty"`
}

func NewClient(broker string) *Client {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID("scarrow-api").
		SetCleanSession(true).
		SetConnectTimeout(10 * time.Second).
		SetAutoReconnection(true).
		SetReconnectingHandler(func(c mqtt.Client, o *mqtt.ClientOptions) {
			fmt.Printf("Reconnecting to MQTT broker...\n")
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("MQTT connection error: %v\n", token.Error())
	}

	return &Client{client: client, broker: broker}
}

func (c *Client) PublishCommand(hubID string, payload CommandPayload) error {
	topic := fmt.Sprintf("hub/%s/commands", hubID)
	msg, err := payload.MarshalJSON()
	if err != nil {
		return err
	}

	token := c.client.Publish(topic, 0, false, msg)
	if token.WaitTimeout(5 * time.Second) && token.Error() != nil {
		return token.Error()
	}

	fmt.Printf("Published command to %s: %s\n", topic, msg)
	return nil
}

func (c *Client) Close() {
	c.client.Disconnect(250)
}
```

Wait - CommandPayload needs MarshalJSON. Better use map instead:

```go
func (c *Client) PublishCommand(hubID string, payload map[string]interface{}) error {
	topic := fmt.Sprintf("hub/%s/commands", hubID)
	
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	token := c.client.Publish(topic, 0, false, jsonBytes)
	if token.WaitTimeout(5 * time.Second) && token.Error() != nil {
		return token.Error()
	}

	fmt.Printf("Published command to %s: %s\n", topic, jsonBytes)
	return nil
}
```

Add `"encoding/json"` to imports.

- [ ] **Step 3: Add MQTT client to cmd/api/main.go or create mqtt init**

Actually - better to create MQTT client once in main.go and pass to routes. But minimal change - create global client in mqtt package.

```go
var GlobalClient *Client

func Init(broker string) {
	GlobalClient = NewClient(broker)
}
```

Add to mqtt/client.go:

```go
var GlobalClient *Client

func Init(broker string) {
	GlobalClient = NewClient(broker)
}
```

- [ ] **Step 4: Add command endpoint to routes/device.go**

Update imports and add:

```go
// In RegisterDeviceRoutes, add new route group:
hubsRoutes.POST("/:hubId/commands", deviceController.SendCommand)
```

Add to deviceController:

```go
type SendCommandReq struct {
	Cmd          string `json:"cmd" binding:"required,oneof=reboot wifichange unpair"`
	WifiSSID     string `json:"wifi_ssid"`
	WifiPassword string `json:"wifi_password"`
}

func (c *DeviceController) SendCommand(ctx *gin.Context) {
	hubID := ctx.Param("hubId")
	
	// Verify ownership
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	isOwner, err := c.deviceService.IsOwner(hubID, callerID.(string))
	if err != nil || !isOwner {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var req SendCommandReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build MQTT payload
	payload := map[string]interface{}{
		"cmd": req.Cmd,
	}
	if req.Cmd == "wifichange" {
		payload["wifi_ssid"] = req.WifiSSID
		payload["wifi_password"] = req.WifiPassword
	}

	// Publish to MQTT
	if mqtt.GlobalClient != nil {
		err = mqtt.GlobalClient.PublishCommand(hubID, payload)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish command: " + err.Error()})
			return
		}
	} else {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "MQTT not available"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Command sent"})
}
```

- [ ] **Step 5: Initialize MQTT in main.go**

In `cmd/api/main.go` after config load:

```go
import (
	"github.com/TriStrac/Scarrow-Go-API/internal/mqtt"
)

func main() {
	// ... existing init
	
	// Init MQTT
	mqtt.Init("tcp://localhost:1883")
	
	router := gin.New()
	// ...
}
```

- [ ] **Step 6: Build and verify**

```bash
cd D:\Codes\Scarrow-Go-API && go build ./...
```

Expected: No errors

- [ ] **Step 7: Commit**

```bash
git add internal/mqtt/go.mod go.mod cmd/api/main.go internal/api/controllers/device.go internal/api/routes/device.go
git commit -m "feat: add MQTT client for hub command control"
```

---

### Task 3: Add MQTT Subscriber to Pi Service

**Files:**
- Create: `D:\Codes\Scarrow-Go-API\.worktrees\pi-data-pipeline\IOT\pi-go-service\internal\mqtt\subscriber.go`
- Modify: `D:\Codes\Scarrow-Go-API\.worktrees\pi-data-pipeline\IOT\pi-go-service\main.go` - add MQTT subscriber in Field Mode

- [ ] **Step 1: Create mqtt/subscriber.go**

```go
package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang/v2"
)

type Subscriber struct {
	client    mqtt.Client
	hubID     string
	 broker   string
	onCommand func(cmd string, data map[string]string)
}

type CommandMessage struct {
	Cmd           string `json:"cmd"`
	WifiSSID      string `json:"wifi_ssid"`
	WifiPassword  string `json:"wifi_password"`
}

func NewSubscriber(hubID, broker string, onCommand func(cmd string, data map[string]string)) *Subscriber {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(fmt.Sprintf("hub-%s", hubID)).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectTimeout(10 * time.Second).
		SetReconnectingHandler(func(c mqtt.Client, o *mqtt.ClientOptions) {
			fmt.Printf("Reconnecting to MQTT broker...\n")
		})
	
	client := mqtt.NewClient(opts)
	
	return &Subscriber{
		client:    client,
		hubID:    hubID,
		broker:   broker,
		onCommand: onCommand,
	}
}

func (s *Subscriber) Start() error {
	topic := fmt.Sprintf("hub/%s/commands", s.hubID)
	
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	
	if token := s.client.Subscribe(topic, 0, s.handleMessage); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	
	fmt.Printf("Subscribed to %s\n", topic)
	return nil
}

func (s *Subscriber) handleMessage(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received command: %s\n", string(msg.Payload()))
	
	var cmd CommandMessage
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		log.Printf("Failed to parse command: %v", err)
		return
	}

	switch cmd.Cmd {
	case "reboot":
		fmt.Println("Rebooting...")
		exec.Command("sudo", "reboot").Run()
		select {}
	case "unpair":
		fmt.Println("Resetting config and rebooting...")
		os.Remove("/home/pi/scarrow.db")
		exec.Command("sudo", "reboot").Run()
		select {}
	case "wifichange":
		if cmd.WifiSSID != "" {
			updateWifi(cmd.WifiSSID, cmd.WifiPassword)
		}
	default:
		log.Printf("Unknown command: %s", cmd.Cmd)
	}
}

func updateWifi(ssid, password string) error {
	fmt.Printf("Connecting to Wi-Fi SSID: %s\n", ssid)
	
	// Generate UUID for connection
	uuid := generateUUID()
	
	config := fmt.Sprintf(`[connection]
id=%s
uuid=%s
type=wifi
autoconnect=true
interface-name=wlan0
method=auto

[wifi]
mode=infrastructure
ssid=%s

[wifi-security]
key-mgmt=wpa-psk
psk=%s

[ipv4]
method=auto

[ipv6]
method=auto
`, ssid, uuid, ssid, password)

	path := fmt.Sprintf("/etc/NetworkManager/system-connections/%s.nmconnection", ssid)
	err := os.WriteFile(path, []byte(config), 0600)
	if err != nil {
		return fmt.Errorf("failed to write NM config: %v", err)
	}
	
	cmd := exec.Command("sudo", "chmod", "600", path)
	cmd.Run()
	
	reloadCmd := exec.Command("sudo", "nmcli", "connection", "reload")
	reloadCmd.Run()
	
	fmt.Println("Wi-Fi config saved!")
	return nil
}

func generateUUID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d-%d-%d-%d-%d", 
		rand.Intn(10000), rand.Intn(10000), rand.Intn(10000), 
		rand.Intn(10000), rand.Intn(10000))
}
```

Need to import `math/rand`:

```go
import (
	"fmt"
	"math/rand"
	"time"
	// ...
)
```

- [ ] **Step 2: Add MQTT subscriber to main.go Field Mode**

Add to imports:
```go
import (
	// ... existing
	"pi-go-service/internal/mqtt"
)
```

In startFieldMode, add after BLE server setup:

```go
// MQTT subscriber for commands
mqttBroker := "tcp://mosquitto-mqtt.striel.xyz:1883"
mqttSub := mqtt.NewSubscriber(hubID, mqttBroker, nil)
if err := mqttSub.Start(); err != nil {
	log.Printf("Warning: MQTT connection failed: %v", err)
} else {
	fmt.Println("MQTT subscriber connected")
}
```

Actually - NewSubscriber takes onCommand callback. But we can just pass nil and use the internal handler.

- [ ] **Step 3: Verify it compiles**

```bash
cd D:\Codes\Scarrow-Go-API\.worktrees\pi-data-pipeline\IOT\pi-go-service && go build ./...
```

Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/mqtt/
git commit -m "feat: add MQTT subscriber for remote commands"
```

---

## Self-Review

**1. Spec coverage:**
- ✅ Reboot command - Task 2 & 3
- ✅ WiFi change - Task 2 & 3
- ✅ Unpair (reset config) - Task 2 & 3
- ✅ API endpoint - Task 2
- ✅ MQTT broker setup - Task 1
- ✅ Pi subscriber - Task 3

**2. Placeholder scan:** No TODOs or TBDs found.

**3. Type consistency:** 
- CommandPayload in Task 2 uses map[string]interface{} - matches CommandMessage in Task 3
- Topic format consistent: hub/{hub_id}/commands

---

## Execution Options

**1. Subagent-Driven (recommended)** - I dispatch subagents per task, each task has its own commit and review

**2. Inline Execution** - Execute tasks in this session with checkpoints

**Which approach?**