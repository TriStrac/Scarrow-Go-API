package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	bypassSecret := "dev_bypass_secret_123"

	// Test reboot command
	cmdReq := map[string]string{"cmd": "reboot"}
	cmdBytes, _ := json.Marshal(cmdReq)

	req, _ := http.NewRequest("POST", "https://scarrow-api.striel.xyz/api/hubs/PI-001/commands", bytes.NewBuffer(cmdBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Dev-Bypass", bypassSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Command response [%d]: %s\n", resp.StatusCode, string(body))
}