package detection

import (
	"errors"
	"log"

	"pi-go-service/internal/ws"
)

var errWSClientNotInit = errors.New("WebSocket client not initialized")

func SendLog(logData map[string]interface{}) error {
	if ws.Client == nil {
		return errWSClientNotInit
	}
	if err := ws.Client.SendInfo(logData); err != nil {
		return err
	}
	log.Printf("Detection log sent: %s", logData["pest_type"])
	return nil
}
