# Wi-Fi Safe Update Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Wi-Fi backup/restore logic to prevent Pi from losing network on invalid credentials.

**Architecture:** Before applying new Wi-Fi, backup existing config. Try new Wi-Fi, wait 15s, verify. If fails, restore backup and reboot.

**Tech Stack:** Go, NetworkManager (nmcli), Raspberry Pi OS Bookworm

---

### Task 1: Extract updateWifi to Shared Package

**Files:**
- Create: `internal/wifi/manager.go` (new shared Wi-Fi manager)
- Modify: `internal/ble/setup.go` - use shared package
- Modify: `internal/mqtt/subscriber.go` - use shared package

- [ ] **Step 1: Create internal/wifi/manager.go**

```go
package wifi

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	backupPath     = "/tmp/wifi-backup.nmconnection"
	connectionDir = "/etc/NetworkManager/system-connections"
)

func ManageWifi(ssid, password string) error {
	fmt.Printf("Managing Wi-Fi connection to: %s\n", ssid)
	
	// Step 1: Backup existing config
	if err := backup(); err != nil {
		fmt.Printf("Warning: backup failed: %v\n", err)
	}
	
	// Step 2: Apply new config
	if err := apply(ssid, password); err != nil {
		return fmt.Errorf("failed to apply Wi-Fi config: %v", err)
	}
	
	// Step 3: Verify connection (15 seconds)
	if !verify(ssid) {
		fmt.Println("Wi-Fi connection failed, restoring backup...")
		restore()
		return fmt.Errorf("connection failed, restored backup and rebooting")
	}
	
	// Success - clean up backup
	cleanup()
	fmt.Println("Wi-Fi connected successfully!")
	return nil
}

func backup() error {
	// Find existing Wi-Fi connection file
	cmd := exec.Command("bash", "-c", "ls /etc/NetworkManager/system-connections/*.nmconnection 2>/dev/null | head -1")
	out, err := cmd.Output()
	if err != nil {
		return nil // No existing config
	}
	
	src := strings.TrimSpace(string(out))
	if src == "" {
		return nil
	}
	
	// Copy to backup
	cp := exec.Command("sudo", "cp", src, backupPath)
	if err := cp.Run(); err != nil {
		return err
	}
	chmod := exec.Command("sudo", "chmod", "600", backupPath)
	chmod.Run()
	
	fmt.Printf("Backed up Wi-Fi config to %s\n", backupPath)
	return nil
}

func apply(ssid, password string) error {
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
	
	path := fmt.Sprintf("%s/%s.nmconnection", connectionDir, ssid)
	if err := os.WriteFile(path, []byte(config), 0600); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}
	
	chmod := exec.Command("sudo", "chmod", "600", path)
	chmod.Run()
	
	reload := exec.Command("sudo", "nmcli", "connection", "reload")
	reload.Run()
	
	return nil
}

func verify(ssid string) bool {
	fmt.Println("Waiting 15 seconds for Wi-Fi connection...")
	time.Sleep(15 * time.Second)
	
	cmd := exec.Command("nmcli", "-t", "-f", "SSID", "dev", "wifi")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == ssid {
			return true
		}
	}
	
	return false
}

func restore() error {
	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		fmt.Println("No backup to restore")
		return nil
	}
	
	// Copy backup back
	cmd := exec.Command("sudo", "cp", backupPath, connectionDir+"/")
	if err := cmd.Run(); err != nil {
		return err
	}
	
	reload := exec.Command("sudo", "nmcli", "connection", "reload")
	reload.Run()
	
	fmt.Println("Restored Wi-Fi backup, rebooting...")
	exec.Command("sudo", "reboot").Run()
	select {}
	
	return nil
}

func cleanup() {
	if _, err := os.Stat(backupPath); err == nil {
		os.Remove(backupPath)
		fmt.Println("Cleaned up Wi-Fi backup")
	}
}

func generateUUID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d-%d-%d-%d-%d",
		rand.Intn(10000), rand.Intn(10000), rand.Intn(10000),
		rand.Intn(10000), rand.Intn(10000))
}
```

- [ ] **Step 2: Update internal/ble/setup.go to use wifi.ManageWifi**

In setup.go, add import and replace updateWifi call:

```go
import (
	// ... existing
	"pi-go-service/internal/wifi"
)
```

Replace the updateWifi function call section from lines 87-94:

```go
if data.WifiSSID != "" && data.HubID != "" {
	fmt.Println("Provisioning data received! Updating system...")
	err = wifi.ManageWifi(data.WifiSSID, data.WifiPassword)
	if err != nil {
		log.Printf("Failed to update wifi: %v", err)
	}
	onSuccess(data)
}
```

- [ ] **Step 3: Update internal/mqtt/subscriber.go to use wifi.ManageWifi**

Add import and replace wifichange handler:

```go
import (
	// ... existing
	"pi-go-service/internal/wifi"
)
```

Replace the wifichange case (around line 83):

```go
case "wifichange":
	if cmd.WifiSSID != "" {
		err := wifi.ManageWifi(cmd.WifiSSID, cmd.WifiPassword)
		if err != nil {
			log.Printf("Wi-Fi update failed: %v", err)
		}
	}
```

- [ ] **Step 4: Build and verify**

```bash
cd D:\Codes\Scarrow-Go-API\.worktrees\pi-data-pipeline\IOT\pi-go-service && go build ./...
```

Expected: No errors

- [ ] **Step 5: Deploy to Pi**

```bash
cd D:\Codes\Scarrow-Go-API\.worktrees\pi-data-pipeline\IOT && python deploy_to_pi.py
```

Expected: Success

- [ ] **Step 6: Commit**

```bash
git add internal/wifi/ internal/ble/setup.go internal/mqtt/subscriber.go
git commit -m "feat: add Wi-Fi backup and restore for safe updates"
```

---

## Self-Review

**1. Spec coverage:**
- ✅ Backup existing config before update
- ✅ Apply new Wi-Fi config
- ✅ Verify connection (15s wait)
- ✅ Restore backup on failure
- ✅ Auto-reboot after restore
- ✅ Both entry points (BLE + MQTT)

**2. Placeholder scan:** No TODOs or TBDs found.

**3. Type consistency:** Single shared wifi package used by both entry points - consistent.

---

## Execution Options

**1. Subagent-Driven (recommended)** - I dispatch subagents per task, fast iteration

**2. Inline Execution** - Execute tasks in this session

**Which approach?**