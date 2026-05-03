# Wi-Fi Safe Update Design

## Problem

Pi loses network access if new Wi-Fi credentials are invalid or unreachable.

## Solution

Backup existing config, try new Wi-Fi, auto-restore on failure.

## Architecture

### Components

| Component | Behavior |
|-----------|-----------|
| Backup | Copy existing Wi-Fi NM config to `/tmp/wifi-backup.nmconnection` |
| Verify Connection | Wait 15s, check `nmcli dev wifi` for connected status |
| Success | Keep new config, delete backup |
| Failure | Restore backup from `/tmp/`, reload NM, auto-reboot |

### Entry Points

- **BLE Setup Mode** (`internal/ble/setup.go`): Initial provisioning
- **MQTT wifichange** (`internal/mqtt/subscriber.go`): Remote Wi-Fi change

### Flow

```
1. Backup existing Wi-Fi config (if exists)
2. Apply new Wi-Fi config to /etc/NetworkManager/system-connections/
3. Run nmcli connection reload
4. Wait 15 seconds
5. Check connection: nmcli dev wifi | grep -i <ssid>
   - If connected → Success, delete backup
   - If not connected → Restore backup, reboot
```

## Implementation Notes

- Use `nmcli device wifi list` to verify connection
- Path for Wi-Fi configs: `/etc/NetworkManager/system-connections/*.nmconnection`
- Backup path: `/tmp/wifi-backup.nmconnection`
- On restore: copy backup back, `nmcli connection reload`, then `sudo reboot`