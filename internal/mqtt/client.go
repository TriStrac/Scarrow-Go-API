package mqtt

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	client mqtt.Client
	broker string
}

var GlobalClient *Client

func Init(broker string) {
	GlobalClient = NewClient(broker)
}

func NewClient(broker string) *Client {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID("scarrow-api").
		SetCleanSession(true).
		SetConnectTimeout(10 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			fmt.Printf("Connected to MQTT broker\n")
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("MQTT connection error: %v\n", token.Error())
	}

	return &Client{client: client, broker: broker}
}

func (c *Client) PublishCommand(hubID string, payload map[string]interface{}) error {
	topic := fmt.Sprintf("hub/%s/commands", hubID)

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	token := c.client.Publish(topic, 0, false, jsonBytes)
	if token.WaitTimeout(5*time.Second) && token.Error() != nil {
		return token.Error()
	}

	fmt.Printf("Published command to %s: %s\n", topic, jsonBytes)
	return nil
}

func (c *Client) Close() {
	c.client.Disconnect(250)
}