package detection

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

const (
	PestTypeLocust = "LOCUST"
	PestTypeRats   = "RATS"
	PestTypeBirds  = "BIRDS"
	PestTypeUnknown = "UNKNOWN"
)

type Detector struct {
	hubID      string
	windowSecs int
}

func NewDetector(hubID string, windowSecs int) *Detector {
	return &Detector{
		hubID:      hubID,
		windowSecs: windowSecs,
	}
}

func (d *Detector) WindowSeconds() int {
	return d.windowSecs
}

func (d *Detector) HubID() string {
	return d.hubID
}

func (d *Detector) AnalyzeAndSend(pestType, nodeID string, dominantFreq float64, duration int, payload string) error {
	logEntry := map[string]interface{}{
		"type":              "detection_log",
		"device_id":         d.hubID,
		"node_id":           nodeID,
		"log_type":          "PEST_DETECTED",
		"pest_type":         pestType,
		"frequency_hz":       dominantFreq,
		"duration_seconds":   duration,
		"payload":            payload,
		"timestamp":          time.Now().UTC().Format(time.RFC3339Nano),
	}

	if err := SendLog(logEntry); err != nil {
		log.Printf("Failed to send detection log: %v", err)
		return err
	}

	return nil
}

func (d *Detector) HandleSimulateCommand(args json.RawMessage) (string, bool) {
	var params struct {
		PestType string `json:"pest_type"`
		NodeID  string `json:"node_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "invalid args", false
	}

	if params.NodeID == "" {
		return "node_id is required for simulation", false
	}

	pestType, freqHz, duration := generateSimulatedDetection(params.PestType)
	payload := fmt.Sprintf(`{"simulated": true}`)

	if err := d.AnalyzeAndSend(pestType, params.NodeID, freqHz, duration, payload); err != nil {
		return fmt.Sprintf("failed to send: %v", err), false
	}

	return fmt.Sprintf("simulated %s detection on node %s", pestType, params.NodeID), true
}

func generateSimulatedDetection(pestType string) (string, float64, int) {
	var freqHz float64
	var duration int
	var classifiedType string

	switch pestType {
	case "rat":
		classifiedType = PestTypeRats
		freqHz = 15000 + rand.Float64()*10000
		duration = 2 + rand.Intn(4)
	case "bird":
		classifiedType = PestTypeBirds
		freqHz = 1000 + rand.Float64()*4000
		duration = 2 + rand.Intn(4)
	default:
		return PestTypeUnknown, 0, 0
	}

	return classifiedType, freqHz, duration
}
