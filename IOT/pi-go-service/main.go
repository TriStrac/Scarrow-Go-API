package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"pi-go-service/internal/ble"
	"pi-go-service/internal/db"
	"pi-go-service/internal/detection"
	ws "pi-go-service/internal/ws"
)

const configPath = "scarrow.json"

func main() {
	fmt.Println("🚀 Scarrow Hub Go Service starting...")

	if err := db.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	centralDeviceID := db.GetHubID()
	centralDeviceSecret := db.GetHubSecret()
	skipFieldMode := db.GetSkipFieldMode()

	if centralDeviceID == "" || centralDeviceSecret == "" || skipFieldMode {
		fmt.Println("⚠️ Central Device configuration missing or SKIP enabled. Entering SETUP MODE...")

		err := ble.RunSetupMode(func(data ble.ProvisioningData) {
			fmt.Printf("Saving Central Device ID: %s\n", data.CentralDeviceID)
			if err := db.SaveConfig("central_device_id", data.CentralDeviceID); err != nil {
				log.Printf("Failed to save central_device_id: %v", err)
				return
			}
			if err := db.SaveConfig("central_device_secret", data.Secret); err != nil {
				log.Printf("Failed to save central_device_secret: %v", err)
				return
			}

			fmt.Println("Provisioning complete! Restarting service in 3 seconds...")
			time.Sleep(3 * time.Second)
			os.Exit(0)
		})

		if err != nil {
			log.Fatalf("Setup Mode failed: %v", err)
		}
	} else {
		fmt.Printf("✅ Central Device ID: %s. Starting FIELD MODE...\n", centralDeviceID)
		startFieldMode(centralDeviceID, centralDeviceSecret)
	}
}

func startFieldMode(centralDeviceID, centralDeviceSecret string) {
	detector := detection.NewDetector(centralDeviceID, 60)
	ws.RegisterSimulateHandler(detector)

	go ws.Reconnect(centralDeviceID)
	go ble.RunFieldMode(centralDeviceID)

	for {
		time.Sleep(60 * time.Second)
	}
}