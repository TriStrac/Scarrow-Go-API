package ble

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

var (
	serviceUUID        = bluetooth.NewUUID([16]byte{0xd2, 0x71, 0x10, 0x01, 0x71, 0x01, 0x44, 0x71, 0xa7, 0x10, 0x11, 0x71, 0x0b, 0x71, 0x0c, 0x71})
	characteristicUUID = bluetooth.NewUUID([16]byte{0xd2, 0x71, 0x10, 0x02, 0x71, 0x01, 0x44, 0x71, 0xa7, 0x10, 0x11, 0x71, 0x0b, 0x71, 0x0c, 0x71})
)

type ProvisioningData struct {
	WifiSSID     string `json:"wifi_ssid"`
	WifiPassword string `json:"wifi_password"`
	HubID        string `json:"hub_id"`
	Secret       string `json:"secret"`
}

func RunSetupMode(onSuccess func(data ProvisioningData)) error {
	err := adapter.Enable()
	if err != nil {
		return err
	}

	adv := adapter.DefaultAdvertisement()
	err = adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "Scarrow_Hub_Setup",
		ServiceUUIDs: []bluetooth.UUID{serviceUUID},
	})
	if err != nil {
		return err
	}

	err = adv.Start()
	if err != nil {
		return err
	}

	fmt.Println("Advertising Scarrow_Hub_Setup...")

	err = adapter.AddService(&bluetooth.Service{
		UUID: serviceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID: characteristicUUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					fmt.Printf("Received data: %s\n", string(value))
					var data ProvisioningData
					err := json.Unmarshal(value, &data)
					if err != nil {
						log.Printf("Failed to unmarshal JSON: %v", err)
						return
					}

					if data.WifiSSID != "" && data.HubID != "" {
						fmt.Println("Provisioning data received! Updating system...")
						err = updateWifi(data.WifiSSID, data.WifiPassword)
						if err != nil {
							log.Printf("Failed to update wifi: %v", err)
						}
						onSuccess(data)
					}
				},
			},
		},
	})

	if err != nil {
		return err
	}

	// Keep running
	select {}
}

func updateWifi(ssid, password string) error {
	fmt.Printf("Updating Wi-Fi to SSID: %s\n", ssid)
	
	// Create the wpa_supplicant entry
	config := fmt.Sprintf(`
network={
    ssid="%s"
    psk="%s"
    key_mgmt=WPA-PSK
}
`, ssid, password)

	// In a real Pi environment, we'd append to /etc/wpa_supplicant/wpa_supplicant.conf
	// or use wpa_cli. For now, we'll write to a temp file and simulate the command.
	
	f, err := os.OpenFile("/etc/wpa_supplicant/wpa_supplicant.conf", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		// Fallback for development if not running as root/on Pi
		log.Printf("Warning: Could not open /etc/wpa_supplicant/wpa_supplicant.conf: %v", err)
		return nil 
	}
	defer f.Close()

	if _, err = f.WriteString(config); err != nil {
		return err
	}

	// Trigger reconfigure
	cmd := exec.Command("wpa_cli", "-i", "wlan0", "reconfigure")
	return cmd.Run()
}
