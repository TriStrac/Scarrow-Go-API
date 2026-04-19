package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"pi-go-service/internal/ble"
	"pi-go-service/internal/db"
)

const dbPath = "scarrow.db"

func main() {
	fmt.Println("🚀 Scarrow Hub Go Service starting...")

	// 1. Initialize Database
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 2. Check for hub_id and secret
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
		
		err = ble.RunSetupMode(func(data ble.ProvisioningData) {
			fmt.Printf("Saving Hub ID: %s\n", data.HubID)
			err := db.SaveConfig(database, "hub_id", data.HubID)
			if err != nil {
				log.Printf("Failed to save hub_id: %v", err)
				return
			}
			err = db.SaveConfig(database, "hub_secret", data.Secret)
			if err != nil {
				log.Printf("Failed to save hub_secret: %v", err)
				return
			}
			
			fmt.Println("Provisioning complete! Restarting service in 3 seconds...")
			time.Sleep(3 * time.Second)
			os.Exit(0) // systemd will restart the service
		})
		
		if err != nil {
			log.Fatalf("Setup Mode failed: %v", err)
		}
	} else {
		fmt.Printf("✅ Hub ID: %s. Starting FIELD MODE...\n", hubID)
		startFieldMode()
	}
}

func startFieldMode() {
	fmt.Println("Scanning for Scarrow Nodes...")
	// TODO: Implement BLE Scanner logic for Field Mode
	// This will involve listening for ESP32 advertisements
	select {} // Keep running
}
