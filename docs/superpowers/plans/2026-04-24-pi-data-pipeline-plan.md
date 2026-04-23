# Pi Data Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Pi accepts BLE connections from ESP32 nodes, queues data locally, sends to API with device auth, deletes on ACK.

**Architecture:** BLE GATT server → local SQLite queue → HTTP client with retry → API

**Tech Stack:** Go, TinyGo (for Pi BLE), SQLite

---

### Task 1: Implement BLE Server for Node Connections

**Files:**
- Create: `internal/ble/server.go`
- Modify: `main.go`

- [ ] **Step 1: Create BLE server for Field Mode**

Create `internal/ble/server.go`:

```go
package ble

import (
	"encoding/json"
	"fmt"
	"log"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

var (
	serviceUUID        = bluetooth.NewUUID([16]byte{0xd2, 0x71, 0x10, 0x01, 0x71, 0x01, 0x44, 0x71, 0xa7, 0x10, 0x11, 0x71, 0x0b, 0x71, 0x0c, 0x71})
	characteristicUUID = bluetooth.NewUUID([16]byte{0xd2, 0x71, 0x10, 0x02, 0x71, 0x01, 0x44, 0x71, 0xa7, 0x10, 0x11, 0x71, 0x0b, 0x71, 0x0c, 0x71})
)

// NodeData represents what a node sends when it detects something
type NodeData struct {
	NodeID           string `json:"node_id"`
	LogType         string `json:"log_type"`
	PestType        string `json:"pest_type"`
	DurationSeconds int    `json:"duration_seconds"`
	Timestamp      string `json:"timestamp"`
}

type Server struct {
	onDataReceived func(data NodeData)
	hubID       string
}

func NewServer(hubID string, onDataReceived func(data NodeData)) *Server {
	return &Server{
		hubID:       hubID,
		onDataReceived: onDataReceived,
	}
}

func (s *Server) Start() error {
	err := adapter.Enable()
	if err != nil {
		return err
	}

	adv := adapter.DefaultAdvertisement()
	err = adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    fmt.Sprintf("Scarrow_Hub_%s", s.hubID),
		ServiceUUIDs: []bluetooth.UUID{serviceUUID},
	})
	if err != nil {
		return err
	}

	err = adv.Start()
	if err != nil {
		return err
	}

	fmt.Printf("Advertising Scarrow_Hub_%s...\n", s.hubID)

	err = adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID: characteristicUUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					var data NodeData
					if err := json.Unmarshal(value, &data); err != nil {
						log.Printf("Failed to unmarshal node data: %v", err)
						return
					}
					fmt.Printf("Received from node %s: %s %s\n", data.NodeID, data.LogType, data.PestType)
					s.onDataReceived(data)
				},
			},
		},
	})

	return err
}

func (s *Server) Stop() {
	// Cleanup if needed
}
```

- [ ] **Step 2: Verify it compiles (on Pi)**

Run on Pi: `cd pi-go-service && GOOS=linux GOARCH=arm64 go build -o scarrow-hub .`

Expected: No errors

---

### Task 2: Create Local Queue for Pending Logs

**Files:**
- Create: `internal/db/queue.go`

- [ ] **Step 1: Create queue table and methods**

Create `internal/db/queue.go`:

```go
package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type PendingLog struct {
	ID              string
	NodeID          string
	LogType         string
	PestType        string
	DurationSeconds int
	Payload        string
	CreatedAt       time.Time
	RetryCount     int
}

func InitQueueDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create pending_logs table
	query := `
	CREATE TABLE IF NOT EXISTS pending_logs (
		id TEXT PRIMARY KEY,
		node_id TEXT NOT NULL,
		log_type TEXT NOT NULL,
		pest_type TEXT,
		duration_seconds INTEGER,
		payload TEXT,
		created_at TEXT NOT NULL,
		retry_count INTEGER DEFAULT 0
	);`
	_, err = db.Exec(query)
	return db, err
}

func InsertPendingLog(db *sql.DB, log *PendingLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	_, err := db.Exec(`
		INSERT INTO pending_logs (id, node_id, log_type, pest_type, duration_seconds, payload, created_at, retry_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		log.ID, log.NodeID, log.LogType, log.PestType, log.DurationSeconds, log.Payload, log.CreatedAt, log.RetryCount)
	return err
}

