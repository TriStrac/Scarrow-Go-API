package ble

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

var fieldServiceUUID = bluetooth.NewUUID([16]byte{0xd2, 0x71, 0x10, 0x03, 0x71, 0x01, 0x44, 0x71, 0xa7, 0x10, 0x11, 0x71, 0x0b, 0x71, 0x0c, 0x71})
var fieldCharacteristicUUID = bluetooth.NewUUID([16]byte{0xd2, 0x71, 0x10, 0x04, 0x71, 0x01, 0x44, 0x71, 0xa7, 0x10, 0x11, 0x71, 0x0b, 0x71, 0x0c, 0x71})

func RunFieldMode(centralDeviceID string) error {
	err := adapter.Enable()
	if err != nil {
		return err
	}

	adv := adapter.DefaultAdvertisement()
	bleName := fmt.Sprintf("SCD_%s", centralDeviceID)
	err = adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    bleName,
		ServiceUUIDs: []bluetooth.UUID{fieldServiceUUID},
	})
	if err != nil {
		return err
	}

	err = adv.Start()
	if err != nil {
		return err
	}

	fmt.Printf("Advertising as %s...\n", bleName)

	for {
		time.Sleep(30 * time.Second)
	}
}