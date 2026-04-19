#include <BLEDevice.h>
#include <BLEUtils.h>
#include <BLEServer.h>
#include <BLE2902.h>
#include <Preferences.h>
#include <ArduinoJson.h>

// --- Pin Definitions ---
#define RADAR_RX 16 
#define RADAR_TX 17 
#define MOSFET_PIN 4   // Controls power to the amplifier
#define AUDIO_PIN 5    // Sends the sound wave to the amplifier

// --- Deterrence Settings ---
#define PEST_FREQ 20000 // 20kHz (Ultrasonic) - Change to 3000 for a loud audible beep

// --- BLE UUIDs ---
#define SERVICE_UUID        "4fafc201-1fb5-459e-8fcc-c5c9c331914b"
#define CHARACTERISTIC_UUID "beb5483e-36e1-4688-b7f5-ea07361b26a8"

// --- LD2420 Protocol Constants ---
uint8_t openCfg[]  = {0xFD, 0xFC, 0xFB, 0xFA, 0x04, 0x00, 0xFF, 0x00, 0x01, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t setRunMode[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x08, 0x00, 0x12, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t closeCfg[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x02, 0x00, 0xFE, 0x00, 0x04, 0x03, 0x02, 0x01};

// --- Global Objects ---
Preferences preferences;
String nodeId = "";
String nodeSecret = "";
String hubFilter = "";
bool isMotionDetected = false;
int lastDistanceCm = -1;
bool isSetupMode = false;

// --- BLE Setup Callbacks ---
class MyCallbacks: public BLECharacteristicCallbacks {
    void onWrite(BLECharacteristic *pCharacteristic) {
      String value = pCharacteristic->getValue();
      if (value.length() > 0) {
        Serial.println("Received Provisioning Data...");
        JsonDocument doc;
        DeserializationError error = deserializeJson(doc, value);

        if (error) {
          Serial.print(F("deserializeJson() failed: "));
          Serial.println(error.f_str());
          return;
        }

        const char* nId = doc["node_id"];
        const char* nSecret = doc["node_secret"];
        const char* hFilter = doc["hub_filter"];

        if (nId && nSecret && hFilter) {
          preferences.begin("scarrow", false);
          preferences.putString("node_id", nId);
          preferences.putString("node_secret", nSecret);
          preferences.putString("hub_filter", hFilter);
          preferences.end();
          
          Serial.println("Config Saved with Secret! Restarting...");
          delay(1000);
          ESP.restart();
        } else {
          Serial.println("Error: Missing required fields (node_id, node_secret, or hub_filter)");
        }
      }
    }
};

void enterSetupMode() {
  isSetupMode = true;
  Serial.println("Entering SETUP MODE...");
  
  BLEDevice::init("Scarrow_Node_Setup");
  BLEDevice::setMTU(512); // Ensure we can receive the full JSON payload
  BLEServer *pServer = BLEDevice::createServer();
  BLEService *pService = pServer->createService(SERVICE_UUID);

  BLECharacteristic *pCharacteristic = pService->createCharacteristic(
                                         CHARACTERISTIC_UUID,
                                         BLECharacteristic::PROPERTY_READ |
                                         BLECharacteristic::PROPERTY_WRITE
                                       );

  pCharacteristic->setCallbacks(new MyCallbacks());
  pService->start();

  BLEAdvertising *pAdvertising = BLEDevice::getAdvertising();
  pAdvertising->addServiceUUID(SERVICE_UUID);
  pAdvertising->setScanResponse(true);
  pAdvertising->setMinPreferred(0x06);  
  pAdvertising->setMinPreferred(0x12);
  BLEDevice::startAdvertising();
  
  Serial.println("Advertising 'Scarrow_Node_Setup'...");
}

void clearRX() {
  while(Serial1.available()) {
    Serial1.read();
  }
}

bool sendCommand(const char* label, uint8_t* cmd, int len) {
  clearRX(); 
  Serial.print("[CMD] "); Serial.print(label); Serial.print("... ");
  Serial1.write(cmd, len);
  Serial1.flush();
  
  unsigned long start = millis();
  uint8_t resp[64];
  int idx = 0;
  while (millis() - start < 300 && idx < 64) {
    if (Serial1.available()) {
      resp[idx++] = Serial1.read();
    }
  }

  if (idx > 0) {
    Serial.print("ACK: ");
    for(int i=0; i<idx; i++) {
      if(resp[i] < 0x10) Serial.print("0");
      Serial.print(resp[i], HEX); Serial.print(" ");
    }
    Serial.println();
    return true; 
  } else {
    Serial.println("TIMEOUT!");
    return false;
  }
}

void setup() {
  Serial.begin(115200);
  
  // Check NVS for node_id and node_secret
  preferences.begin("scarrow", true);
  nodeId = preferences.getString("node_id", "");
  nodeSecret = preferences.getString("node_secret", "");
  hubFilter = preferences.getString("hub_filter", "");
  preferences.end();

  if (nodeId == "" || nodeSecret == "") {
    enterSetupMode();
    return; // Don't proceed to sensor initialization
  }

  Serial.println("Node ID: " + nodeId);
  Serial.println("Node Secret: [STORED]");
  Serial.println("Hub Filter: " + hubFilter);

  // Initialize Output Pins
  pinMode(MOSFET_PIN, OUTPUT);
  pinMode(AUDIO_PIN, OUTPUT);
  digitalWrite(MOSFET_PIN, LOW);

  Serial1.begin(115200, SERIAL_8N1, RADAR_RX, RADAR_TX);
  
  delay(2000); 
  Serial.println("\n\n=========================================");
  Serial.println("🚀 Scarrow Node - FIELD MODE");
  Serial.println("=========================================");
  
  sendCommand("Open Config", openCfg, sizeof(openCfg));
  delay(100);
  sendCommand("Set Normal Mode", setRunMode, sizeof(setRunMode));
  delay(100);
  sendCommand("Close Config", closeCfg, sizeof(closeCfg));
  
  Serial.println("\n--- Radar Active ---");
}

void loop() {
  if (isSetupMode) {
    delay(1000); // Just wait for provisioning
    return;
  }

  if (Serial1.available()) {
    String data = Serial1.readStringUntil('\n');
    data.trim();
    
    if (data.length() > 0) {
      if (data == "ON") {
        if (!isMotionDetected) {
           Serial.println("🚨 MOTION!");
           isMotionDetected = true;
           digitalWrite(MOSFET_PIN, HIGH);
           delay(50);
           tone(AUDIO_PIN, PEST_FREQ);
        }
      } 
      else if (data == "OFF") {
        if (isMotionDetected) {
           Serial.println("✅ CLEAR");
           isMotionDetected = false;
           noTone(AUDIO_PIN);
           digitalWrite(MOSFET_PIN, LOW);
        }
      } 
      else if (data.startsWith("Range")) {
        int spaceIdx = data.indexOf(' ');
        if (spaceIdx > 0) {
          int distance_cm = data.substring(spaceIdx + 1).toInt();
          Serial.print("🎯 Distance: "); Serial.print(distance_cm); Serial.println(" cm");
        }
      }
    }
  }
}