func GetPendingLogs(db *sql.DB, limit int) ([]PendingLog, error) {
	rows, err := db.Query(`
		SELECT id, node_id, log_type, pest_type, duration_seconds, payload, created_at, retry_count
		FROM pending_logs
		WHERE retry_count < 5
		ORDER BY created_at ASC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []PendingLog
	for rows.Next() {
		var log PendingLog
		err := rows.Scan(&log.ID, &log.NodeID, &log.LogType, &log.PestType,
			&log.DurationSeconds, &log.Payload, &log.CreatedAt, &log.RetryCount)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func DeletePendingLog(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM pending_logs WHERE id = ?", id)
	return err
}

func IncrementRetryCount(db *sql.DB, id string) error {
	_, err := db.Exec("UPDATE pending_logs SET retry_count = retry_count + 1 WHERE id = ?", id)
	return err
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd pi-go-service && go build ./...`

Expected: No errors

---

### Task 3: Implement HTTP Client to Send Logs

**Files:**
- Create: `internal/api/client.go`

- [ ] **Step 1: Create API client**

Create `internal/api/client.go`:

```go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL  string
	deviceID string
	secret  string
	client  *http.Client
}

type NodeLogRequest struct {
	NodeID           string `json:"node_id"`
	LogType         string `json:"log_type"`
	PestType        string `json:"pest_type"`
	DurationSeconds int    `json:"duration_seconds"`
	FrequencyHz     int    `json:"frequency_hz"`
	Payload        string `json:"payload"`
}

type NodeLogResponse struct {
	Message string `json:"message"`
}

func NewClient(baseURL, deviceID, secret string) *Client {
	return &Client{
		baseURL:  baseURL,
		deviceID: deviceID,
		secret:  secret,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendNodeLog(log *NodeLogRequest) error {
	jsonData, err := json.Marshal(log)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/device/%s/logs", c.baseURL, c.deviceID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-ID", c.deviceID)
	req.Header.Set("X-Device-Secret", c.secret)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("API returned status %d", resp.StatusCode)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd pi-go-service && go build ./...`

Expected: No errors

---

### Task 4: Implement Queue Processor with Retry

**Files:**
- Modify: `internal/db/queue.go` (add batch send method)
- Modify: `main.go`

- [ ] **Step 1: Add batch send to client**

In `internal/api/client.go`, add:

```go
func (c *Client) SendNodeLogs(logs []*NodeLogRequest) error {
	for _, log := range logs {
		if err := c.SendNodeLog(log); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 2: Create queue processor**

Create `internal/processor/queue.go`:

```go
package processor

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"pi-go-service/internal/api"
	"pi-go-service/internal/db"
)

type QueueProcessor struct {
	db        *sql.DB
	apiClient *api.Client
	interval time.Duration
}

func NewQueueProcessor(db *sql.DB, apiClient *api.Client, interval time.Duration) *QueueProcessor {
	return &QueueProcessor{
		db:        db,
		apiClient: apiClient,
		interval: interval,
	}
}

func (p *QueueProcessor) Start() {
	ticker := time.NewTicker(p.interval)
	go func() {
		for range ticker.C {
			p.process()
		}
	}()
	// Process immediately on start
	p.process()
}

func (p *QueueProcessor) process() {
	logs, err := db.GetPendingLogs(p.db, 10)
	if err != nil {
		log.Printf("Failed to get pending logs: %v", err)
		return
	}

	for _, log := range logs {
		nodeLog := &api.NodeLogRequest{
			NodeID:           log.NodeID,
			LogType:         log.LogType,
			PestType:        log.PestType,
			DurationSeconds: log.DurationSeconds,
			Payload:        log.Payload,
		}

		err := p.apiClient.SendNodeLog(nodeLog)
		if err != nil {
			log.Printf("Failed to send log %s: %v", log.ID, err)
			db.IncrementRetryCount(p.db, log.ID)
			continue
		}

		if err := db.DeletePendingLog(p.db, log.ID); err != nil {
			log.Printf("Failed to delete log %s: %v", log.ID, err)
		}
		fmt.Printf("Log %s sent and deleted\n", log.ID)
	}
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd pi-go-service && go build ./...`

Expected: No errors

---

### Task 5: Update main.go for Field Mode

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Wire up Field Mode**

Update `main.go`:

```go
package main

import (
	"fmt"
	"log"
	"time"

	"pi-go-service/internal/ble"
	"pi-go-service/internal/api"
	"pi-go-service/internal/db"
	"pi-go-service/internal/processor"
)

const dbPath = "scarrow.db"
const apiURL = "https://api.scarrow.io"  // or env var

func main() {
	fmt.Println("🚀 Scarrow Hub Go Service starting...")

	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	hubID, err := db.GetHubID(database)
	if err != nil {
		log.Fatalf("Failed to check hub_id: %v", err)
	}
	hubSecret, err := db.GetHubSecret(database)
	if err != nil {
		log.Fatalf("Failed to check hub_secret: %v", err)
	}

	if hubID == "" || hubSecret == "" {
		fmt.Println("⚠️ Hub configuration missing. Entering SETUP MODE...")
		// ... setup mode code
	} else {
		fmt.Printf("✅ Hub ID: %s. Starting FIELD MODE...\n", hubID)
		startFieldMode(database, hubID, hubSecret)
	}
}

func startFieldMode(database *sql.DB, hubID, hubSecret string) {
	queueDB, err := db.InitQueueDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize queue DB: %v", err)
	}
	defer queueDB.Close()

	// Create API client
	apiClient := api.NewClient(apiURL, hubID, hubSecret)

	// Create queue processor
	queueProc := processor.NewQueueProcessor(queueDB, apiClient, 10*time.Second)
	queueProc.Start()

	// Create BLE server
	bleServer := ble.NewServer(hubID, func(data ble.NodeData) {
		log := &db.PendingLog{
			NodeID:           data.NodeID,
			LogType:         data.LogType,
			PestType:        data.PestType,
			DurationSeconds: data.DurationSeconds,
			Payload:        "{}",  // Add more fields as needed
		}
		if err := db.InsertPendingLog(queueDB, log); err != nil {
			log.Printf("Failed to queue log: %v", err)
		} else {
			fmt.Printf("Queued log from node %s\n", data.NodeID)
		}
	})

	if err := bleServer.Start(); err != nil {
		log.Fatalf("Failed to start BLE server: %v", err)
	}

	fmt.Println("Field Mode running. Waiting for node data...")
	select {}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd pi-go-service && go build ./...`

Expected: No errors

---

### Task 6: Testing

**Files:**
- Manual test

- [ ] **Step 1: Start service on Pi**

Run on Pi: `./scarrow-hub`

Expected: "Starting FIELD MODE..." output

- [ ] **Step 2: Simulate node data**

Send test data to BLE characteristic, or add debug endpoint to test queue:

```go
// Add test handler in main.go for debugging
http.HandleFunc("/debug/queue", func(w http.ResponseWriter, r *http.Request) {
	// Test queue insert
})
```

- [ ] **Step 3: Verify logs sent to API**

Check API logs table, or check Pi console for success message

---

### Task 7: Commit

- [ ] **Step 1: Commit Pi changes**

```bash
git add internal/ble/server.go internal/db/queue.go internal/api/client.go internal/processor/queue.go main.go
git commit -m "feat: add field mode with BLE server, queue, and API client"
```

---

## Summary

| Task | Description | Files |
|------|------------|-------|
| 1 | BLE server for nodes | Create: `internal/ble/server.go` |
| 2 | Local queue | Create: `internal/db/queue.go` |
| 3 | API HTTP client | Create: `internal/api/client.go` |
| 4 | Queue processor | Create: `internal/processor/queue.go` |
| 5 | Wire in main.go | Modify: `main.go` |
| 6 | Manual test | - |
| 7 | Commit | - |