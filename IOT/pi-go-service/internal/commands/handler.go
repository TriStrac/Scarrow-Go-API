package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"pi-go-service/internal/db"
)

func ExecuteReboot() (string, bool) {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("nohup", "sudo", "-b", "reboot")
		cmd.Run()
	}
	return "reboot scheduled", true
}

func ExecuteWifi(ssid, password string) (string, bool) {
	if runtime.GOOS != "linux" {
		return "wifi command only available on Linux", false
	}

	cmd := exec.Command("nmcli", "device", "wifi", "connect", ssid, "password", password)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("wifi failed: %v, output: %s", err, string(output)), false
	}
	return fmt.Sprintf("connected to %s", ssid), true
}

func ExecuteReset() (string, bool) {
	if err := db.ResetConfig(); err != nil {
		return fmt.Sprintf("reset failed: %v", err), false
	}

	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()

	return "reset executed - entering setup mode", true
}